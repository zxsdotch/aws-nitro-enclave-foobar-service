package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
	"github.com/mdlayher/vsock"

	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/constants"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/utils"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
)

func CreateKeyHandler(ctx context.Context, req messages.CreateKeyRequest) (*messages.CreateKeyResponse, error) {
	r := &messages.CreateKeyResponse{}

	// The AWS SDK must talk to the vsock. Thankfully, the AWS SDK allows setting
	// custom http clients.
	httpClient := awshttp.NewBuildableClient().WithTransportOptions(func(tr *http.Transport) {
		tr.DialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
			log.Printf("Connecting to vsock (cid=%d, port=%d)\n", constants.INSTANCE_CID, constants.INSTANCE_LISTENING_PORT)
			return vsock.Dial(constants.INSTANCE_CID, constants.INSTANCE_LISTENING_PORT, nil)
		}
	})

	// The AWS SDK must use the credentials from the parent intance.
	credentialProvider := credentials.NewStaticCredentialsProvider(req.Credentials.AccessKeyId, req.Credentials.SecretAccessKey, req.Credentials.Token)

	// We can now talk to KMS.
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(req.Region),
		config.WithHTTPClient(httpClient),
		config.WithCredentialsProvider(credentialProvider))
	if err != nil {
		return nil, err
	}
	kmsClient := kms.NewFromConfig(cfg)

	// Grab the enclave's PCR0. We need it for the key policy. In a real world
	// application, we would grab other PCR measurements.
	sess, err := nsm.OpenDefaultSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	res, err := sess.Send(&request.DescribePCR{Index: 0})
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, fmt.Errorf("request.DescribePCR error: %s", res.Error)
	}
	pcr0 := fmt.Sprintf("%02x", res.DescribePCR.Data)

	// Build a restrictive key policy.
	accountId := req.AccountId
	awsRole := req.AwsIamRole
	enclavePrincipal := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountId, awsRole)
	rootPrincipal := fmt.Sprintf("arn:aws:iam::%s:root", accountId)

	policy := Policy{
		Version: "2012-10-17",
		Statements: []Statement{
			// First statement: allow the enclave to get the public key.
			// Note: we'll be setting BypassPolicyLockoutSafetyCheck since the policy
			// is locked down.
			{
				Effect: "Allow",
				Principal: map[string]string{
					"AWS": enclavePrincipal,
				},
				Action:   []string{"kms:GetPublicKey"},
				Resource: "*",
			},
			// Second statement: only enclave can perform ECDH with this key.
			{
				Effect: "Allow",
				Principal: map[string]string{
					"AWS": enclavePrincipal,
				},
				Action:   []string{"kms:DeriveSharedSecret"},
				Resource: "*",
				Condition: Condition{
					StringEqualsIgnoreCase: map[string]string{"kms:RecipientAttestation:PCR0": pcr0},
				},
			},
			// Third statement: allow root to delete the key. Keep in mind that IAM
			// roles only work for permissions granted to root.
			{
				Effect:    "Allow",
				Principal: map[string]string{"AWS": rootPrincipal},
				Action:    []string{"kms:DescribeKey", "kms:GetKeyPolicy", "kms:ScheduleKeyDeletion"},
				Resource:  "*",
			},
		},
	}
	policyString, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}

	// Create the key
	createKeyResult, err := kmsClient.CreateKey(ctx, &kms.CreateKeyInput{
		Description:                    aws.String("github.com/zxsdotch/aws-nitro-enclave-foobar-service"),
		KeySpec:                        types.KeySpecEccNistP256,
		KeyUsage:                       types.KeyUsageTypeKeyAgreement,
		Policy:                         utils.Ref(string(policyString)),
		BypassPolicyLockoutSafetyCheck: true,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("key id: %s\n", *createKeyResult.KeyMetadata.KeyId)

	// Grab the public key
	getPublicKeyResult, err := kmsClient.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: createKeyResult.KeyMetadata.KeyId,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("public key: %s\n", base64.RawURLEncoding.EncodeToString(getPublicKeyResult.PublicKey))

	// Return the public key in an attestation
	userData := messages.AttestationUserData{
		KeyId:     *createKeyResult.KeyMetadata.KeyId,
		PublicKey: getPublicKeyResult.PublicKey,
		Region:    req.Region,
	}
	userDataBytes, err := json.Marshal(userData)
	if err != nil {
		return nil, err
	}

	res, err = sess.Send(&request.Attestation{
		Nonce:     []byte{},
		UserData:  userDataBytes,
		PublicKey: []byte{},
	})
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, fmt.Errorf("request.Attestation error: %s", res.Error)
	}
	if res.Attestation == nil || res.Attestation.Document == nil {
		return nil, errors.New("NSM did not return an attestation")
	}

	r.Attestation = res.Attestation.Document

	return r, nil
}

// These structs probably exist somewhere in the SDK. I didn't find them.
type Policy struct {
	Version    string      `json:"Version"`
	Statements []Statement `json:"Statement"`
}

type Statement struct {
	Effect    string            `json:"Effect"`
	Principal map[string]string `json:"Principal"`
	Action    []string          `json:"Action"`
	Resource  string            `json:"Resource"`
	Condition Condition         `json:"Condition,omitempty"`
}

type Condition struct {
	StringEqualsIgnoreCase map[string]string `json:"StringEqualsIgnoreCase,omitempty"`
}
