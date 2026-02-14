APPS = $(wildcard apps/*)

# sqlc targets iterate over apps/. No-op when apps/ is empty.
.PHONY: sqlc-generate
sqlc-generate:
ifeq ($(APPS),)
	@echo "sqlc-generate: no apps found in apps/, nothing to generate."
else
	@$(foreach d,$(APPS),(cd $(d) && sqlc generate) &&) true
endif

.PHONY: sqlc-vet
sqlc-vet:
ifeq ($(APPS),)
	@echo "sqlc-vet: no apps found in apps/, nothing to vet."
else
	@$(foreach d,$(APPS),(cd $(d) && sqlc vet) &&) true
endif

.PHONY: docs-schema
docs-schema:
	./tools/scripts/generate-schema-doc.sh
