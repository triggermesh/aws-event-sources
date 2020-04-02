FROM golang:1.14-alpine AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awscognito

RUN apk update && apk add --no-cache openssh-client ca-certificates

COPY awscognito/go.mod awscognito/go.sum /go/src/awscognito/
RUN go mod download

COPY awscognito/ /go/src/awscognito/
RUN go install -installsuffix cgo ./...

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh AWS Cognito Event Source"
LABEL vendor "Triggermesh"
LABEL version "v0.1.0"
LABEL release "1"
LABEL summary "The Triggermesh Cognito Source"
LABEL description "This is the Triggermesh Knative Event Source for AWS Cognito"

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awscognito /go/bin/awscognito

ENTRYPOINT ["/go/bin/awscognito"]
