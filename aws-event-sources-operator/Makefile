DOCKER      ?= docker
IMAGE_REPO  ?= gcr.io/triggermesh/aws-event-sources-operator
IMAGE_TAG   ?= latest

image:
	$(DOCKER) build -t $(IMAGE_REPO):$(IMAGE_TAG) -f build/Dockerfile .
