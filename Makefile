.PHONY: help fmt test build
.SILENT:

.DEFAULT_GOAL = help
GOVERSION := $(shell go version | awk '{print $$3}')
BUILT_BY  := local
ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(ARGS):;@:)

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

build: ## Build the Go application
	echo "🚀 Building the Go application..."
	BUILT_BY=$(BUILT_BY) GOVERSION=$(GOVERSION) goreleaser build --snapshot --clean

run: ## Run the Go application
	go run cmd/svz/main.go $(ARGS)