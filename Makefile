# Project Metadata
GOVERSION := $(shell go version | awk '{print $$3}')
BUILT_BY  := local
OS        := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH      := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
.SILENT:

# Ensure Go tools are in PATH
export PATH := $(HOME)/go/bin:$(PATH)

# Make args forwarding
ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(ARGS):;@:)

.PHONY: help fmt test build

.DEFAULT_GOAL = help

help: ## Display all commands available
	$(eval PADDING=$(shell grep -E '^[a-zA-Z_-]+:.*##.*$$' Makefile | awk '{ print length($$1)-1 }' | sort -n | tail -n 1))
	clear
	echo '╔──────────────────────────────────────────────────╗'
	echo '║ ██╗  ██╗███████╗██╗     ██████╗ ███████╗██████╗  ║'
	echo '║ ██║  ██║██╔════╝██║     ██╔══██╗██╔════╝██╔══██╗ ║'
	echo '║ ███████║█████╗  ██║     ██████╔╝█████╗  ██████╔╝ ║'
	echo '║ ██╔══██║██╔══╝  ██║     ██╔═══╝ ██╔══╝  ██╔══██╗ ║'
	echo '║ ██║  ██║███████╗███████╗██║     ███████╗██║  ██║ ║'
	echo '║ ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝     ╚══════╝╚═╝  ╚═╝ ║'
	echo '╟──────────────────────────────────────────────────╝'
	grep -E '^[a-zA-Z_-]+:.*##.*$$' Makefile | awk 'BEGIN {FS = ":.*##"}; {gsub(/(^ +| +$$)/, "", $$2);printf "╟─[ \033[36m%-$(PADDING)s\033[0m %s\n", $$1, "] "$$2}'
	echo '╚──────────────────────────────────────────────────>'
	echo ''

fmt: ## Format all code: Go, Terraform, YAML, Bazel
	echo "🔧 Formatting Go files..."
	go fmt ./...
	echo "🔧 Formatting Bazel BUILD files..."
	bazel run //:gazelle
	echo "🔧 Formatting Bazel files with buildifier..."
	find . -name "*.bzl" -not -name "build_vars.bzl" -exec buildifier {} \;
	find . -name "BUILD" -o -name "BUILD.bazel" -exec buildifier {} \;
	echo "🔧 Formatting Terraform files..."
	terraform fmt -recursive .
	echo "🔧 Formatting YAML and JSON files..."
	prettier --write "**/*.yml" "**/*.yaml" "**/*.json" "**/*.md"

test: ## Run all tests (unit tests only)
	echo "🧪 Running all tests..."
	$(MAKE) test-unit

test-unit: ## Run unit tests only
	echo "🧪 Running unit tests..."
	bazel test //... --test_output=errors --test_tag_filters=unit

test-e2e: ## E2E tests have been removed (unit tests only)
	@echo "ℹ️  E2E tests have been removed from this project"
	@echo "ℹ️  Only unit tests are available - use 'make test-unit' or 'make test'"

build: ## Build cross-platform binaries for all supported platforms
	echo "🚀 Building cross-platform binaries..."
	echo "📦 Creating .dist/bin directory..."
	mkdir -p .dist/bin
	echo "� Building all platforms with Bazel..."
	bazel build //cmd/svz:svz_linux_amd64 //cmd/svz:svz_linux_arm64 //cmd/svz:svz_darwin_amd64 //cmd/svz:svz_darwin_arm64 //cmd/svz:svz_windows_amd64 //cmd/svz:svz_windows_arm64
	echo "📋 Copying binaries to .dist/bin/..."
	cp bazel-bin/cmd/svz/svz_linux_amd64_/svz_linux_amd64 .dist/bin/svz-linux-amd64
	cp bazel-bin/cmd/svz/svz_linux_arm64_/svz_linux_arm64 .dist/bin/svz-linux-arm64
	cp bazel-bin/cmd/svz/svz_darwin_amd64_/svz_darwin_amd64 .dist/bin/svz-darwin-amd64
	cp bazel-bin/cmd/svz/svz_darwin_arm64_/svz_darwin_arm64 .dist/bin/svz-darwin-arm64
	cp bazel-bin/cmd/svz/svz_windows_amd64_/svz_windows_amd64.exe .dist/bin/svz-windows-amd64.exe
	cp bazel-bin/cmd/svz/svz_windows_arm64_/svz_windows_arm64.exe .dist/bin/svz-windows-arm64.exe
	# Create a default 'svz' symlink to Linux AMD64 for convenience
	ln -sf svz-linux-amd64 .dist/bin/svz
	echo "✅ Cross-platform build completed!"
	echo "📁 Binaries available in .dist/bin/:"
	ls -la .dist/bin/

update: ## Update all dependencies (Go modules, Bazel, tools)
	echo "🔄 Updating all dependencies and tools..."
	echo "📦 Updating Go modules..."
	go get -u ./...
	go mod tidy
	echo "🔧 Updating Bazel dependencies..."
	bazel run //:gazelle-update-repos
	echo "🛠️  Updating Go rules for Bazel..."
	bazel run //:gazelle
	echo "🎯 Running tests to verify updates..."
	make test
	echo "✅ All dependencies updated successfully!"

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
