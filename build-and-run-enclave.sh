#!/bin/bash
set -ex

nitro-cli terminate-enclave --all
docker build -t foobar-enclave .
nitro-cli build-enclave --docker-uri foobar-enclave:latest --output-file foobar-enclave.eif

# Uncomment when debugging:
# nitro-cli run-enclave --cpu-count 2 --memory 512 --enclave-cid 16 --eif-path foobar-enclave.eif --debug-mode
# nitro-cli console --enclave-name foobar-enclave

# comment when debugging:
nitro-cli run-enclave --cpu-count 2 --memory 512 --enclave-cid 16 --eif-path foobar-enclave.eif
nitro-cli describe-enclaves
