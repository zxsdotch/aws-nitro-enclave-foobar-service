package messages

// Requests a fresh attestation. KMS requires attestations no older than
// 5 minutes.
type GetAttestationRequest struct {
}

// Returns an attestation. The public_key contains an ephemeral RSA key.
type GetAttestationResponse struct {
	Attestation []byte `json:"attestation"`
}
