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
tool to interact with the enclave. Both components are written in Go. Unlike
other examples I have seen, `foobar-enclave` is designed to be fully
verifiable, including the AWS KMS key creation aspect.

AWS Nitro Enclaves provide a mechanism for generating attestations. Attestations
are documents signed with a key which is outside the control of AWS customers.
The attestation proves that a given computation took place inside an AWS Ntro
Enclave running a specific software.

At a very high level, an attestation contains the code's SHA-384, additional
measurements (such as a hash of the AWS account id), and custom fields. The
code running in the enclave can use the custom fields to return data which
cannot be spoofed, e.g. a public key or a computation result.

## Operations

The code performs the following operations:

1. The enclave connects to AWS KMS and creates an ECC key (curve P-256). The key
policy is locked down and the public half of the key is returned in an
attestation.
2. The command line tool verifies the attestation and encrypts a string, e.g.
"attack at dawn".
3. The enclave decrypts the string and returns the number of occurances of the
letter 'a' in the original plaintext, e.g. 4 for attack at dawn". The response
is also returned in an attestation, enabling anyone to verify the result without
revealing the plaintext.

### Key creation
The enclave cannot blindly trust any AWS KMS key. In order to be fully
attestable, the enclave must talk directly to AWS KMS. Enclaves can however
only communicate to their parent instance using a vsock. The command line tool
therefore sets up an AWS KMS <=> vsock proxy. The instance shares its
IAM credentials with the enclave.

The following key policy is used to lock down the key:
TODO: describe key policy

### Encryption
TODO: explain CEK derived using an ephemeral key.

### Decryption
TODO: explain attestation + encrypted-CEK

# AWS setup
TODO: describe how to setup AWS, including IAM role.

# Building and running foobar-service
Use the following command from your EC2 instance to build & run the enclave:
```
./build-and-run-enclave.sh
```

You can then build and run the CLI:
```
cd foobar-instance
go build .

# create-key requires root to bind to vsock
sudo ./foobar-instance create-key

# verify the attestation and encrypt a message
CIPHERTEXT=`./foobar-instance encrypt --plaintext="attack at dawn"`

# ask enclave to decrypt ciphertext and return count of 'a'
./foobar-instance decrypt --ciphertext $CIPHERTEXT
```
