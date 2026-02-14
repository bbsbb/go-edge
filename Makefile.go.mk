export REPOSITORY_ROOT := $(shell git rev-parse --show-cdup)

.PHONY: build
build:
	@go build -o bin/${APP} -v .

.PHONY: test-default
test-default:
	APP_ENVIRONMENT=testing gotestsum --format pkgname -- -cover -v -race -tags=testing ./...

.PHONY: lint
lint:
	@golangci-lint run --fix --config $(REPOSITORY_ROOT).golangci.yml

.PHONY: generate-mocks
generate-mocks:
	@mockery

.PHONY: deps-add
deps-add:
ifndef PKG
	$(error PKG is required. Usage: make deps-add PKG=example.com/pkg@version)
endif
	go get $(PKG)
	go mod tidy

.PHONY: deps-tidy
deps-tidy:
	go mod tidy
