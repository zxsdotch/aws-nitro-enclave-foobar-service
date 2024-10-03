module github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-instance

go 1.21.4

require (
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/alokmenghrajani/go-nitro-enclave-attestation-document v1.0.1
	github.com/aws/aws-sdk-go-v2 v1.31.0
	github.com/aws/aws-sdk-go-v2/config v1.27.33
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.13
	github.com/aws/aws-sdk-go-v2/service/kms v1.36.2
	github.com/mdlayher/vsock v1.2.1
	github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared v0.0.0
)

replace github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared => ../foobar-shared

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.32 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.7 // indirect
	github.com/aws/smithy-go v1.21.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/veraison/go-cose v1.0.0-rc.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)
