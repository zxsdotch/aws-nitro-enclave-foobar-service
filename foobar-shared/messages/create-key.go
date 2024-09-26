package messages

type CreateKeyRequest struct {
	Region      string      `json:"region"`
	AccountId   string      `json:"accountId"`
	AwsIamRole  string      `json:"awsIamRole"`
	Credentials Credentials `json:"credentials"`
}

type CreateKeyResponse struct {
	Attestation string `json:"attestation"`
}

type AttestationUserData struct {
	KeyId     string `json:"keyId"`
	PublicKey []byte `json:"pubKey"`
}

type Credentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Token           string `json:"Token"`
}
