.PHONY: validate-docs
validate-docs:
	./tools/scripts/validate-docs.sh

.PHONY: validate-architecture
validate-architecture:
	./tools/scripts/validate-architecture.sh

.PHONY: validate-naming
validate-naming:
	./tools/scripts/validate-naming.sh

.PHONY: validate-quality
validate-quality:
	./tools/scripts/scan-quality.sh

.PHONY: validate-security
validate-security:
	@$(foreach d,$(MODULES),(cd $(d) && echo "Executing [govulncheck] on $(d):" && govulncheck ./...) &&) true

.PHONY: update-doc-hashes
update-doc-hashes:
	./tools/scripts/update-doc-hashes.sh

.PHONY: guard
guard: validate-docs validate-architecture validate-naming validate-quality validate-security
