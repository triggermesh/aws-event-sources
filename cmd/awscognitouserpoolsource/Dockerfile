FROM golang:1.14-stretch AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/awscognitouserpoolsource

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN BIN_OUTPUT_DIR=/bin make awscognitouserpoolsource && \
    mkdir /kodata && \
    mv .git/* /kodata/ && \
    rm -rf ${GOPATH} && \
    rm -rf ${HOME}/.cache


FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL name "Triggermesh AWS Cognito User Pool Event Source"
LABEL vendor "Triggermesh"
LABEL version "v0.4.1"
LABEL release "1"
LABEL summary "The Triggermesh Cognito User Pool Source"
LABEL description "This is the Triggermesh Knative Event Source for AWS Cognito User Pool"

# Emulate ko builds
# https://github.com/google/ko/blob/v0.5.0/README.md#including-static-assets
ENV KO_DATA_PATH /kodata

COPY --from=builder /kodata/ ${KO_DATA_PATH}/
COPY --from=builder /bin/awscognitouserpoolsource /
COPY licenses/ /licenses/

ENTRYPOINT ["/awscognitouserpoolsource"]
