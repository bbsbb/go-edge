APPS = $(wildcard apps/*)
MODULES = ./core $(APPS)

include ./Makefile.ci.mk
include ./Makefile.setup.mk
include ./Makefile.codegen.mk
include ./Makefile.observability.mk
include ./Makefile.validate.mk
include ./Makefile.docker.mk

.PHONY: ci
ci: ci-lint-all ci-test-all ci-build-all

.PHONY: tidy
tidy:
	@$(foreach d,$(MODULES),(cd $(d) && go mod tidy) &&) true
