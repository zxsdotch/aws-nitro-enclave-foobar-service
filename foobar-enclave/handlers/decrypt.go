package handlers

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/edgebitio/nitro-enclaves-sdk-go/crypto/cms"
	"github.com/hf/nsm"
	"github.com/hf/nsm/request"

	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
)

func DecryptHandler(ctx context.Context, ephemeralRsaKey *rsa.PrivateKey, req messages.DecryptRequest, reqBytes []byte) (*messages.DecryptResponse, error) {
	r := &messages.DecryptResponse{}

	// Decrypt CEK
	cmsMessage, err := cms.Parse(req.EncryptedCek)
	if err != nil {
		return nil, err
	}

	cek, err := cmsMessage.Decrypt(ephemeralRsaKey)
	if err != nil {
		return nil, err
	}

	// AES-GCM decryption
	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, req.Nonce, req.Ciphertext, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("plaintext: %02x", plaintext)

	// Compute result

	count := 0
	for i := 0; i < len(plaintext); i++ {
		if plaintext[i] == 'a' {
			count += 1
		}
	}

	// Hash the inputs to defend against input swapping
	h := sha256.New()
	h.Write(reqBytes)

	userData := messages.DecryptResponseAttestationUserData{
		InitialRequest: h.Sum(nil),
		Count:          count,
	}
	userDataBytes, err := json.Marshal(userData)
	if err != nil {
		return nil, err
	}

	sess, err := nsm.OpenDefaultSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	res, err := sess.Send(&request.Attestation{
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
