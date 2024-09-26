package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/mdlayher/vsock"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-enclave/handlers"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/constants"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/utils"
)

var (
	ephemeralRsaKey *rsa.PrivateKey
)

func main() {
	log.Println("foobar-enclave is starting")

	// Generate an ephemeral RSA key, which will be used by AWS KMS to encrypt
	// responses. The choice of RSA is currently not configurable. In addition,
	// this is the only way to create policies which bind to specific PCR0 hashes.
	var err error
	ephemeralRsaKey, err = rsa.GenerateKey(rand.Reader, 2048)
	utils.PanicOnErr(err)

	log.Printf("rsaKey: %s\n", base64.RawURLEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&ephemeralRsaKey.PublicKey)))

	fmt.Printf("listening on vsock (port=%d)\n", constants.ENCLAVE_LISTENING_PORT)
	listener, err := vsock.Listen(constants.ENCLAVE_LISTENING_PORT, nil)
	utils.PanicOnErr(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept() failed: ", err)
			continue
		}

		// handle connection in a goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		ctx := context.TODO()

		log.Printf("recv: %+v", scanner.Text())
		var req messages.FoobarRequest
		var res messages.FoobarResponse

		var err error
		err = json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			err = fmt.Errorf("json.Unmarshal failed: %w", err)
		} else {
			if req.CreateKey != nil {
				res.CreateKey, err = handlers.CreateKeyHandler(ctx, *req.CreateKey)
			} else if req.GetAttestation != nil {
				res.GetAttestation, err = handlers.GetAttestationHandler(ctx, ephemeralRsaKey, *req.GetAttestation)
			} else if req.Decrypt != nil {
				res.Decrypt, err = handlers.DecryptHandler(ctx, ephemeralRsaKey, *req.Decrypt)
			} else {
				err = fmt.Errorf("unexpected command")
			}
		}

		if err != nil {
			res.Error = utils.Ref(err.Error())
		}
		log.Printf("send: %+v", res)
		resBytes, err := json.Marshal(res)
		utils.PanicOnErr(err)
		conn.Write(resBytes)
		conn.Write([]byte{'\n'})
	}
	if err := scanner.Err(); err != nil {
		log.Printf("scanner.Scan() failed: %s\n", err)
	}
}
