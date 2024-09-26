package handlers

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"

	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
)

func GetAttestationHandler(ctx context.Context, ephemeralRsaKey *rsa.PrivateKey, req messages.GetAttestationRequest) (*messages.GetAttestationResponse, error) {
	r := &messages.GetAttestationResponse{}

	sess, err := nsm.OpenDefaultSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	ephemeralRsaPublicKey, err := x509.MarshalPKIXPublicKey(&ephemeralRsaKey.PublicKey)
	if err != nil {
		return nil, err
	}

	res, err := sess.Send(&request.Attestation{
		Nonce:     []byte{},
		UserData:  []byte{},
		PublicKey: ephemeralRsaPublicKey,
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
