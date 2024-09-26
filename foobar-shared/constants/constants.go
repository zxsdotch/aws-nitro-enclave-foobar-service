package constants

// Context identifier for the enclave. This value must match the value used
// with `nitro-cli run-enclave`.
const ENCLAVE_CID = 16

// Port the enclave listens on for commands.
const ENCLAVE_LISTENING_PORT = 1000

// Context identifier for the instance. This is always 3.
const INSTANCE_CID = 3

// Port the instance listens on and forward data to KMS.
const INSTANCE_LISTENING_PORT = 1001
