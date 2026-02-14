# Tool versions — pin exact versions, never use "latest".
GOLANGCI_LINT_VERSION := v2.8.0
GOTESTSUM_VERSION := v1.13.0
MOCKERY_VERSION := v2.53.5
SQLC_VERSION := v1.30.0
GOVULNCHECK_VERSION := v1.1.4
GOFUMPT_VERSION := v0.9.2

# Required tools
REQUIRED_TOOLS := go docker golangci-lint

.PHONY: setup-check
setup-check: check-go check-docker check-golangci-lint check-gotestsum check-mockery check-govulncheck check-gofumpt
	@echo "All required tools are installed."

.PHONY: check-go
check-go:
	@command -v go >/dev/null 2>&1 || { echo "Error: go is not installed. Visit https://go.dev/dl/"; exit 1; }
	@echo "go: $$(go version)"

.PHONY: check-docker
check-docker:
	@command -v docker >/dev/null 2>&1 || { echo "Error: docker is not installed. Visit https://docs.docker.com/get-docker/"; exit 1; }
	@echo "docker: $$(docker --version)"

.PHONY: check-golangci-lint
check-golangci-lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Error: golangci-lint is not installed. Run 'make install-golangci-lint'"; exit 1; }
	@echo "golangci-lint: $$(golangci-lint --version | head -n1)"

.PHONY: check-gotestsum
check-gotestsum:
	@command -v gotestsum >/dev/null 2>&1 || { echo "Error: gotestsum is not installed. Run 'make install-gotestsum'"; exit 1; }
	@echo "gotestsum: $$(gotestsum --version)"

.PHONY: check-mockery
check-mockery:
	@command -v mockery >/dev/null 2>&1 || { echo "Error: mockery is not installed. Run 'make install-mockery'"; exit 1; }
	@echo "mockery: $$(mockery --version)"

.PHONY: install-golangci-lint
install-golangci-lint:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@mkdir -p $$(go env GOPATH)/bin
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "golangci-lint installed successfully."

.PHONY: install-gotestsum
install-gotestsum:
	@echo "Installing gotestsum..."
	go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)
	@echo "gotestsum installed successfully."

.PHONY: install-mockery
install-mockery:
	@echo "Installing mockery $(MOCKERY_VERSION)..."
	go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION)
	@echo "mockery installed successfully."

.PHONY: check-sqlc
check-sqlc:
	@command -v sqlc >/dev/null 2>&1 || { echo "Error: sqlc is not installed. Run 'make install-sqlc'"; exit 1; }
	@echo "sqlc: $$(sqlc version)"

.PHONY: install-sqlc
install-sqlc:
	@echo "Installing sqlc $(SQLC_VERSION)..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	@echo "sqlc installed successfully."

.PHONY: check-govulncheck
check-govulncheck:
	@command -v govulncheck >/dev/null 2>&1 || { echo "Error: govulncheck is not installed. Run 'make install-govulncheck'"; exit 1; }
	@echo "govulncheck: $$(govulncheck -version 2>&1 | head -n1)"

.PHONY: install-govulncheck
install-govulncheck:
	@echo "Installing govulncheck..."
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	@echo "govulncheck installed successfully."

.PHONY: check-gofumpt
check-gofumpt:
	@command -v gofumpt >/dev/null 2>&1 || { echo "Error: gofumpt is not installed. Run 'make install-gofumpt'"; exit 1; }
	@echo "gofumpt: $$(gofumpt --version 2>&1 || echo 'installed')"

.PHONY: install-gofumpt
install-gofumpt:
	@echo "Installing gofumpt..."
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	@echo "gofumpt installed successfully."

.PHONY: install-hooks
install-hooks:
	@echo "Installing git hooks..."
	@ln -sf ../../tools/hooks/pre-commit .git/hooks/pre-commit
	@ln -sf ../../tools/hooks/pre-push .git/hooks/pre-push
	@echo "Git hooks installed successfully."

.PHONY: install-tools
install-tools: install-golangci-lint install-gotestsum install-mockery install-sqlc install-govulncheck install-gofumpt install-hooks
	@echo "All tools installed successfully."
