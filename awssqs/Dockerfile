FROM golang:1.14-alpine AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awssqs

RUN apk update && apk add --no-cache openssh-client ca-certificates

COPY awssqs/go.mod awssqs/go.sum /go/src/awssqs/
RUN go mod download

COPY awssqs/ /go/src/awssqs/
RUN go install -installsuffix cgo ./...

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh AWS SQS Event Source"
LABEL vendor "Triggermesh"
LABEL version "v0.1.0"
LABEL release "1"
LABEL summary "The Triggermesh SQS Source"
LABEL description "This is the Triggermesh Knative Event Source for AWS SQS"

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awssqs /go/bin/awssqs

ENTRYPOINT ["/go/bin/awssqs"]
