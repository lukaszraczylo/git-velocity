.PHONY: all build build-spa build-quick install clean test test-coverage lint security dev dev-spa serve help

# Build configuration
BINARY_NAME := git-velocity
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X github.com/lukaszraczylo/git-velocity/pkg/version.Version=$(VERSION) \
           -X github.com/lukaszraczylo/git-velocity/pkg/version.BuildTime=$(BUILD_TIME)

# Directories
WEB_DIR := web
DIST_DIR := $(WEB_DIR)/dist
EMBED_DIR := internal/generator/site/dist

all: build

## Build the Vue SPA
build-spa:
	@echo "Building Vue SPA..."
	@rm -f $(WEB_DIR)/public/data  # Remove dev symlink if exists (breaks vite build)
	@cd $(WEB_DIR) && npm install && npm run build
	@rm -rf $(EMBED_DIR)
	@mkdir -p $(EMBED_DIR)
	@cp -r $(DIST_DIR)/* $(EMBED_DIR)/
	@echo "SPA built and copied to $(EMBED_DIR)"

## Build the Go binary (requires SPA to be built first)
build: build-spa
	@echo "Building Go binary..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/git-velocity
	@echo "Built $(BINARY_NAME)"

## Build without rebuilding SPA (faster for Go-only changes)
build-quick:
	@echo "Building Go binary (quick)..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/git-velocity
	@echo "Built $(BINARY_NAME)"

## Install the binary
install: build
	@go install -ldflags "$(LDFLAGS)" ./cmd/git-velocity

## Run tests
test:
	@echo "Running tests..."
	@go test -race -v ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

## Run security scanner (uses .golangci.yml config)
security:
	@echo "Running security scanner..."
	@golangci-lint run --enable gosec ./...

## Run Vue dev server for frontend development
dev-spa:
	@mkdir -p ./dist/data  # Ensure data dir exists for symlink
	@test -L $(WEB_DIR)/public/data || ln -sf ../../dist/data $(WEB_DIR)/public/data
	@cd $(WEB_DIR) && npm run dev

## Run Go binary with sample config
dev:
	@go run ./cmd/git-velocity analyze --config config.example.yaml --output ./dist

## Serve generated output
serve:
	@rm -f ./dist/index.html
	@rm -rf ./dist/assets
	@cp -r $(EMBED_DIR)/* ./dist/
	@go run ./cmd/git-velocity serve --directory ./dist --port 8080

## Clean build artifacts
clean:
	@rm -rf $(BINARY_NAME) $(DIST_DIR) $(EMBED_DIR) coverage.out coverage.html dist
	@echo "Cleaned build artifacts"

## Show help
help:
	@echo "Git Velocity - Developer Metrics Dashboard"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          Build everything (default)"
	@echo "  build        Build Vue SPA and Go binary"
	@echo "  build-spa    Build only the Vue SPA"
	@echo "  build-quick  Build Go binary without rebuilding SPA"
	@echo "  install      Install binary to GOPATH/bin"
	@echo "  test         Run tests with race detector"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  lint         Run golangci-lint"
	@echo "  security     Run gosec security scanner"
	@echo "  dev-spa      Run Vue dev server"
	@echo "  dev          Run analyzer with sample config"
	@echo "  serve        Serve generated output locally"
	@echo "  clean        Remove build artifacts"
	@echo "  help         Show this help"
