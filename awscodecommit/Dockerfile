FROM golang:alpine AS builder

COPY awscodecommit /go/src/awscodecommit
WORKDIR /go/src/awscodecommit

RUN apk update && apk add --no-cache git openssh-client ca-certificates
RUN go get github.com/golang/dep/cmd/dep && \
    dep ensure -v && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/awscodecommit

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/awscodecommit /go/bin/awscodecommit

ENTRYPOINT ["/go/bin/awscodecommit"]
