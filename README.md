# AWS Nitro Enclave Foobar Service

> ⚠️ At the time of writing, the code in this repo requires running paid Amazon
> Web Services (AWS) resources. You will be billed for the various resources
> you'll use. Make sure to delete and turn off resources you no longer need --
> the code in this repo does not handle that aspect. A budget of $1.00/day
> should be sufficient, but you exact expense will vary.

## Overview

This repo contains an example AWS Nitro Enclave service. The service has two
components. The first component, `foobar-enclave`, is a toy service which runs
inside the enclave. The second component, `foobar-instance`, is a command line
tool to interact with the enclave. Both components are written in Go.
`foobar-enclave` is designed to be fully verifiable, including the AWS KMS key
creation aspect.

AWS Nitro Enclaves provide a mechanism for generating attestations. Attestations
are documents signed with a key which is outside the control of AWS customers.
The attestation proves that a given computation took place inside an AWS Nitro
Enclave running a specific software.

At a very high level, an attestation contains the SHA-384 hash of the image.
Additional measurements (such as a hash of the AWS account id), and custom
fields. The code running in the enclave can use the custom fields to return data
which cannot be spoofed, e.g. a public key or a computation result.

## Operations

The code performs the following operations:

1. The enclave connects to AWS KMS and creates an Ecdsa key (curve P-256). The
key policy is locked down and the public half of the key is returned in an
attestation.
2. The command line tool verifies the attestation and encrypts a string, e.g.
"attack at dawn".
3. The enclave decrypts the string and returns the number of occurances of the
letter 'a' in the original plaintext, e.g. 4 for attack at dawn". The response
is also returned in an attestation, enabling anyone to verify the result without
revealing the plaintext.

### Key creation
In some cases, the enclave cannot blindly trust AWS KMS keys. In order to be
fully attestable, the enclave must talk directly to KMS and generate an
attestation of the key creation process.

Enclaves can only communicate to their parent instance using a vsock. The
command line tool which is running on the parent instance sets up an AWS KMS <=>
vsock proxy. The instance shares its IAM credentials with the enclave.

The following key policy is used to lock down the key. Root cannot use the
key or alter the policy -- they can only delete the key.
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::123456789012:role/nitro-test-iam-role"
            },
            "Action": "kms:GetPublicKey",
            "Resource": "*",
            "Condition": {}
        },
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::123456789012:role/nitro-test-iam-role"
            },
            "Action": "kms:DeriveSharedSecret",
            "Resource": "*",
            "Condition": {
                "StringEqualsIgnoreCase": {
                    "kms:RecipientAttestation:PCR0": "d026d9458187d50f45f2569e98f8140d1be5c6c3df8117b39794f72ece07d7d9ffc579bd451409bf56386837d66b3e72"
                }
            }
        },
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::123456789012:root"
            },
            "Action": [
                "kms:DescribeKey",
                "kms:GetKeyPolicy",
                "kms:ScheduleKeyDeletion"
            ],
            "Resource": "*",
            "Condition": {}
        }
    ]
}
```

### Encryption
The command line tool can encrypt strings without needing to communicate with
KMS or the enclave. The process to encrypt a string is:
- extract the KMS-backed Ecdsa's public key from the attestation.
- generate an ephemeral Ecdsa keypair.
- use ECDH with the KMS-backed Ecdsa public key and the ephemeral Ecdsa private
  key to derive a content encryption key (CEK).
- use AES-GCM to encrypt the plaintext with the CEK.
- store the ephemeral Ecdsa's public key, nonce, and ciphertext as the encrypted
  message.

### Decryption
Decryption works as following:
- the enclave creates an ephemeral RSA key at startup. This key is never
  persisted.
- the command line tool requests a fresh attestation from the enclave (KMS
  requires an attestation no old than 5 minutes). The attestation contains the
  ephemeral RSA public key.
- the command line tool requests KMS to perform an ECDH operation. The
  fresh attestation is used to authenticate the request and encrypt the
  response.
- the command line tool sends the encrypted response, nonce, and ciphertext
  to the enclave. It is important to keep in mind that the encrypted response
  uses RSA with AES-CBC and is not authenticated! The setup in this example
  is secure, because the unauthenticated encryption backs a CEK which is
  then used with an authenticated encryption (AES-GCM). It is possible
  (and probably best) to use the KMS <=> vsock proxy for all KMS operations.
  Better TLS ciphers are then protecting the data and AES-CBC becomes a
  non-issue. Users cannot pick alternate ciphers and are forced to use RSA with
  AES-CBC.
- the enclave uses the ephemeral RSA key to decrypt the CEK. The enclave
  decrypts the ciphertext.
- the enclave returns a count of the letter 'a' in the plaintext inside an
  attestation. The attestation also contains a hash of the inputs (encrypted
  cek, nonce, and ciphertext).

## AWS setup
TODO: describe how to setup AWS, including IAM role.

## Building and running foobar-service
Use the following command from your EC2 instance to build & run the enclave:
```bash
./build-and-run-enclave.sh
```

You can then build and run the CLI:
```bash
cd foobar-instance
go build .

