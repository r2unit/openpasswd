.PHONY: build install clean test version release

# Version information
VERSION := $(shell grep 'Version = ' pkg/version/version.go | sed 's/.*"\(.*\)".*/\1/')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# Build flags
LDFLAGS := -ldflags="-X 'github.com/r2unit/openpasswd/pkg/version.Version=$(VERSION)' \
                      -X 'github.com/r2unit/openpasswd/pkg/version.GitCommit=$(GIT_COMMIT)' \
                      -X 'github.com/r2unit/openpasswd/pkg/version.BuildDate=$(BUILD_DATE)'"

# Binary name
BINARY := openpasswd

# Build the binary
build:
	@echo "Building OpenPasswd v$(VERSION) (commit: $(GIT_COMMIT))..."
	@go build $(LDFLAGS) -o $(BINARY) ./cmd/openpasswd
	@echo "✓ Build complete: ./$(BINARY)"

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	@sudo cp $(BINARY) /usr/local/bin/$(BINARY)
	@sudo chmod +x /usr/local/bin/$(BINARY)
	@echo "✓ Installed to /usr/local/bin/$(BINARY)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY)
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Show version
version:
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Create a release build (optimized)
release:
	@echo "Building release version v$(VERSION)..."
	@go build $(LDFLAGS) -trimpath -o $(BINARY) ./cmd/openpasswd
	@echo "✓ Release build complete: ./$(BINARY)"

# Development build (with race detector)
dev:
	@echo "Building development version..."
	@go build -race $(LDFLAGS) -o $(BINARY) ./cmd/openpasswd
	@echo "✓ Development build complete: ./$(BINARY)"

# Cross-compile for multiple platforms
cross-compile:
	@echo "Cross-compiling for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/openpasswd
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 ./cmd/openpasswd
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/openpasswd
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/openpasswd
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/openpasswd
	@echo "✓ Cross-compilation complete. Binaries in ./dist/"

# Help
help:
	@echo "OpenPasswd Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build the binary with version info"
	@echo "  install        Build and install to /usr/local/bin"
	@echo "  clean          Remove build artifacts"
	@echo "  test           Run tests"
	@echo "  version        Show version information"
	@echo "  release        Build optimized release binary"
	@echo "  dev            Build with race detector"
	@echo "  cross-compile  Build for multiple platforms"
	@echo "  help           Show this help message"
