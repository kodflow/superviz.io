# Project Metadata
GOVERSION := $(shell go version | awk '{print $$3}')
BUILT_BY  := local
OS        := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH      := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
.SILENT:
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

test: ## Run all tests
	echo "🧪 Running linter..."
	golangci-lint run ./...
	echo "🧪 Running tests..."
	gotestsum --packages ./... -f github-actions -- -v -coverprofile=./coverage.out -covermode=atomic

test-basic: ## Run basic functionality tests (no Docker required)
	echo "🧪 Running basic functionality tests..."
	./test/e2e/test_without_docker.sh

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

e2e-setup: ## Setup E2E test environment (build containers)
	echo "🐳 Setting up E2E test environment..."
	cd test/e2e/docker && docker-compose build

e2e-test: ## Run E2E tests on all distributions
	echo "🧪 Running E2E tests on all distributions..."
	./test/e2e/run_e2e_tests.sh

e2e-test-single: ## Run E2E test on single distribution (usage: make e2e-test-single DISTRO=ubuntu)
	echo "🧪 Running E2E test on $(or $(DISTRO),ubuntu)..."
	./test/e2e/test_single_distro.sh $(or $(DISTRO),ubuntu)

e2e-clean: ## Clean up E2E test environment
	echo "🧹 Cleaning up E2E test environment..."
	cd test/e2e/docker && docker-compose down --remove-orphans --volumes
	docker system prune -f --filter "label=com.docker.compose.project=docker"
