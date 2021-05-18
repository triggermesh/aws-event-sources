KREPO              = aws-event-sources
KREPO_DESC         = TriggerMesh AWS event sources
COMMANDS           = aws-event-sources-controller awscloudwatchsource awscloudwatchlogssource awscodecommitsource awscognitoidentitysource awscognitouserpoolsource awsdynamodbsource awskinesissource awsperformanceinsightssource awssnssource awssqssource

TARGETS           ?= linux/amd64

BASE_DIR          ?= $(CURDIR)

OUTPUT_DIR        ?= $(BASE_DIR)/_output

BIN_OUTPUT_DIR    ?= $(OUTPUT_DIR)
TEST_OUTPUT_DIR   ?= $(OUTPUT_DIR)
COVER_OUTPUT_DIR  ?= $(OUTPUT_DIR)
DIST_DIR          ?= $(OUTPUT_DIR)

DOCKER            ?= docker
IMAGE_REPO        ?= gcr.io/triggermesh
IMAGE_TAG         ?= latest
IMAGE_SHA         ?= $(shell git rev-parse HEAD)

GO                ?= go
GOFMT             ?= gofmt
GOLINT            ?= golangci-lint run
GOTOOL            ?= go tool
GOTEST            ?= gotestsum --junitfile $(TEST_OUTPUT_DIR)/$(KREPO)-unit-tests.xml --format pkgname-and-test-fails --

GOPKGS             = ./cmd/... ./pkg/apis/... ./pkg/adapter/... ./pkg/reconciler/...
LDFLAGS            = -extldflags=-static -w -s

HAS_GOTESTSUM     := $(shell command -v gotestsum;)
HAS_GOLANGCI_LINT := $(shell command -v golangci-lint;)

.PHONY: help mod-download build install release test coverage lint fmt fmt-test images cloudbuild-test cloudbuild clean

all: build

install-gotestsum:
ifndef HAS_GOTESTSUM
	curl -SL https://github.com/gotestyourself/gotestsum/releases/download/v0.4.2/gotestsum_0.4.2_linux_amd64.tar.gz | tar -C $(shell go env GOPATH)/bin -zxf -
endif

install-golangci-lint:
ifndef HAS_GOLANGCI_LINT
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.26.0
endif

$(COMMANDS):
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_OUTPUT_DIR)/$@ ./cmd/$@

help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "; printf "\n$(KREPO_DESC)\nUsage:\n  make \033[36m<source>\033[0m\n"} /^[a-zA-Z0-9._-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

mod-download: ## Download go modules
	$(GO) mod download

build: $(COMMANDS) ## Build the binary

release: ## Build release binaries
	@set -e ; \
	for bin in $(COMMANDS) ; do \
		for platform in $(TARGETS); do \
			GOOS=$${platform%/*} ; \
			GOARCH=$${platform#*/} ; \
			RELEASE_BINARY=$$bin-$${GOOS}-$${GOARCH} ; \
			[ $${GOOS} = "windows" ] && RELEASE_BINARY=$${RELEASE_BINARY}.exe ; \
			echo "GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$${RELEASE_BINARY}" ./cmd/$$bin ; \
			GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$${RELEASE_BINARY} ./cmd/$$bin ; \
		done ; \
	done

test: install-gotestsum ## Run unit tests
	@mkdir -p $(TEST_OUTPUT_DIR)
	$(GOTEST) -p=1 -race -cover -coverprofile=$(TEST_OUTPUT_DIR)/$(KREPO)-c.out $(GOPKGS)

cover: test ## Generate code coverage
	@mkdir -p $(COVER_OUTPUT_DIR)
	$(GOTOOL) cover -html=$(TEST_OUTPUT_DIR)/$(KREPO)-c.out -o $(COVER_OUTPUT_DIR)/$(KREPO)-coverage.html

lint: install-golangci-lint ## Lint source files
	$(GOLINT) $(GOPKGS)

fmt: ## Format source files
	$(GOFMT) -s -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GOPKGS))

fmt-test: ## Check source formatting
	@test -z $(shell $(GOFMT) -l $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GOPKGS)))

IMAGES = $(foreach cmd,$(COMMANDS),$(cmd).image)
images: $(IMAGES) ## Builds container images
$(IMAGES): %.image:
	$(DOCKER) build -t $(IMAGE_REPO)/$* -f ./cmd/$*/Dockerfile . --build-arg VERSION=${IMAGE_TAG};

CLOUDBUILD_TEST = $(foreach cmd,$(COMMANDS),$(cmd).cloudbuild-test)
cloudbuild-test: $(CLOUDBUILD_TEST) ## Test container image build with Google Cloud Build
$(CLOUDBUILD_TEST): %.cloudbuild-test:
# NOTE (antoineco): Cloud Build started failing recently with authentication errors when --no-push is specified.
# Pushing images with the "_" tag is our hack to avoid those errors and ensure the build cache is always updated.
	gcloud builds submit $(BASE_DIR) --config cloudbuild.yaml --substitutions _CMD=$*,COMMIT_SHA=${IMAGE_SHA},_KANIKO_IMAGE_TAG=_

CLOUDBUILD = $(foreach cmd,$(COMMANDS),$(cmd).cloudbuild)
cloudbuild: $(CLOUDBUILD) ## Build and publish image to GCR
$(CLOUDBUILD): %.cloudbuild:
	gcloud builds submit $(BASE_DIR) --config cloudbuild.yaml --substitutions _CMD=$*,COMMIT_SHA=${IMAGE_SHA},_KANIKO_IMAGE_TAG=${IMAGE_TAG}

clean: ## Clean build artifacts
	@for bin in $(COMMANDS) ; do \
		for platform in $(TARGETS); do \
			GOOS=$${platform%/*} ; \
			GOARCH=$${platform#*/} ; \
			RELEASE_BINARY=$$bin-$${GOOS}-$${GOARCH} ; \
			[ $${GOOS} = "windows" ] && RELEASE_BINARY=$${RELEASE_BINARY}.exe ; \
			$(RM) -v $(DIST_DIR)/$${RELEASE_BINARY}; \
		done ; \
		$(RM) -v $(BIN_OUTPUT_DIR)/$$bin; \
	done
	@$(RM) -v $(TEST_OUTPUT_DIR)/$(KREPO)-c.out $(TEST_OUTPUT_DIR)/$(KREPO)-unit-tests.xml
	@$(RM) -v $(COVER_OUTPUT_DIR)/$(KREPO)-coverage.html

# Code generation
include $(BASE_DIR)/hack/inc.Codegen.mk
