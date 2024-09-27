package messages

// Requests decryption. EncryptedCek comes from KMS and is formatted as CMS.
// RSA is used to encrypt an AES key, which then encrypts the CEK with AES-CMS.
type DecryptRequest struct {
	EncryptedCek []byte `json:"cek"`
	Nonce        []byte `json:"nonce"`
	Ciphertext   []byte `json:"ciphertext"`
}

// Response is an attestation which contains DecryptResponseAttestationUserData.
type DecryptResponse struct {
	Attestation []byte `json:"attestation"`
}

// InitialRequest is a SHA-256 of the DecryptRequest and is used to tie the
// request with the result.
type DecryptResponseAttestationUserData struct {
	InitialRequest []byte `json:"request"`
	Count          int    `json:"count"`
}
