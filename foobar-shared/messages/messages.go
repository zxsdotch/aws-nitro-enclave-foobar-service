package messages

// These messages define the API between the foobar-instance and foobar-enclave.
// Only one of each field is expected to be set at any given time.
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
