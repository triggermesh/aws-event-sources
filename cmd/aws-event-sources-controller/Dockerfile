FROM golang:1.14-alpine AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/aws-event-sources-controller

RUN apk add --no-cache openssh-client ca-certificates

COPY ./ .
RUN go install -installsuffix cgo ./cmd/aws-event-sources-controller

FROM scratch

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/aws-event-sources-controller /go/bin/aws-event-sources-controller

ENTRYPOINT ["/go/bin/aws-event-sources-controller"]
