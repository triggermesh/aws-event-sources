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
GOLINT     ?= golangci-lint run
GOTOOL     ?= go tool
GOTEST     ?= gotestsum --junitfile $(OUTPUT_DIR)$(PACKAGE)-unit-tests.xml --format pkgname-and-test-fails --

GOPKGS      = ./...
LDFLAGS     =
