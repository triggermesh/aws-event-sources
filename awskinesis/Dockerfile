FROM golang:alpine AS builder

COPY awskinesis /go/src/awskinesis
WORKDIR /go/src/awskinesis

RUN apk update && apk add --no-cache git openssh-client ca-certificates
RUN go get github.com/golang/dep/cmd/dep && \
    dep ensure -v && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/awskinesis

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awskinesis /go/bin/awskinesis

ENTRYPOINT ["/go/bin/awskinesis"]
