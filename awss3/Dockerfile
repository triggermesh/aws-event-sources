FROM golang:alpine AS builder

COPY awss3 /go/src/awss3
WORKDIR /go/src/awss3

RUN apk update && apk add --no-cache git openssh-client ca-certificates
RUN go get github.com/golang/dep/cmd/dep && \
    dep ensure -v && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/awss3

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awss3 /go/bin/awss3

ENTRYPOINT ["/go/bin/awss3"]
