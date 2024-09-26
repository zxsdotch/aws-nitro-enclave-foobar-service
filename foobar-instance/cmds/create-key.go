package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/mdlayher/vsock"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/constants"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/utils"
)

// In order to securely create an AWS KMS key, we need to connect to KMS from
// within the enclave, which requires proxying TLS over vsock. If we don't
// do this complicated proxying, the enclave doesn't know if the key its
// using is actually backed by KMS or not!

func CreateKey(ctx context.Context, awsIamRole, attestationPath string) {
	// Step 1:
	//   Grab various pieces of information from the Instance Metadata Service
	//   (imds). This includes our region, account id, IAM credentials, etc.
	cfg, err := config.LoadDefaultConfig(ctx)
	utils.PanicOnErr(err)

	client := imds.NewFromConfig(cfg)
	region, err := client.GetRegion(ctx, &imds.GetRegionInput{})
	utils.PanicOnErr(err)
	log.Printf("region: %s\n", region.Region)

	iamInfo, err := client.GetIAMInfo(ctx, &imds.GetIAMInfoInput{})
	utils.PanicOnErr(err)
	arn, err := arn.Parse(iamInfo.InstanceProfileArn)
	utils.PanicOnErr(err)
	log.Printf("account id: %s\n", arn.AccountID)

	path := fmt.Sprintf("iam/security-credentials/%s", awsIamRole)
	metadataResp, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: path})
	utils.PanicOnErr(err)
	rawCredentials, err := io.ReadAll(metadataResp.Content)
	utils.PanicOnErr(err)
	var credentials messages.Credentials
	err = json.Unmarshal(rawCredentials, &credentials)
	utils.PanicOnErr(err)
	// The credentials we get from the metadata server should look like:
	// {
	// 	"LastUpdated" : "2024-09-06T16:31:19Z",
	// 	"Type" : "AWS-HMAC",
	// 	"AccessKeyId" : "AKIAIOSFODNN7EXAMPLE",
	// 	"SecretAccessKey" : "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	// 	"Token" : "KtzpFWDJ...",
	// 	"Expiration" : "2024-09-06T23:07:03Z"
	// }
	log.Printf("IAM credentials:\n%v\n", credentials)

	// Step 2:
	//   Setup a proxy that fuses the "left" and "right" sockets:
	//   AWS KMS <==> instance <==> enclave. Given that this is a CLI, we don't
	//   need to think about tearing down the proxy -- it happens automatically
	//   when the CLI exits.
	_ = NewProxy(fmt.Sprintf("kms.%s.amazonaws.com:443", region.Region))

	// Step 3:
	//   Tell enclave to create the key
	resp := sendRequest(messages.FoobarRequest{
		CreateKey: &messages.CreateKeyRequest{
			Region:      region.Region,
			AccountId:   arn.AccountID,
			AwsIamRole:  awsIamRole,
			Credentials: credentials,
		},
	})

	// Step 4:
	//   Save the attestation for the next operation.
	os.WriteFile(attestationPath, []byte(resp.CreateKey.Attestation), 0644)
}

// Quick and hacky proxy. Previous familiarity with
// https://github.com/ghostunnel/ghostunnel helped!

type Proxy struct {
	leftHost string
}

func NewProxy(leftHost string) *Proxy {
	proxy := &Proxy{
		leftHost: leftHost,
	}
	go func() { proxy.start() }()
	return proxy
}

func (proxy Proxy) start() {
	log.Printf("starting proxy, listening on vsock (cid=%d, port=%d)\n", constants.INSTANCE_CID, constants.INSTANCE_LISTENING_PORT)
	listener, err := vsock.ListenContextID(constants.INSTANCE_CID, constants.INSTANCE_LISTENING_PORT, nil)
	utils.PanicOnErr(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener.Accept() failed: ", err)
			continue
		}

		// handle connection in a goroutine
		go proxy.handleConnection(conn)
	}
}

func (proxy Proxy) handleConnection(right net.Conn) {
	left, err := net.Dial("tcp", proxy.leftHost)
	if err != nil {
		log.Println("Dial() failed: ", err)
		right.Close()
		return
	}

	log.Printf("new connection: %#v\n", right)
	log.Printf("new connection: %+v\n", right)
	log.Printf("new connection: %v\n", right)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() { proxy.copy(left, right); wg.Done() }()
	go func() { proxy.copy(right, left); wg.Done() }()
	wg.Wait()
	log.Println("connection closed")
}

func (proxy Proxy) copy(dst net.Conn, src net.Conn) {
	defer src.Close()
	defer dst.Close()
	_, err := io.Copy(dst, src)
	if err != nil && !isClosedConnectionError(err) {
		log.Println("io.Copy error: ", err)
	}
}

func isClosedConnectionError(err error) bool {
	if e, ok := err.(*net.OpError); ok {
		return e.Op == "read" && strings.Contains(err.Error(), "closed network connection")
	}
	return false
}
