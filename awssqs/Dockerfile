FROM golang:alpine AS builder

COPY awssqs /go/src/sqs
WORKDIR /go/src/sqs

RUN apk update && apk add --no-cache openssh-client ca-certificates
RUN GO111MODULE=on go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/sqs

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name="Triggermesh AWS SQS Event Source" \
      vendor="Triggermesh" \
      version="v0.1.0" \
      release="1" \
      summary="The Triggermesh SQS Source" \
      description="This is the Triggermesh Knative Event Source for AWS SQS"

RUN microdnf update
COPY ./licenses/* /licenses/

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/sqs /go/bin/sqs

ENTRYPOINT ["/go/bin/sqs"]
