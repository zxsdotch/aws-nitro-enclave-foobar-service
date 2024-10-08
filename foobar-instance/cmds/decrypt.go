package cmds

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	nitro_eclave_attestation_document "github.com/alokmenghrajani/go-nitro-enclave-attestation-document"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/utils"
)

// Decryption works as followingL
// 1. tell enclave to create an attestation with an ephemeral RSA key
// 2. use the attestation with KMS to derive an encrypted CEK.
// 3. give the ciphertext and CEK to the enclave.
// 4. receive a response inside an attestation.
// 5. decode the attestation and print the result.

func Decrypt(ctx context.Context, attestationPath, rootPath, ciphertext string) {
	// Step 1: Use the attestation from createKey to get the key id
	attestationBytes, err := os.ReadFile(attestationPath)
	utils.PanicOnErr(err)

	root, err := os.ReadFile(rootPath)
	utils.PanicOnErr(err)

	rootPublicKeyBlock, _ := pem.Decode(root)
	rootPublicKey, err := x509.ParseCertificate(rootPublicKeyBlock.Bytes)
	utils.PanicOnErr(err)

	attestation, err := nitro_eclave_attestation_document.AuthenticateDocument(attestationBytes, *rootPublicKey, true)
	utils.PanicOnErr(err)

	var userData messages.CreateKeyResponseAttestationUserData
	err = json.Unmarshal(attestation.UserData, &userData)
	utils.PanicOnErr(err)

	log.Printf("key id: %s", userData.KeyId)

	// Step 2: grab the ephemeral ecdsa public key from the ciphertext message
	ciphertextMessageBytes, err := base64.RawURLEncoding.DecodeString(ciphertext)
	utils.PanicOnErr(err)
	var ciphertextMessage ciphertextMessage
	err = json.Unmarshal(ciphertextMessageBytes, &ciphertextMessage)
	utils.PanicOnErr(err)

	// Step 3: request a fresh attestation from the enclave. We don't need to
	// valdidate it, KMS takes care of that.
	resp, _ := sendRequest(messages.FoobarRequest{GetAttestation: &messages.GetAttestationRequest{}})
	freshAttestation := resp.GetAttestation.Attestation

	// Step 4: get an encrypted-shared secret from KMS
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(userData.Region))
	utils.PanicOnErr(err)

	kmsClient := kms.NewFromConfig(cfg)
	deriveSharedSecretOutput, err := kmsClient.DeriveSharedSecret(ctx, &kms.DeriveSharedSecretInput{
		KeyAgreementAlgorithm: types.KeyAgreementAlgorithmSpecEcdh,
		KeyId:                 &userData.KeyId,
		PublicKey:             ciphertextMessage.EphemeralKey,
		Recipient: &types.RecipientInfo{
			AttestationDocument:    freshAttestation,
			KeyEncryptionAlgorithm: types.KeyEncryptionMechanismRsaesOaepSha256, // encryption algorithm for the second ciphertext
		},
	})
	utils.PanicOnErr(err)

	log.Printf("Encrypted shared secret: %s", base64.RawURLEncoding.EncodeToString(deriveSharedSecretOutput.CiphertextForRecipient))

	// Step 5: send the encrypted shared secret to the enclave
	resp2, msgBytes := sendRequest(messages.FoobarRequest{Decrypt: &messages.DecryptRequest{
		EncryptedSharedSecret: deriveSharedSecretOutput.CiphertextForRecipient,
		Nonce:                 ciphertextMessage.Nonce,
		Ciphertext:            ciphertextMessage.Ciphertext,
	}})

	// Step 6: verify attestation is valid and extract response.
	responseAttestation, err := nitro_eclave_attestation_document.AuthenticateDocument(resp2.Decrypt.Attestation, *rootPublicKey, true)
	utils.PanicOnErr(err)
	log.Printf("attestation valid")
	log.Printf("PCR0: %02x", responseAttestation.PCRs[0])

	var response messages.DecryptResponseAttestationUserData
	err = json.Unmarshal(responseAttestation.UserData, &response)
	utils.PanicOnErr(err)

	log.Printf("Request SHA-256: %02x", response.InitialRequest)

	// Calculate expected sha
	h := sha256.New()
	h.Write(msgBytes)
	log.Printf("expected:        %02x", h.Sum(nil))

	fmt.Printf("Count 'a': %d\n", response.Count)
}
