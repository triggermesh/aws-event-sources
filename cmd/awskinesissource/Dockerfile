FROM golang:1.14-alpine AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awskinesissource

RUN apk add --no-cache openssh-client ca-certificates

COPY ./ .
RUN go install -installsuffix cgo ./cmd/awskinesissource

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh AWS Kinesis Event Source"
LABEL vendor "Triggermesh"
LABEL version "v0.1.0"
LABEL release "1"
LABEL summary "The Triggermesh Kinesis Source"
LABEL description "This is the Triggermesh Knative Event Source for AWS Kinesis"

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awskinesissource /go/bin/awskinesissource

ENTRYPOINT ["/go/bin/awskinesissource"]
