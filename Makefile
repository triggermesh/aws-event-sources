SUBDIRS ?= cmd/awscodecommitsource cmd/awscognitosource cmd/awsdynamodbsource awskinesis awssqs
TARGETS := $(shell awk '{FS = ":";} /^[a-zA-Z0-9._-]+:.*?/ { printf "%s ", $$1 }' scripts/inc.Makefile)

.PHONY: $(SUBDIRS) $(TARGETS)

$(TARGETS): $(SUBDIRS)

$(SUBDIRS):
	@$(MAKE) -C $@ $(MAKECMDGOALS)

# Code generation
include scripts/Makefile.codegen
