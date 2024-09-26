package messages

type DecryptRequest struct {
	EncryptedCek []byte `json:"cek"`
	Nonce        []byte `json:"nonce"`
	Ciphertext   []byte `json:"ciphertext"`
}

type DecryptResponse struct {
	Attestation []byte `json:"attestation"`
}

type AttestationUserData2 struct {
	RequestSha256 []byte `json:"request"`
	Count         int    `json:"count"`
}
