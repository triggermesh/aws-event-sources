FROM golang:1.14-stretch AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awskinesissource

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN BIN_OUTPUT_DIR=/bin make awskinesissource && \
    mkdir /kodata && \
    mv .git/* /kodata/ && \
    rm -rf ${GOPATH} && \
    rm -rf ${HOME}/.cache

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh AWS Kinesis Event Source"
LABEL vendor "Triggermesh"
LABEL version="v0.4.2"
LABEL release "1"
LABEL summary "The Triggermesh Kinesis Source"
LABEL description "This is the Triggermesh Knative Event Source for AWS Kinesis"

# Emulate ko builds
# https://github.com/google/ko/blob/v0.5.0/README.md#including-static-assets
ENV KO_DATA_PATH /kodata

COPY --from=builder /kodata/ ${KO_DATA_PATH}/
COPY --from=builder /bin/awskinesissource /
COPY licenses/ /licenses/

ENTRYPOINT ["/awskinesissource"]
