# Project Metadata
GOVERSION := $(shell go version | awk '{print $$3}')
BUILT_BY  := local
OS        := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH      := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

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

build: ## Build the Go application with GoReleaser
	echo "🚀 Building with GoReleaser..."
	$(ENV_EXPORTS) goreleaser build --snapshot --clean

run: ## Run the Go application
	go run $(CURDIR)/cmd/svz/main.go $(ARGS)
