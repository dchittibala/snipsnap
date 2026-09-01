MODULE      := github.com/yourusername/snipsnap
BINARY_NAME := snipsnap
CMD_DIR     := ./cmd/snipsnap

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS    := -X '$(MODULE)/pkg/fileops.Version=$(VERSION)' \
              -X '$(MODULE)/pkg/fileops.Commit=$(COMMIT)' \
              -X '$(MODULE)/pkg/fileops.BuildDate=$(BUILD_DATE)' \
              -s -w

.PHONY: all build test test-race lint init-hooks verify-hooks clean

all: init-hooks test build

## build: Compiles the binary locally with version injection
build:
	@echo "==> Building $(BINARY_NAME) $(VERSION)..."
	@go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) $(CMD_DIR)

## test: Runs unit tests
test:
	@echo "==> Running unit tests..."
	@go test -v ./...

## test-race: Runs unit tests with race detection
test-race:
	@echo "==> Running tests with race detector..."
	@go test -v -race -cover ./...

## lint: Runs golangci-lint
lint:
	@echo "==> Running linters..."
	@golangci-lint run ./...

## init-hooks: Installs pre-commit CLI (if missing) and sets up git hook stages
init-hooks:
	@echo "==> Checking for pre-commit installation..."
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		echo "==> pre-commit not found. Installing pre-commit..."; \
		if command -v brew >/dev/null 2>&1; then \
			brew install pre-commit; \
		elif command -v pip3 >/dev/null 2>&1; then \
			pip3 install pre-commit; \
		elif command -v pip >/dev/null 2>&1; then \
			pip install pre-commit; \
		else \
			echo "ERROR: Neither brew nor pip found. Please install pre-commit manually: https://pre-commit.com/"; \
			exit 1; \
		fi; \
	fi
	@echo "==> Installing Git hooks..."
	@pre-commit install --hook-type commit-msg --hook-type pre-commit
	@echo "==> Pre-commit hooks successfully active!"

## verify-hooks: Runs pre-commit checks manually against all files
verify-hooks:
	@pre-commit run --all-files

## clean: Cleans build artifacts
clean:
	@echo "==> Cleaning artifacts..."
	@rm -rf bin/ dist/