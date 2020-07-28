FROM golang:1.14-stretch AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/aws-event-sources-controller

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN BIN_OUTPUT_DIR=/bin make aws-event-sources-controller && \
    mkdir /kodata && \
    mv .git/* /kodata/ && \
    rm -rf ${GOPATH} && \
    rm -rf ${HOME}/.cache

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.2

LABEL name="AWS Event Sources Controller"
LABEL vendor="Triggermesh Inc."
LABEL version="v0.4.1"
LABEL release="1"
LABEL summary="AWS Event Sources controller for Kubernetes"
LABEL description="aws-event-sources is a Kubernetes controller that implements Knative event sources for AWS services"

# Emulate ko builds
# https://github.com/google/ko/blob/v0.5.0/README.md#including-static-assets
ENV KO_DATA_PATH /kodata

COPY --from=builder /kodata/ ${KO_DATA_PATH}/
COPY --from=builder /bin/aws-event-sources-controller /
COPY licenses/ /licenses/

ENTRYPOINT ["/aws-event-sources-controller"]
