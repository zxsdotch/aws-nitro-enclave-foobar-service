package constants

// Context identifier for the enclave. This value must match the value used
// with `nitro-cli run-enclave`. It might be possible to bind to 0xFFFFFFFF
// but using the enclave's specific CID seems cleaner.
const ENCLAVE_CID = 16

// Port the enclave listens on for commands.
const ENCLAVE_LISTENING_PORT = 1000

// Context identifier for the parent instance. This is always 3.
const INSTANCE_CID = 3

// Port the parent instance listens on and forward data to KMS.
const INSTANCE_LISTENING_PORT = 1001
