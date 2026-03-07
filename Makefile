.PHONY: download-dependencies install-lint run test build lint lint-fix clean

MODULES := racing api sport
GOLANGCI_LINT ?= golangci-lint
GOLANGCI_LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint@latest
GO_BIN := $(shell go env GOBIN)
ifeq ($(GO_BIN),)
GO_BIN := $(shell go env GOPATH)/bin
endif
export PATH := $(PATH):$(GO_BIN)

download-dependencies:
	@for module in $(MODULES); do \
		cd $$module && go mod download; \
		cd - >/dev/null; \
	done

install-lint:
	go install $(GOLANGCI_LINT_PKG)
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || (echo "golangci-lint not found in PATH after install" && exit 1)

run: download-dependencies
	@echo "Starting racing, sport and api services..."
	@trap 'kill 0' EXIT; \
		(cd racing && go run .) & \
		(cd sport && go run .) & \
		(cd api && go run .) & \
		wait

test: download-dependencies
	@set -e; for module in $(MODULES); do \
		cd $$module; \
		go test ./...; \
		cd - >/dev/null; \
	done

build: download-dependencies
	@for module in $(MODULES); do \
		cd $$module && go build ./...; \
		cd - >/dev/null; \
	done

lint: download-dependencies install-lint
	@echo "Running lint checks..."
	@for module in $(MODULES); do \
		cd $$module && $(GOLANGCI_LINT) run ./...; \
		cd - >/dev/null; \
	done

lint-fix: download-dependencies install-lint
	@echo "Applying lint fixes..."
	@for module in $(MODULES); do \
		cd $$module && $(GOLANGCI_LINT) run --fix ./...; \
		cd - >/dev/null; \
	done

clean:
	@for module in $(MODULES); do \
		cd $$module && go clean ./...; \
		cd - >/dev/null; \
	done
	rm -f racing/racing api/api
