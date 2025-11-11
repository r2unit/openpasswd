# Developer Guide

Complete guide for developers who want to contribute to OpenPasswd or understand its internals.

## Table of Contents

- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Setup](#development-setup)
- [Building and Testing](#building-and-testing)
- [Code Style](#code-style)
- [Contributing](#contributing)
- [Architecture](#architecture)
- [Adding Features](#adding-features)
- [Security Considerations](#security-considerations)
- [Release Process](#release-process)

## Getting Started

### Prerequisites

- **Go 1.21 or later** - [Download](https://golang.org/dl/)
- **Git** - Version control
- **Make** - Build automation (optional but recommended)
- **golangci-lint** - Code linting (optional)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Install locally
make install
```

## Project Structure

```
openpasswd/
├── cmd/
│   ├── client/          # Client CLI application
│   │   └── main.go      # Entry point, command handling
│   └── server/          # Server (future sync functionality)
│       └── main.go      # Server entry point
├── pkg/
│   ├── auth/            # Authentication providers
│   │   ├── auth.go      # Auth interface
│   │   └── providers.go # Provider implementations
│   ├── client/          # Client library
│   │   └── client.go    # HTTP client for server
│   ├── config/          # Configuration management
│   │   ├── config.go    # Config loading/saving
│   │   └── colors.go    # Color configuration
│   ├── crypto/          # Cryptography
│   │   ├── crypto.go    # AES-256-GCM encryption
│   │   ├── argon2.go    # Argon2id implementation
│   │   ├── blake2b.go   # BLAKE2b hashing
│   │   ├── integrity.go # Database integrity checks
│   │   ├── recovery.go  # Recovery key generation
│   │   ├── securemem.go # Secure memory handling
│   │   └── version.go   # KDF versioning
│   ├── database/        # Database operations
│   │   ├── database.go  # CRUD operations
│   │   └── integrity.go # Integrity verification
│   ├── mfa/             # Multi-factor authentication
│   │   ├── totp.go      # TOTP implementation
│   │   └── yubikey.go   # YubiKey support
│   ├── models/          # Data models
│   │   └── password.go  # Password struct
│   ├── proton/          # Proton integrations
│   │   ├── proton.go    # Main package
│   │   └── pass/        # Proton Pass importer
│   │       ├── importer.go
│   │       └── provider.go
│   ├── qrcode/          # QR code generation
│   │   └── qrcode.go    # ASCII QR codes
│   ├── server/          # Server implementation
│   │   ├── server.go    # HTTP server
│   │   └── qr_server.go # QR code server
│   ├── sources/         # Import sources
│   │   ├── sources.go   # Importer interface
│   │   └── other_importers.go
│   ├── toml/            # TOML parser
│   │   └── parser.go    # Custom TOML parsing
│   ├── tui/             # Terminal UI
│   │   ├── tui.go       # Main TUI logic
│   │   ├── bubbletea.go # Bubble Tea integration
│   │   ├── list.go      # Password list view
│   │   ├── add.go       # Add password view
│   │   ├── init.go      # Initialization UI
│   │   ├── setup.go     # Setup wizard
│   │   ├── auth_login.go # Authentication UI
│   │   ├── totp.go      # TOTP setup UI
│   │   ├── version.go   # Version check UI
│   │   ├── upgrade.go   # Upgrade UI
│   │   ├── colors.go    # Color schemes
│   │   └── terminal*.go # Terminal utilities
│   └── version/         # Version management
│       └── version.go   # Version info, update checks
├── docs/                # Documentation
│   ├── README.md
│   ├── USER_GUIDE.md
│   ├── SECURITY.md
│   ├── CLI_REFERENCE.md
│   ├── CONFIGURATION.md
│   ├── IMPORT_GUIDE.md
│   ├── TROUBLESHOOTING.md
│   └── DEVELOPER_GUIDE.md (this file)
├── .github/
│   └── workflows/
│       └── ci.yml       # CI/CD pipeline
├── .gitignore
├── .golangci.yml        # Linter configuration
├── Dockerfile           # Docker build
├── go.mod               # Go dependencies
├── go.sum               # Dependency checksums
├── install.sh           # Installation script
├── uninstall.sh         # Uninstallation script
├── Makefile             # Build automation
├── LICENSE              # MIT License
└── README.md            # Main README
```

## Development Setup

### Environment Setup

```bash
# Clone repository
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd

# Set up Go environment
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
```

### IDE Setup

**VS Code:**

Install extensions:
- Go (golang.go)
- Go Test Explorer
- Error Lens

Settings (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "editor.formatOnSave": true,
  "go.formatTool": "goimports"
}
```

**GoLand/IntelliJ:**

- Enable Go modules support
- Configure golangci-lint
- Enable format on save

### Running Locally

```bash
# Build and run
go run ./cmd/client init
go run ./cmd/client list

# Or use make
make build
./openpasswd init
./openpasswd list

# Development build with race detector
make dev
./openpasswd list
```

## Building and Testing

### Building

```bash
# Build client
make build-client

# Build server
make build-server

# Build both
make build

# Release build (optimized)
make release

# Cross-compile for all platforms
make cross-compile
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run specific package tests
go test ./pkg/crypto/

# Run specific test
go test -run TestEncryption ./pkg/crypto/

# Verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
# Run linter
golangci-lint run

# Run linter with auto-fix
golangci-lint run --fix

# Run specific linters
golangci-lint run --enable-only=gosec,govet

# Configuration in .golangci.yml
```

### Formatting

```bash
# Format code
go fmt ./...

# Or use goimports (better)
goimports -w .

# Check formatting
gofmt -l .
```

## Code Style

### Go Style Guide

Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key points:**

1. **Naming:**
   - Use camelCase for unexported names
   - Use PascalCase for exported names
   - Use short, descriptive names
   - Avoid stuttering (e.g., `crypto.Crypto` → `crypto.Encryptor`)

2. **Error handling:**
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to encrypt: %w", err)
   }
   
   // Bad
   if err != nil {
       panic(err)  // Don't panic in library code
   }
   ```

3. **Comments:**
   ```go
   // Encrypt encrypts plaintext using AES-256-GCM.
   // Returns base64-encoded ciphertext or error.
   func (e *Encryptor) Encrypt(plaintext string) (string, error) {
       // ...
   }
   ```

4. **Package organization:**
   - One package per directory
   - Keep packages focused and cohesive
   - Avoid circular dependencies

### Security Code Style

**Sensitive data handling:**

```go
// Good: Clear sensitive data
defer func() {
    for i := range passphrase {
        passphrase[i] = 0
    }
}()

// Good: Use constant-time comparison
if subtle.ConstantTimeCompare(hash1, hash2) == 1 {
    // ...
}

// Bad: Don't log sensitive data
log.Printf("Password: %s", password)  // Never!

// Good: Log safely
log.Printf("Processing password entry: %s", entryName)
```

**Cryptography:**

```go
// Good: Use crypto/rand
nonce := make([]byte, 12)
if _, err := rand.Read(nonce); err != nil {
    return err
}

// Bad: Don't use math/rand for crypto
nonce := make([]byte, 12)
rand.Read(nonce)  // Wrong rand package!

// Good: Check errors from crypto operations
ciphertext, err := gcm.Seal(nonce, nonce, plaintext, nil)
if err != nil {
    return nil, fmt.Errorf("encryption failed: %w", err)
}
```

### Testing Style

```go
func TestEncryption(t *testing.T) {
    tests := []struct {
        name      string
        plaintext string
        wantErr   bool
    }{
        {
            name:      "simple text",
            plaintext: "hello world",
            wantErr:   false,
        },
        {
            name:      "empty string",
            plaintext: "",
            wantErr:   false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

## Contributing

### Contribution Workflow

1. **Fork the repository**
   - Click "Fork" on GitHub
   - Clone your fork

2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make your changes**
   - Write code
   - Add tests
   - Update documentation

4. **Test your changes**
   ```bash
   make test
   golangci-lint run
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

6. **Push to your fork**
   ```bash
   git push origin feature/amazing-feature
   ```

7. **Open a Pull Request**
   - Go to GitHub
   - Click "New Pull Request"
   - Describe your changes

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Adding tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(crypto): add Argon2id support

Implement Argon2id key derivation function as an alternative to PBKDF2.
Provides better resistance to GPU attacks.

Closes #123

fix(tui): correct clipboard clearing on macOS

The clipboard wasn't being cleared properly on macOS due to
platform-specific behavior.

docs(security): update encryption documentation

Add details about Argon2id parameters and security margins.
```

### Pull Request Guidelines

**Before submitting:**
- [ ] Tests pass (`make test`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] Documentation is updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts

**PR description should include:**
- What changes were made
- Why the changes were made
- How to test the changes
- Screenshots (if UI changes)
- Related issues

### Code Review Process

1. **Automated checks:**
   - CI/CD runs tests
   - Linter checks code quality
   - Build verification

2. **Manual review:**
   - Maintainer reviews code
   - Feedback provided
   - Changes requested if needed

3. **Approval and merge:**
   - Once approved, PR is merged
   - Branch is deleted
   - Changes appear in next release

## Architecture

### High-Level Overview

```
┌─────────────┐
│   CLI       │  User interacts with CLI
│  (cmd/)     │
└──────┬──────┘
       │
┌──────▼──────┐
│    TUI      │  Terminal UI (Bubble Tea)
│  (pkg/tui)  │
└──────┬──────┘
       │
┌──────▼──────┐
│  Database   │  Password CRUD operations
│ (pkg/db)    │
└──────┬──────┘
       │
┌──────▼──────┐
│   Crypto    │  Encryption/Decryption
│ (pkg/crypto)│
└─────────────┘
```

### Encryption Flow

```
User Input (Plaintext Password)
    ↓
Master Passphrase + Salt
    ↓
Argon2id/PBKDF2 (Key Derivation)
    ↓
Encryption Key (32 bytes)
    ↓
AES-256-GCM Encryption
    ↓
Ciphertext + Nonce + Auth Tag
    ↓
Base64 Encoding
    ↓
JSON Database Storage
```

### Database Schema

```go
type Password struct {
    ID        int64             // Unique identifier
    Type      PasswordType      // login, card, note, etc.
    Name      string            // Encrypted name
    Username  string            // Encrypted username
    Password  string            // Encrypted password
    URL       string            // Encrypted URL
    Notes     string            // Encrypted notes
    Fields    map[string]string // Encrypted custom fields
    CreatedAt time.Time         // Plaintext timestamp
    UpdatedAt time.Time         // Plaintext timestamp
}
```

### Package Dependencies

```
cmd/client
  └─> pkg/tui
       └─> pkg/database
            └─> pkg/crypto
                 └─> pkg/models
  └─> pkg/config
  └─> pkg/mfa
  └─> pkg/version
```

## Adding Features

### Adding a New Command

1. **Add command handler in `cmd/client/main.go`:**

```go
func handleMyCommand() {
    // Command logic
}

// In main():
case "mycommand":
    handleMyCommand()
```

2. **Update help text:**

```go
func showHelp() {
    help := `...
    openpasswd mycommand         My new command
    ...`
}
```

3. **Add tests:**

```go
func TestMyCommand(t *testing.T) {
    // Test logic
}
```

### Adding a New Password Type

1. **Add type to `pkg/models/password.go`:**

```go
const (
    // ...
    TypeMyType PasswordType = "mytype"
)
```

2. **Add TUI support in `pkg/tui/add.go`:**

```go
case TypeMyType:
    // Handle mytype input
```

3. **Update documentation:**
   - Add to USER_GUIDE.md
   - Update CLI_REFERENCE.md

### Adding a New Importer

1. **Create importer in `pkg/sources/`:**

```go
// mymanager_importer.go
type MyManagerImporter struct{}

func (m *MyManagerImporter) Import(filePath string, passphrase string) ([]*models.Password, error) {
    // Import logic
}

func (m *MyManagerImporter) GetName() string {
    return "My Manager"
}

func (m *MyManagerImporter) SupportsFormat(format string) bool {
    return format == ".json" || format == ".csv"
}
```

2. **Register importer in `pkg/sources/sources.go`:**

```go
const (
    // ...
    SourceMyManager Source = "mymanager"
)

func GetImporter(source Source) Importer {
    switch source {
    // ...
    case SourceMyManager:
        return &MyManagerImporter{}
    }
}
```

3. **Add tests:**

```go
func TestMyManagerImporter(t *testing.T) {
    // Test import logic
}
```

4. **Update documentation:**
   - Add to IMPORT_GUIDE.md
   - Include export instructions

## Security Considerations

### Security Checklist

When adding features:

- [ ] No sensitive data in logs
- [ ] Use crypto/rand for random generation
- [ ] Clear sensitive data from memory
- [ ] Use constant-time comparisons
- [ ] Validate all inputs
- [ ] Handle errors securely
- [ ] No hardcoded secrets
- [ ] Proper file permissions
- [ ] Secure defaults
- [ ] Document security implications

### Common Security Pitfalls

**❌ Don't:**
```go
// Don't log sensitive data
log.Printf("Password: %s", password)

// Don't use math/rand for crypto
key := make([]byte, 32)
rand.Read(key)  // Wrong rand!

// Don't ignore errors
encrypted, _ := Encrypt(plaintext)

// Don't use == for secret comparison
if hash1 == hash2 {  // Timing attack!
```

**✅ Do:**
```go
// Log safely
log.Printf("Processing entry: %s", entryName)

// Use crypto/rand
key := make([]byte, 32)
if _, err := cryptorand.Read(key); err != nil {
    return err
}

// Handle errors
encrypted, err := Encrypt(plaintext)
if err != nil {
    return fmt.Errorf("encryption failed: %w", err)
}

// Use constant-time comparison
if subtle.ConstantTimeCompare(hash1, hash2) == 1 {
```

### Security Testing

```go
func TestEncryptionSecurity(t *testing.T) {
    // Test that encryption is non-deterministic
    plaintext := "test"
    enc1, _ := Encrypt(plaintext)
    enc2, _ := Encrypt(plaintext)
    if enc1 == enc2 {
        t.Error("Encryption is deterministic (bad!)")
    }
    
    // Test that decryption fails with wrong key
    wrongKey := []byte("wrong key")
    _, err := DecryptWithKey(enc1, wrongKey)
    if err == nil {
        t.Error("Decryption succeeded with wrong key (bad!)")
    }
}
```

## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes (v1.0.0 → v2.0.0)
- **MINOR**: New features, backward compatible (v1.0.0 → v1.1.0)
- **PATCH**: Bug fixes (v1.0.0 → v1.0.1)

### Creating a Release

1. **Update version in `pkg/version/version.go`:**

```go
const Version = "0.2.0"
```

2. **Update CHANGELOG.md:**

```markdown
## [0.2.0] - 2025-01-15

### Added
- Password generation feature
- Breach detection with Have I Been Pwned

### Changed
- Improved TUI performance
- Updated dependencies

### Fixed
- Clipboard clearing on macOS
- TOTP time sync issues
```

3. **Commit and tag:**

```bash
git add .
git commit -m "chore: bump version to 0.2.0"
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin main --tags
```

4. **GitHub Actions builds release:**
   - Automatically builds binaries
   - Creates GitHub release
   - Uploads artifacts

5. **Announce release:**
   - GitHub Discussions
   - Update README.md
   - Social media (optional)

## Resources

### Documentation
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)

### Cryptography
- [NIST Guidelines](https://csrc.nist.gov/publications)
- [OWASP Cheat Sheets](https://cheatsheetseries.owasp.org/)
- [Argon2 RFC 9106](https://www.rfc-editor.org/rfc/rfc9106.html)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [gosec](https://github.com/securego/gosec)
- [staticcheck](https://staticcheck.io/)

## Getting Help

- **Questions:** GitHub Discussions
- **Bugs:** GitHub Issues
- **Security:** GitHub Security Advisories
- **Chat:** Coming soon

## License

OpenPasswd is released under the MIT License. See [LICENSE](../LICENSE) for details.

---

Thank you for contributing to OpenPasswd! 🎉
