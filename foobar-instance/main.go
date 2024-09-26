package main

import (
	"context"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-instance/cmds"
)

var (
	app = kingpin.New("foorbar-instance", "The AWS EC2 instance part of Foobar Service.")

	createKeyCmd             = app.Command("create-key", "Tells enclave to create an AWS KMS key. Sets up a vsock<=>kms proxy.")
	createKeyCmdRole         = createKeyCmd.Flag("role", "AWS IAM Role").Default("nitro-test-iam-role").String()
	createKeyAttestationPath = createKeyCmd.Flag("attestationPath", "Path to save attestation").Default("./attestation.out").String()

	encryptCmd             = app.Command("encrypt", "Encrypts a string to the KMS-backed key.")
	encryptAttestationPath = encryptCmd.Flag("attestationPath", "Path to read attestation from, as returned by createKey command.").Default("./attestation.out").String()
	encryptRootPath        = encryptCmd.Flag("rootPath", "Path to Enclave PKI root CA file").Default("./root.pem").String()
	encryptPlaintext       = encryptCmd.Flag("plaintext", "Text to encrypt.").Required().String()
)

func main() {
	ctx := context.TODO()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case createKeyCmd.FullCommand():
		cmds.CreateKey(ctx, *createKeyCmdRole, *createKeyAttestationPath)
	case encryptCmd.FullCommand():
		cmds.Encrypt(*encryptAttestationPath, *encryptRootPath, *encryptPlaintext)
	default:
		panic("invalid command")
	}
}
