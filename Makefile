# Makefile for license-audit
# Provides manual and semantic release functionality

.PHONY: help build test clean install release-patch release-minor release-major tag-current current-version next-version

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=license-audit

# Get the current version from git tags, default to v0.0.0 if no tags exist
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Strip 'v' prefix for version calculations
VERSION_NUMBER := $(shell echo $(CURRENT_VERSION) | sed 's/^v//')

# Extract major, minor, patch numbers
MAJOR := $(shell echo $(VERSION_NUMBER) | cut -d. -f1)
MINOR := $(shell echo $(VERSION_NUMBER) | cut -d. -f2)
PATCH := $(shell echo $(VERSION_NUMBER) | cut -d. -f3)

# Calculate next versions
NEXT_PATCH := v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH) + 1)))
NEXT_MINOR := v$(MAJOR).$(shell echo $$(($(MINOR) + 1))).0
NEXT_MAJOR := v$(shell echo $$(($(MAJOR) + 1))).0.0

# Build flags
BUILD_FLAGS := -ldflags "-X main.version=$(CURRENT_VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

help: ## Display this help message
	@echo "License Audit Makefile"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage report
	@echo "Coverage report:"
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(BUILD_FLAGS) .

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1)
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

current-version: ## Show current version
	@echo "Current version: $(CURRENT_VERSION)"

next-version: ## Show next possible versions
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Next patch:      $(NEXT_PATCH)"
	@echo "Next minor:      $(NEXT_MINOR)"
	@echo "Next major:      $(NEXT_MAJOR)"

tag-current: ## Create a git tag for the current version (for initial setup)
	@if [ "$(CURRENT_VERSION)" = "v0.0.0" ]; then \
		echo "Creating initial version tag v0.1.0..."; \
		git tag -a v0.1.0 -m "Initial release v0.1.0"; \
		echo "Tagged as v0.1.0"; \
	else \
		echo "Current version already tagged: $(CURRENT_VERSION)"; \
	fi

validate-clean: ## Ensure working directory is clean
	@echo "Checking if working directory is clean..."
	@if ! git diff-index --quiet HEAD --; then \
		echo "Error: Working directory is not clean. Please commit or stash your changes."; \
		exit 1; \
	fi
	@echo "Working directory is clean."

validate-main: ## Ensure we're on main branch
	@echo "Checking if on main branch..."
	@current_branch=$$(git branch --show-current); \
	if [ "$$current_branch" != "main" ]; then \
		echo "Error: Not on main branch (currently on $$current_branch). Please switch to main branch."; \
		exit 1; \
	fi
	@echo "On main branch."

pre-release: validate-main validate-clean check ## Pre-release validation
	@echo "Pre-release validation completed successfully."

release-patch: pre-release ## Create a patch release (x.y.Z+1)
	@echo "Creating patch release: $(CURRENT_VERSION) -> $(NEXT_PATCH)"
	@echo "Changelog entry for $(NEXT_PATCH):"
	@echo "- Bug fixes and small improvements"
	@read -p "Continue with patch release $(NEXT_PATCH)? [y/N]: " confirm && [ "$$confirm" = "y" ]
	git tag -a $(NEXT_PATCH) -m "Release $(NEXT_PATCH)"
	@echo ""
	@echo "✅ Patch release $(NEXT_PATCH) created successfully!"
	@echo "To push the tag: git push origin $(NEXT_PATCH)"

release-minor: pre-release ## Create a minor release (x.Y+1.0)
	@echo "Creating minor release: $(CURRENT_VERSION) -> $(NEXT_MINOR)"
	@echo "Changelog entry for $(NEXT_MINOR):"
	@echo "- New features and improvements"
	@echo "- Backward compatible changes"
	@read -p "Continue with minor release $(NEXT_MINOR)? [y/N]: " confirm && [ "$$confirm" = "y" ]
	git tag -a $(NEXT_MINOR) -m "Release $(NEXT_MINOR)"
	@echo ""
	@echo "✅ Minor release $(NEXT_MINOR) created successfully!"
	@echo "To push the tag: git push origin $(NEXT_MINOR)"

release-major: pre-release ## Create a major release (X+1.0.0)
	@echo "Creating major release: $(CURRENT_VERSION) -> $(NEXT_MAJOR)"
	@echo "⚠️  WARNING: This is a MAJOR release with potential breaking changes!"
	@echo "Changelog entry for $(NEXT_MAJOR):"
	@echo "- Breaking changes"
	@echo "- Major new features"
	@echo "- API changes"
	@read -p "Continue with major release $(NEXT_MAJOR)? [y/N]: " confirm && [ "$$confirm" = "y" ]
	git tag -a $(NEXT_MAJOR) -m "Release $(NEXT_MAJOR)"
	@echo ""
	@echo "✅ Major release $(NEXT_MAJOR) created successfully!"
	@echo "To push the tag: git push origin $(NEXT_MAJOR)"

release-info: ## Show release information and commands
	@echo "Release Management Commands:"
	@echo ""
	@echo "Current version: $(CURRENT_VERSION)"
	@echo ""
	@echo "Available release commands:"
	@echo "  make release-patch  - Create patch release: $(NEXT_PATCH)"
	@echo "  make release-minor  - Create minor release: $(NEXT_MINOR)"
	@echo "  make release-major  - Create major release: $(NEXT_MAJOR)"
	@echo ""
	@echo "Semantic Versioning:"
	@echo "  PATCH (x.y.Z) - Bug fixes, no breaking changes"
	@echo "  MINOR (x.Y.z) - New features, backward compatible"
	@echo "  MAJOR (X.y.z) - Breaking changes"
	@echo ""
	@echo "After creating a release tag:"
	@echo "  git push origin <tag>  - Push the tag to trigger CI/CD"

# Development commands
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	go mod download
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment ready!"

mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

mod-update: ## Update go modules
	@echo "Updating go modules..."
	go get -u ./...
	go mod tidy

# Quick development workflow
quick-check: fmt vet ## Quick check (format and vet only)
	@echo "Quick checks completed."