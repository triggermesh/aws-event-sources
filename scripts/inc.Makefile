RM         ?= rm
CP         ?= cp
MV         ?= mv

GO         ?= go
GOFMT      ?= gofmt
GOLINT     ?= golint
GOTEST     ?= go test
GOTOOL     ?= go tool

DOCKER     ?= docker

LDFLAGS     =

all: build

help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "; printf "\n$(PACKAGE_DESC)\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9._-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

mod-download: ## Download modules using 'go mod download'
	$(GO) mod download

build: ## Build the binary using 'go build'
	$(GO) build -ldflags "$(LDFLAGS)" -installsuffix cgo -o $(PACKAGE)

install: ## Install the binary using the 'go install'
	$(GO) install -ldflags "$(LDFLAGS)" -installsuffix cgo

test: ## Run unit tests
	$(GOTEST) ./...

coverage: ## Generate code coverage
	@$(GOTEST) -coverprofile=c.out ./...
	@$(GOTOOL) cover -html=c.out -o coverage.html

lint: ## Link source files using 'golint'
	$(GOLINT) ./...

vet: ## Vet source files using 'go vet'
	$(GO) vet ./...

fmt: ## Format source files using 'gofmt'
	$(GOFMT) -s -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}}{{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}}{{end}}' ./...)

fmt-test: ## Check source formatting using 'gofmt'
	@test -z $(shell $(GOFMT) -l $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}}{{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}}{{end}}' ./...))

image: ## Build docker image using 'docker build'
	$(DOCKER) build -t gcr.io/triggermesh/$(PACKAGE) -f Dockerfile ../

clean: ## Clean build artifacts
	@$(RM) -rf $(PACKAGE)
	@$(RM) -rf c.out coverage.html