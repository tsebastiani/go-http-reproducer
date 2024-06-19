#FROM registry.access.redhat.com/ubi9/go-toolset:1.19
#FROM golang:1.22.4-alpine
FROM golang1.22-mod

#FROM registry-proxy.engineering.redhat.com/rh-osbs/openshift-golang-builder:rhel_8_golang_1.22
WORKDIR /opt/app-root/src

COPY server.go server.crt server.key ./

RUN go mod init github.com/jhutar/go-http-reproducer
RUN go mod tidy
RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o server server.go

USER 65532:65532

ENTRYPOINT /opt/app-root/src/server
