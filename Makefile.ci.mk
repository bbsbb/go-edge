APPS = $(wildcard apps/*)
MODULES = ./core $(APPS)

.PHONY: ci-test-all
ci-test-all:
	@$(foreach d,$(MODULES),(cd $(d) && make test) &&) true

.PHONY: ci-lint-all
ci-lint-all:
	@$(foreach d,$(MODULES), \
		(cd $(d) && \
		echo "Executing [golangci-lint] on $(d):" && \
		golangci-lint run --timeout 10m --path-prefix $(d) && \
		echo "Executing [tidy + diff]   on $(d):" && \
		go mod tidy && \
		git diff --exit-code -- go.mod go.sum) &&) true

.PHONY: ci-build-all
ci-build-all:
ifeq ($(APPS),)
	@echo "ci-build-all: no apps found in apps/, nothing to build."
else
	@$(foreach d,$(APPS),(cd $(d) && make build) &&) true
endif
