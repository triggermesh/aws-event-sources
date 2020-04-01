FROM golang:1.14-alpine AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awsdynamodb

RUN apk update && apk add --no-cache openssh-client ca-certificates

COPY awsdynamodb/go.mod awsdynamodb/go.sum /go/src/awsdynamodb/
RUN go mod download

COPY awsdynamodb/ /go/src/awsdynamodb/
RUN go install -installsuffix cgo ./...

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh AWS DynamoDB Event Source"
LABEL vendor "Triggermesh"
LABEL version "v0.1.0"
LABEL release "1"
LABEL summary "The Triggermesh DynamoDB Source"
LABEL description "This is the Triggermesh Knative Event Source for AWS DynamoDB"

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awsdynamodb /go/bin/awsdynamodb

ENTRYPOINT ["/go/bin/awsdynamodb"]
