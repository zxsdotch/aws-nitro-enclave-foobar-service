package messages

type FoobarRequest struct {
	CreateKey *CreateKeyRequest `json:"createKey,omitempty"`
}

type FoobarResponse struct {
	CreateKey *CreateKeyResponse `json:"createKey,omitempty"`
	Error     *string            `json:"error,omitempty"`
}
