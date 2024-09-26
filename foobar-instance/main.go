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
	createKeyAttestationPath = createKeyCmd.Flag("attestationPath", "Path to save attestation").Default("./key.out").String()
)

func main() {
	ctx := context.TODO()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case createKeyCmd.FullCommand():
		cmds.CreateKey(ctx, *createKeyCmdRole, *createKeyAttestationPath)
	default:
		panic("invalid command")
	}
}