# create-key requires root to bind to vsock
sudo ./foobar-instance create-key

# verify the attestation and encrypt a message
CIPHERTEXT=`./foobar-instance encrypt --plaintext="attack at dawn"`

# ask enclave to decrypt ciphertext and return count of 'a'
./foobar-instance decrypt --ciphertext $CIPHERTEXT
```

Don't forget to turn off any resources you no longer need.

## Thinking about trust
In order to trust this toy service, you would need to:
- audit the enclave's source code and its recursive dependencies. There's 
  about 300k lines of code just for this simple toy enclave (this number
  goes down once you eliminate unreachable code). Lines of code isn't a
  meaningful measure of anything, except to quantify the amount of code we
  are dealing with.
- audit or trust the Go standard library and compiler.
- build the code and obtain a bit-for-bit identical binary.
- independently download and confirm aws-ca-certificates.crt
  (Amazon's root CA certificates) is correct.
- independently download and confirm root.pem (Amazon Nitro Enclave PKI root
  key) is correct.
- trust Amazon Nitro's security claims and the nitro-cli tooling.
- trust Amazon KMS to properly protect its keys.

The following cryptographic primitives need to be trusted:
- SHA-384, the hash function used to compute PCR measurements.
- TLS, used to download and check aws-ca-certificates.crt and root.pem. Also
  used for communication between the enclave and AWS KMS.
  - RSA and SHA-256 used during the TLS handshake to verify certificates.
  - Algorithms negotiated during the TLS handshake. Currently, the server favors
    ECDH, AES-GCM, and SHA-384. The preferred ciphers can vary without notice.
  - Algorithms used by CRL and OCSP.
- AES-GCM, used to encrypt the plaintext.
- ECDH, used to derive the CEK.
- RSA + AES-CBC, used to encrypt the CEK.
- ECDSA signatures and SHA-384 used to verify attestations.

You also has to trust your own hardware and software stack
(operating system, browser, etc.). You have to trust that your clock is somewhat
accurate to ensure you aren't been served expired certificates.

To summarize, you need to audit or blindly trust what amounts to
probably 35M+ lines of code! Hundreds of people have reviewed bits and
pieces of this overall software but bugs leading to security failures are
very likely still lurking in there. Will we one day have large amounts of
provably correct code?

## Limitations

The code in this repo is meant to be an example only. The current design
does not permit upgrades to the code while keeping the same AWS KMS key --
any code changes to the enclave will result in a different PCR0 hash. The KMS
key policy is tied to a specific PCR0 value.

The current implementation isn't developer friendly. Developer ergonomics can
be improved by mocking AWS infrastructure or using a cloud emulator.

# About Zxs

[Zxs](https://zxs.ch/) has audited several AWS Nitro Enclave-based solutions
as well as various other secure compute environments. This code served an
important starting point to deepen understanding.

# Related Links
- Trail Of Bits
  - [A few notes on AWS Nitro Enclaves: Images and attestation](https://blog.trailofbits.com/2024/02/16/a-few-notes-on-aws-nitro-enclaves-images-and-attestation/)
  - [A few notes on AWS Nitro Enclaves: Attack surface](https://blog.trailofbits.com/2024/09/24/notes-on-aws-nitro-enclaves-attack-surface/)
- [The Security Design of the AWS Nitro System](https://docs.aws.amazon.com/pdfs/whitepapers/latest/security-design-of-aws-nitro-system/security-design-of-aws-nitro-system.pdf#security-design-of-aws-nitro-system)
- [AWS Nitro System gets independent affirmation of its confidential compute capabilities](https://aws.amazon.com/blogs/compute/aws-nitro-system-gets-independent-affirmation-of-its-confidential-compute-capabilities/)
- [nitriding by Brave](https://github.com/brave/nitriding-daemon)
- [What is Nitro Enclaves?](https://docs.aws.amazon.com/enclaves/latest/user/nitro-enclave.html)
- [AWS Key Management Service](https://docs.aws.amazon.com/kms/latest/developerguide/overview.html)
- [How AWS Nitro Enclaves uses AWS KMS](https://docs.aws.amazon.com/kms/latest/developerguide/services-nitro-enclaves.html)
- AWS' GitHub repos
  - [aws-nitro-enclaves-image-format](https://github.com/aws/aws-nitro-enclaves-image-format)
  - [aws-nitro-enclaves-sdk-c](https://github.com/aws/aws-nitro-enclaves-sdk-c)
  - etc.
