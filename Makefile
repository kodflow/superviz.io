# Project Metadata
GOVERSION := $(shell go version | awk '{print $$3}')
BUILT_BY  := local
OS        := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH      := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
.SILENT:

# Ensure Go tools are in PATH
export PATH := $(HOME)/go/bin:$(PATH)

# Ensure GoReleaser is installed
# Make args forwarding
ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(ARGS):;@:)

# Envs for GoReleaser
ENV_EXPORTS := BUILT_BY=$(BUILT_BY) GOVERSION=$(GOVERSION) OS=$(OS) ARCH=$(ARCH)

.PHONY: help fmt test build run go-build

.DEFAULT_GOAL = help

help: ## Display all commands available
	$(eval PADDING=$(shell grep -x -E '^[a-zA-Z_-]+:.*?##[\s]?.*$$' Makefile | awk '{ print length($$1)-1 }' | sort -n | tail -n 1))
	clear
	echo '╔──────────────────────────────────────────────────╗'
	echo '║ ██╗  ██╗███████╗██╗     ██████╗ ███████╗██████╗  ║'
	echo '║ ██║  ██║██╔════╝██║     ██╔══██╗██╔════╝██╔══██╗ ║'
	echo '║ ███████║█████╗  ██║     ██████╔╝█████╗  ██████╔╝ ║'
	echo '║ ██╔══██║██╔══╝  ██║     ██╔═══╝ ██╔══╝  ██╔══██╗ ║'
	echo '║ ██║  ██║███████╗███████╗██║     ███████╗██║  ██║ ║'
	echo '║ ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝     ╚══════╝╚═╝  ╚═╝ ║'
	echo '╟──────────────────────────────────────────────────╝'
	grep -E '^[a-zA-Z_-]+:.*?##[\s]?.*$$' Makefile | awk 'BEGIN {FS = ":.*?##"}; {gsub(/(^ +| +$$)/, "", $$2);printf "╟─[ \033[36m%-$(PADDING)s\033[0m %s\n", $$1, "] "$$2}'
	echo '╚──────────────────────────────────────────────────>'
	echo ''

fmt: ## Format all code: Go, Terraform, YAML, Bazel
	echo "🔧 Formatting Go files..."
	go fmt ./...
	echo "🔧 Formatting Terraform files..."
	terraform fmt -recursive .
	echo "🔧 Formatting YAML and JSON files..."
	prettier --write "**/*.yml" "**/*.yaml" "**/*.json" "**/*.md"

test: ## Run all tests (unit tests, linting, and E2E tests)
	echo "🧪 Running linter..."
	golangci-lint run ./...
	echo "🧪 Running unit tests..."
	gotestsum --packages ./... -f github-actions -- -v -coverprofile=./coverage.out -covermode=atomic
	echo "🧪 Running E2E tests..."
	$(MAKE) _e2e-test

test-unit: ## Run only unit tests and linting (no Docker required)
	echo "🧪 Running linter..."
	golangci-lint run ./...
	echo "🧪 Running unit tests..."
	gotestsum --packages ./... -f github-actions -- -v -coverprofile=./coverage.out -covermode=atomic

test-basic: ## Run basic functionality tests (no Docker required)
	echo "🧪 Running basic functionality tests..."
	go test ./... -v

build: ## Build the Go application with GoReleaser
	echo "🚀 Building with GoReleaser..."
	$(ENV_EXPORTS) goreleaser build --snapshot --clean

run: ## Run the Go application
	go run $(CURDIR)/cmd/svz/main.go $(ARGS)

generate-copilot: fmt ## Generate copilot instructions from sectioned files
	echo "🔧 Generating copilot instructions..."
	{ \
		echo '````instructions'; \
		for file in .github/copilot-sections/*.md; do \
			[ -f "$$file" ] || continue; \
			[ "$$file" != ".github/copilot-sections/01-prime-directive.md" ] && printf "\n---\n\n"; \
			if [ "$$(basename "$$file")" = "01-prime-directive.md" ]; then \
				sed '1s/^## /# /' "$$file"; \
			else \
				cat "$$file"; \
			fi; \
		done; \
		echo '````'; \
	} > .github/copilot-instructions.md && echo "✅ Generated .github/copilot-instructions.md"

# E2E test targets (internal - not shown in help)
_e2e-setup:
	echo "🐳 Setting up E2E test environment..."
	cd test/e2e/docker && docker-compose build

_e2e-test:
	echo "🧪 Running comprehensive E2E tests on all distributions..."
	./test/e2e/final_e2e_test.sh

_e2e-test-single:
	echo "🧪 Running E2E test on $(or $(DISTRO),ubuntu)..."
	./test/e2e/final_e2e_test.sh $(or $(DISTRO),ubuntu)

_e2e-clean:
	echo "🧹 Cleaning up E2E test environment..."
	docker ps -aq --filter "name=svz-" | xargs -r docker rm -f || true
	docker images --filter "reference=svz-test-*" -q | xargs -r docker rmi -f || true

# E2E development targets (for manual testing)
e2e-setup: _e2e-setup  ## Setup E2E test environment (for development)

e2e-test-single: _e2e-test-single  ## Run E2E test on single distribution (usage: make e2e-test-single DISTRO=ubuntu)

e2e-clean: _e2e-clean  ## Clean up E2E test environment
