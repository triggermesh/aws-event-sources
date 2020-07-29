FROM quay.io/operator-framework/helm-operator:v0.18.2

LABEL name="AWS Event Sources Operator"
LABEL vendor="Triggermesh Inc."
LABEL version="v0.1.0"
LABEL release="1"
LABEL summary="AWS Event Sources controller for Kubernetes"
LABEL description="aws-event-sources is a Kubernetes controller that implements Knative event sources for AWS services"

COPY licenses/ /licenses/
COPY watches.yaml ${HOME}/watches.yaml
COPY helm-charts/ ${HOME}/helm-charts/
