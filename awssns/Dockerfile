FROM golang:1.14-alpine AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awssns

RUN apk update && apk add --no-cache openssh-client ca-certificates

COPY awssns/go.mod awssns/go.sum /go/src/awssns/
RUN go mod download

COPY awssns/ /go/src/awssns/
RUN go install -installsuffix cgo ./...

FROM scratch

COPY licenses/ /licenses/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awssns /go/bin/awssns

ENTRYPOINT ["/go/bin/awssns"]
