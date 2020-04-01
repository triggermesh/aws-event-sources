SUBDIRS ?= awscodecommit awscognito awsdynamodb awskinesis awssqs
TARGETS := $(shell awk '{FS = ":";} /^[a-zA-Z0-9._-]+:.*?/ { printf "%s ", $$1 }' scripts/inc.Makefile)

.PHONY: $(SUBDIRS) $(TARGETS)

$(TARGETS): $(SUBDIRS)

$(SUBDIRS):
	@$(MAKE) --no-print-directory -C $@ $(MAKECMDGOALS)
