package cmds

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	nitro_eclave_attestation_document "github.com/alokmenghrajani/go-nitro-enclave-attestation-document"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/utils"
)

// Encryption works as following:
// 1. verify attestation is valid and extract public ECC key.
// 2. generate an ephemeral ECC keypair.
// 3. derive a CEK using ECDH.
// 4. Use the CEK to encrypt the plaintext with AES-GCM.
// 5. Return the ciphertext as a json blob, containing the ephemeral ECC public
//    key.

func Encrypt(attestationPath, rootPath, plaintext string) {
	attestationBytes, err := os.ReadFile(attestationPath)
	utils.PanicOnErr(err)

	root, err := os.ReadFile(rootPath)
	utils.PanicOnErr(err)

	rootPublicKeyBlock, _ := pem.Decode(root)
	rootPublicKey, err := x509.ParseCertificate(rootPublicKeyBlock.Bytes)
	utils.PanicOnErr(err)

	// Step 1: verify attestation is valid and extract public Ecdsa key.
	// Normally, the user would specify here which PCR0 values to trust.
	attestation, err := nitro_eclave_attestation_document.AuthenticateDocument(attestationBytes, *rootPublicKey, true)
	utils.PanicOnErr(err)
	log.Printf("attestation valid")
	log.Printf("PCR0: %02x", attestation.PCRs[0])

	var userData messages.AttestationUserData
	err = json.Unmarshal(attestation.UserData, &userData)
	utils.PanicOnErr(err)

	pkixPublicKey, err := x509.ParsePKIXPublicKey(userData.PublicKey)
	utils.PanicOnErr(err)
	kmsPublicKey := pkixPublicKey.(*ecdsa.PublicKey)

	// Step 2: generate an ephemeral Ecdsa keypair.
	ephemeralEcdsaKey, err := ecdsa.GenerateKey(kmsPublicKey.Curve, rand.Reader)
	utils.PanicOnErr(err)

	// Step 3: derive a content encryption key (CEK)
	privateKey, err := ephemeralEcdsaKey.ECDH()
	utils.PanicOnErr(err)
	publicKey, err := kmsPublicKey.ECDH()
	utils.PanicOnErr(err)
	cek, err := privateKey.ECDH(publicKey)
	utils.PanicOnErr(err)

	// Step 4: AES-GCM encrypt plaintext with CEK
	block, err := aes.NewCipher(cek)
	utils.PanicOnErr(err)
	aesgcm, err := cipher.NewGCM(block)
	utils.PanicOnErr(err)
	nonce := make([]byte, aesgcm.NonceSize())
	_, err = rand.Read(nonce)
	utils.PanicOnErr(err)
	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Step 5: print the result
	ephemeralEcdsaKeyPublicKeyBytes, err := x509.MarshalPKIXPublicKey(&ephemeralEcdsaKey.PublicKey)
	utils.PanicOnErr(err)
	message := ciphertextMessage{
		EphemeralKey: ephemeralEcdsaKeyPublicKeyBytes,
		Nonce:        nonce,
		Ciphertext:   ciphertext,
	}
	messageBytes, err := json.Marshal(message)
	utils.PanicOnErr(err)

	messageString := base64.RawURLEncoding.EncodeToString(messageBytes)
	fmt.Println(messageString)
}

type ciphertextMessage struct {
	EphemeralKey []byte `json:"e"`
	Nonce        []byte `json:"n"`
	Ciphertext   []byte `json:"c"`
}
