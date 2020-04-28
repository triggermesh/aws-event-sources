PACKAGE=aws-event-sources
SUBDIRS ?= cmd/awscodecommitsource cmd/awscognitosource cmd/awsdynamodbsource cmd/awskinesissource cmd/awssqssource
TARGETS := $(shell awk '{FS = ":";} /^[a-zA-Z0-9._-]+:.*?/ { printf "%s ", $$1 }' scripts/inc.Target)

.PHONY: $(SUBDIRS) $(TARGETS)

$(TARGETS): $(SUBDIRS)

$(SUBDIRS):
	@$(MAKE) -C $@ $(MAKECMDGOALS)

# Code generation
include scripts/inc.Codegen

include scripts/inc.Makefile

GOPKGS = ./pkg/adapter/... ./pkg/apis/... ./pkg/reconciler/...

test:
	$(GOTEST) -p=1 -cover -coverprofile=c.out $(GOPKGS)

coverage: ## Generate code coverage
	$(GOTOOL) cover -html=c.out -o $(OUTPUT_DIR)$(PACKAGE)-coverage.html

lint:
	$(GOLINT) $(GOPKGS)

vet:
	$(GO) vet $(GOPKGS)

fmt:
	$(GOFMT) -s -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GOPKGS))

fmt-test:
	@test -z $(shell $(GOFMT) -l $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GOPKGS)))

clean:
	$(RM) -rf $(PACKAGE)-unit-tests.xml
	$(RM) -rf c.out $(PACKAGE)-coverage.html
