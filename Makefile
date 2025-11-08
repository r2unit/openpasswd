.PHONY: build build-client build-server install clean test version release

# Version information
VERSION := $(shell grep 'Version = ' pkg/version/version.go | sed 's/.*"\(.*\)".*/\1/')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# Build flags
LDFLAGS := -ldflags="-X 'github.com/r2unit/openpasswd/pkg/version.Version=$(VERSION)' \
                      -X 'github.com/r2unit/openpasswd/pkg/version.GitCommit=$(GIT_COMMIT)' \
                      -X 'github.com/r2unit/openpasswd/pkg/version.BuildDate=$(BUILD_DATE)'"

# Binary names
CLIENT_BINARY := openpasswd
SERVER_BINARY := openpasswd-server

# Build both client and server
build: build-client build-server

# Build the client binary
build-client:
	@echo "Building OpenPasswd Client v$(VERSION) (commit: $(GIT_COMMIT))..."
	@go build $(LDFLAGS) -o $(CLIENT_BINARY) ./cmd/client
	@echo "✓ Client build complete: ./$(CLIENT_BINARY)"

# Build the server binary
build-server:
	@echo "Building OpenPasswd Server v$(VERSION) (commit: $(GIT_COMMIT))..."
	@go build $(LDFLAGS) -o $(SERVER_BINARY) ./cmd/server
	@echo "✓ Server build complete: ./$(SERVER_BINARY)"

# Install to /usr/local/bin
install: build
	@echo "Installing binaries to /usr/local/bin..."
	@sudo cp $(CLIENT_BINARY) /usr/local/bin/$(CLIENT_BINARY)
	@sudo chmod +x /usr/local/bin/$(CLIENT_BINARY)
	@sudo cp $(SERVER_BINARY) /usr/local/bin/$(SERVER_BINARY)
	@sudo chmod +x /usr/local/bin/$(SERVER_BINARY)
	@echo "✓ Installed $(CLIENT_BINARY) and $(SERVER_BINARY) to /usr/local/bin/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(CLIENT_BINARY) $(SERVER_BINARY)
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
	@go build $(LDFLAGS) -trimpath -o $(CLIENT_BINARY) ./cmd/client
	@go build $(LDFLAGS) -trimpath -o $(SERVER_BINARY) ./cmd/server
	@echo "✓ Release build complete: ./$(CLIENT_BINARY) and ./$(SERVER_BINARY)"

# Development build (with race detector)
dev:
	@echo "Building development version..."
	@go build -race $(LDFLAGS) -o $(CLIENT_BINARY) ./cmd/client
	@go build -race $(LDFLAGS) -o $(SERVER_BINARY) ./cmd/server
	@echo "✓ Development build complete: ./$(CLIENT_BINARY) and ./$(SERVER_BINARY)"

# Cross-compile for multiple platforms
cross-compile:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(CLIENT_BINARY)-linux-amd64 ./cmd/client
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(CLIENT_BINARY)-linux-arm64 ./cmd/client
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(CLIENT_BINARY)-darwin-amd64 ./cmd/client
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(CLIENT_BINARY)-darwin-arm64 ./cmd/client
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(CLIENT_BINARY)-windows-amd64.exe ./cmd/client
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(SERVER_BINARY)-linux-amd64 ./cmd/server
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(SERVER_BINARY)-linux-arm64 ./cmd/server
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(SERVER_BINARY)-darwin-amd64 ./cmd/server
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(SERVER_BINARY)-darwin-arm64 ./cmd/server
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(SERVER_BINARY)-windows-amd64.exe ./cmd/server
	@echo "✓ Cross-compilation complete. Binaries in ./dist/"

# Help
help:
	@echo "OpenPasswd Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build both client and server binaries"
	@echo "  build-client   Build only the client binary"
	@echo "  build-server   Build only the server binary"
	@echo "  install        Build and install both to /usr/local/bin"
	@echo "  clean          Remove build artifacts"
	@echo "  test           Run tests"
	@echo "  version        Show version information"
	@echo "  release        Build optimized release binaries"
	@echo "  dev            Build with race detector"
	@echo "  cross-compile  Build for multiple platforms"
	@echo "  help           Show this help message"
