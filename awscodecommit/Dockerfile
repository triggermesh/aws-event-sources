FROM golang:1.14-alpine AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awscodecommit

RUN apk update && apk add --no-cache openssh-client ca-certificates

COPY awscodecommit/go.mod awscodecommit/go.sum /go/src/awscodecommit/
RUN go mod download

COPY awscodecommit/ /go/src/awscodecommit/
RUN go install -installsuffix cgo ./...

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh AWS CodeCommit Event Source"
LABEL vendor "Triggermesh"
LABEL version "v0.1.0"
LABEL release "1"
LABEL summary "The Triggermesh CodeCommit Source"
LABEL description "This is the Triggermesh Knative Event Source for AWS CodeCommit"

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awscodecommit /go/bin/awscodecommit

ENTRYPOINT ["/go/bin/awscodecommit"]
