# Makefile for Go project

# Variables
BINARY_NAME=ebm-cli
GO=go
GOFLAGS=-v
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=$(GO) fmt
GOLINT=golangci-lint
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
RUN_ARGS?=

# Directories
CMD_DIR=./cmd/ebm
BUILD_DIR=./build
INSTALL_DIR?=/usr/local/bin
FIXTURES_REPAIR?=./test-library/repair-fixtures

# ANSI color codes
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
MAGENTA=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[0;37m
BOLD=\033[1m
RESET=\033[0m

.PHONY: help
help: ## Display this help message
	@echo ""
	@echo "$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BOLD)$(CYAN)  Available Make Targets$(RESET)"
	@echo "$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BOLD)$(GREEN)Build Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^build/ || /^install/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(BLUE)Test Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^test/ || /^coverage/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(MAGENTA)Quality Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^lint/ || /^fmt/ || /^vet/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(RED)Development Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^run/ || /^clean/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(CYAN)Docs Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; /^docs/ || /^wiki/ {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""

.PHONY: build
build: ## Build the application binary
	@echo "$(BOLD)$(GREEN)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(BOLD)$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: test
test: ## Run all tests
	@echo "$(BOLD)$(BLUE)Running all tests...$(RESET)"
	$(GOTEST) $(GOFLAGS) -race -timeout 5m ./...
	@echo "$(BOLD)$(BLUE)✓ All tests passed$(RESET)"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(BOLD)$(BLUE)Running unit tests...$(RESET)"
	$(GOTEST) $(GOFLAGS) -short -race -timeout 30s ./...
	@echo "$(BOLD)$(BLUE)✓ Unit tests passed$(RESET)"

.PHONY: coverage
coverage: ## Generate test coverage report
	@echo "$(BOLD)$(BLUE)Generating coverage report...$(RESET)"
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(BOLD)$(BLUE)✓ Coverage report generated: $(COVERAGE_HTML)$(RESET)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n 1

.PHONY: lint
lint: ## Run linter on the codebase
	@echo "$(BOLD)$(MAGENTA)Running linter...$(RESET)"
	@if command -v $(GOLINT) > /dev/null 2>&1; then \
		$(GOLINT) run ./...; \
		echo "$(BOLD)$(MAGENTA)✓ Linting complete$(RESET)"; \
	else \
		echo "$(BOLD)$(RED)✗ golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin$(RESET)"; \
		exit 1; \
	fi

.PHONY: fmt
fmt: ## Format code with go fmt
	@echo "$(BOLD)$(MAGENTA)Formatting code...$(RESET)"
	$(GOFMT) ./...
	@echo "$(BOLD)$(MAGENTA)✓ Code formatted$(RESET)"

.PHONY: vet
vet: ## Run go vet on the codebase
	@echo "$(BOLD)$(MAGENTA)Running go vet...$(RESET)"
	$(GOVET) ./...
	@echo "$(BOLD)$(MAGENTA)✓ Vet complete$(RESET)"

.PHONY: clean
clean: ## Clean build artifacts and cache
	@echo "$(BOLD)$(RED)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	$(GO) clean -cache -testcache -modcache
	@echo "$(BOLD)$(RED)✓ Clean complete$(RESET)"

.PHONY: install
install: ## Install dependencies
	@echo "$(BOLD)$(GREEN)Installing dependencies...$(RESET)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(BOLD)$(GREEN)✓ Dependencies installed$(RESET)"

.PHONY: install-cli
install-cli: build ## Install the CLI binary to INSTALL_DIR (default: /usr/local/bin)
	@echo "$(BOLD)$(GREEN)Installing $(BINARY_NAME) to $(INSTALL_DIR)...$(RESET)"
	@install -m 0755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(BOLD)$(GREEN)✓ Installed: $(INSTALL_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: uninstall-cli
uninstall-cli: ## Remove the CLI binary from INSTALL_DIR
	@echo "$(BOLD)$(RED)Removing $(BINARY_NAME) from $(INSTALL_DIR)...$(RESET)"
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(BOLD)$(GREEN)✓ Removed: $(INSTALL_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: run
run: ## Run the interactive TUI
	@echo "$(BOLD)$(RED)Running application (TUI)...$(RESET)"
	$(GO) run $(CMD_DIR)

.PHONY: run-cli
run-cli: ## Run the CLI validation (usage: make run-cli RUN_ARGS="path/to/file")
	@echo "$(BOLD)$(RED)Running application (CLI)...$(RESET)"
	@if [ -z "$(RUN_ARGS)" ]; then \
		echo "$(BOLD)$(YELLOW)Usage: make run-cli RUN_ARGS=\"path/to/file\"$(RESET)"; \
		exit 1; \
	fi
	$(GO) run $(CMD_DIR) $(RUN_ARGS)

.PHONY: run-repair
run-repair: ## Run single-file repair (usage: make run-repair RUN_ARGS="book.epub --no-backup --aggressive")
	@echo "$(BOLD)$(RED)Running repair...$(RESET)"
	@if [ -z "$(RUN_ARGS)" ]; then \
		echo "$(BOLD)$(YELLOW)Usage: make run-repair RUN_ARGS=\"book.epub --no-backup --aggressive\"$(RESET)"; \
		exit 1; \
	fi
	$(GO) run $(CMD_DIR) repair $(RUN_ARGS)

.PHONY: run-batch-validate
run-batch-validate: ## Run batch validation (usage: make run-batch-validate RUN_ARGS="./books --jobs 8")
	@echo "$(BOLD)$(RED)Running batch validation...$(RESET)"
	@if [ -z "$(RUN_ARGS)" ]; then \
		echo "$(BOLD)$(YELLOW)Usage: make run-batch-validate RUN_ARGS=\"./books --jobs 8\"$(RESET)"; \
		exit 1; \
	fi
	$(GO) run $(CMD_DIR) batch validate $(RUN_ARGS)

.PHONY: run-batch-repair
run-batch-repair: ## Run batch repair (usage: make run-batch-repair RUN_ARGS="./books --no-backup --aggressive --jobs 8")
	@echo "$(BOLD)$(RED)Running batch repair...$(RESET)"
	@if [ -z "$(RUN_ARGS)" ]; then \
		echo "$(BOLD)$(YELLOW)Usage: make run-batch-repair RUN_ARGS=\"./books --no-backup --aggressive --jobs 8\"$(RESET)"; \
		exit 1; \
	fi
	$(GO) run $(CMD_DIR) batch repair $(RUN_ARGS)

.PHONY: fixtures-repair
fixtures-repair: ## Generate repair fixtures (usage: make fixtures-repair [FIXTURES_REPAIR=...])
	@echo "$(BOLD)$(CYAN)Generating repair fixtures...$(RESET)"
	@chmod +x scripts/generate-repair-fixtures.sh
	@OUT_DIR="$(FIXTURES_REPAIR)" ./scripts/generate-repair-fixtures.sh

.PHONY: check
check: build test lint vet ## Run build, tests, lint, and vet
	@echo "$(BOLD)$(GREEN)✓ Checks complete$(RESET)"

.PHONY: docs
docs: docs-links docs-lint docs-spell ## Validate documentation
	@echo "$(BOLD)$(GREEN)✓ Documentation checks complete$(RESET)"

.PHONY: docs-links
docs-links: ## Validate local markdown links
	@echo "$(BOLD)$(BLUE)Checking markdown links...$(RESET)"
	@python3 scripts/check-docs-links.py

.PHONY: docs-lint
docs-lint: ## Lint markdown files (requires markdownlint-cli2)
	@echo "$(BOLD)$(BLUE)Linting markdown...$(RESET)"
	@if command -v markdownlint-cli2 > /dev/null 2>&1; then \
		markdownlint-cli2 "**/*.md" --config .markdownlint.yaml --fix; \
		printf "%b\n" "$(BOLD)$(GREEN)✓ Markdown lint complete$(RESET)"; \
	else \
		printf "%b\n" "$(BOLD)$(RED)✗ markdownlint-cli2 not installed. Install with: npm i -g markdownlint-cli2$(RESET)"; \
		exit 1; \
	fi

.PHONY: docs-spell
docs-spell: ## Spellcheck docs (requires codespell)
	@echo "$(BOLD)$(BLUE)Spellchecking docs...$(RESET)"
	@if command -v codespell > /dev/null 2>&1; then \
		codespell --check-filenames --check-hidden --config .codespellrc .; \
		printf "%b\n" "$(BOLD)$(GREEN)✓ Spellcheck complete$(RESET)"; \
	else \
		printf "%b\n" "$(BOLD)$(YELLOW)! codespell not installed. Install with: pip install codespell$(RESET)"; \
	fi

# Wiki Operations
.PHONY: wiki-clone
wiki-clone: ## Clone the GitHub wiki repository
	@echo "$(BOLD)$(BLUE)Cloning wiki repository...$(RESET)"
	@chmod +x scripts/wiki-sync.sh
	@./scripts/wiki-sync.sh clone
	@echo "$(BOLD)$(GREEN)✓ Wiki repository cloned$(RESET)"

.PHONY: wiki-sync
wiki-sync: ## Sync documentation to wiki (does not push)
	@echo "$(BOLD)$(BLUE)Syncing documentation to wiki...$(RESET)"
	@chmod +x scripts/wiki-sync.sh
	@./scripts/wiki-sync.sh sync
	@echo "$(BOLD)$(GREEN)✓ Documentation synced to wiki$(RESET)"

.PHONY: wiki-push
wiki-push: ## Commit and push wiki changes
	@echo "$(BOLD)$(BLUE)Pushing wiki changes...$(RESET)"
	@chmod +x scripts/wiki-sync.sh
	@./scripts/wiki-sync.sh push
	@echo "$(BOLD)$(GREEN)✓ Wiki changes pushed$(RESET)"

.PHONY: wiki-pull
wiki-pull: ## Pull latest wiki changes from remote
	@echo "$(BOLD)$(BLUE)Pulling latest wiki changes...$(RESET)"
	@chmod +x scripts/wiki-sync.sh
	@./scripts/wiki-sync.sh pull
	@echo "$(BOLD)$(GREEN)✓ Wiki updated from remote$(RESET)"

.PHONY: wiki-update
wiki-update: ## Full wiki update (clone if needed, sync, and push)
	@echo "$(BOLD)$(BLUE)Performing full wiki update...$(RESET)"
	@chmod +x scripts/wiki-sync.sh
	@./scripts/wiki-sync.sh full
	@echo "$(BOLD)$(GREEN)✓ Wiki fully updated and pushed$(RESET)"

.PHONY: wiki-status
wiki-status: ## Show wiki repository status
	@chmod +x scripts/wiki-sync.sh
	@./scripts/wiki-sync.sh status

.PHONY: wiki-clean
wiki-clean: ## Remove wiki directory
	@echo "$(BOLD)$(RED)Removing wiki directory...$(RESET)"
	@chmod +x scripts/wiki-sync.sh
	@./scripts/wiki-sync.sh clean
	@echo "$(BOLD)$(GREEN)✓ Wiki directory removed$(RESET)"

.DEFAULT_GOAL := help
