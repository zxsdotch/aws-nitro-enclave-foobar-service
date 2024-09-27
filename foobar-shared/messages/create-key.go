package messages

// Requests key creation. The key is an asymmetric key, backed by KMS.
type CreateKeyRequest struct {
	Region      string      `json:"region"`
	AccountId   string      `json:"accountId"`
	AwsIamRole  string      `json:"awsIamRole"`
	Credentials Credentials `json:"credentials"`
}

// Response is an attestation which contains the keyid and related information.
type CreateKeyResponse struct {
	Attestation []byte `json:"attestation"`
}

type CreateKeyResponseAttestationUserData struct {
	KeyId     string `json:"keyId"`
	PublicKey []byte `json:"pubKey"`
	Region    string
}

// Credentials struct as returned by
// http://169.254.169.254/latest/meta-data/iam/security-credentials/<iam role>
//
// This struct should probably exist in the AWS SDK.
type Credentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Token           string `json:"Token"`
}
