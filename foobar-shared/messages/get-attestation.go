package messages

type GetAttestationRequest struct {
}

type GetAttestationResponse struct {
	Attestation []byte `json:"attestation"`
}
