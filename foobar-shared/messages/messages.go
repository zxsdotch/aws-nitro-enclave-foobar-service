package messages

type FoobarRequest struct {
	CreateKey      *CreateKeyRequest      `json:"createKey,omitempty"`
	GetAttestation *GetAttestationRequest `json:"getAttestation,omitempty"`
	Decrypt        *DecryptRequest        `json:"decrypt,omitempty"`
}

type FoobarResponse struct {
	CreateKey      *CreateKeyResponse      `json:"createKey,omitempty"`
	GetAttestation *GetAttestationResponse `json:"getAttestation,omitempty"`
	Decrypt        *DecryptResponse        `json:"decrypt,omitempty"`
	Error          *string                 `json:"error,omitempty"`
}
