TARGETS    ?= darwin/amd64 linux/amd64

RM         ?= rm
CP         ?= cp
MV         ?= mv
MKDIR      ?= mkdir

DOCKER     ?= docker
IMAGE_REPO ?= gcr.io/triggermesh
IMAGE      ?= $(IMAGE_REPO)/$(PACKAGE)

GO         ?= go
GOFMT      ?= gofmt
GOLINT     ?= golint
GOTOOL     ?= go tool
GOTEST     ?= gotestsum --junitfile $(OUTPUT_DIR)$(PACKAGE)-unit-tests.xml --
LDFLAGS    +=

.PHONY: help mod-download build install release test coverage lint vet fmt fmt-test image clean

all: build

help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "; printf "\n$(PACKAGE_DESC)\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9._-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

mod-download: ## Download go modules
	$(GO) mod download

build: ## Build the binary
	$(GO) build -ldflags "$(LDFLAGS)" -o $(PACKAGE) -installsuffix cgo

install: ## Install the binary
	$(GO) install -ldflags "$(LDFLAGS)" -installsuffix cgo

release: ## Build release binaries
	@set -e ; \
	for platform in $(TARGETS); do \
		GOOS=$${platform%/*} ; \
		GOARCH=$${platform#*/} ; \
		RELEASE_BINARY=$(PACKAGE)-$${GOOS}-$${GOARCH} ; \
		[ $${GOOS} = "windows" ] && RELEASE_BINARY=$${RELEASE_BINARY}.exe ; \
		echo "GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)$${RELEASE_BINARY} -installsuffix cgo" ; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)$${RELEASE_BINARY} -installsuffix cgo ; \
	done

test: ## Run unit tests
	$(GOTEST) -cover -coverprofile=c.out ./...

coverage: ## Generate code coverage
	$(GOTOOL) cover -html=c.out -o $(OUTPUT_DIR)$(PACKAGE)-coverage.html

lint: ## Link source files
	$(GOLINT) ./...

vet: ## Vet source files
	$(GO) vet ./...

fmt: ## Format source files
	$(GOFMT) -s -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' ./...)

fmt-test: ## Check source formatting
	@test -z $(shell $(GOFMT) -l $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' ./...))

image: ## Builds the container image
	$(DOCKER) build -t $(IMAGE) -f Dockerfile ../

clean: ## Clean build artifacts
	$(RM) -rf $(PACKAGE)
	$(RM) -rf $(PACKAGE)-unit-tests.xml
	$(RM) -rf c.out $(PACKAGE)-coverage.html
	@for platform in $(TARGETS); do $(RM) -rf $(DIST_DIR)$(PACKAGE)-$${platform%/*}-$${platform#*/}; done
