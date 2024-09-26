FROM golang:1.23.1-alpine as build
COPY foobar-enclave /foobar-enclave
COPY foobar-shared /foobar-shared
RUN cd /foobar-enclave && \
    go build .

FROM scratch

COPY --from=build /foobar-enclave/foobar-enclave /bin/foobar-enclave
COPY foobar-enclave/aws-ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["/bin/foobar-enclave"]
