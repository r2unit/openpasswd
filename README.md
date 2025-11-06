# OpenPasswd

A secure, terminal-based password manager built with Go. Store and manage your passwords locally with end-to-end encryption.

> [!WARNING]
> Openpasswd is still under development and currently in a pre‑alpha phase. Do not use it in production.

## Installation

### Quick Install

Install directly from GitHub using curl:

```bash
curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
```

This will:
- Clone the repository automatically
- Build the binary
- Install to `/usr/local/bin`
- Create convenient aliases (`openpass` and `pw`)
- Set up shell completions for bash and zsh

### Using Install Scripts

```bash
# Clone the repository
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd

# Run install script
./install.sh
```

### Manual Installation

```bash
# Clone and build
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd
go build -o openpasswd ./cmd/openpasswd

# Move to your PATH (optional)
sudo mv openpasswd /usr/local/bin/
```

## Quick Start

```bash
# Initialize configuration
openpasswd init

# Add a password
openpasswd add

# List passwords
openpasswd list

# Configure settings (MFA, passphrase, etc.)
openpasswd settings set-totp
openpasswd settings set-passphrase

# Check version and updates
openpasswd version --check

# Upgrade to latest version
openpasswd upgrade
```

## Usage

### Commands

- `openpasswd init` - Initialize configuration and database
- `openpasswd add` - Add a new password entry
- `openpasswd list` - List and search passwords
- `openpasswd settings` - Manage settings (passphrase, MFA, etc.)
- `openpasswd version` - Show version information
- `openpasswd upgrade` - Upgrade to the latest version
- `openpasswd help` - Show help message

### Version Commands

```bash
openpasswd version              # Show version number
openpasswd version --verbose    # Show detailed build information
openpasswd version --check      # Check for updates
openpasswd upgrade              # Upgrade to latest version
```

## Supported Password Types

OpenPasswd stores all your credentials locally with AES-256-GCM encryption:

- **Login Credentials** - Username, password, URL, notes
- **Credit Cards** - Number, cardholder, expiry, CVV
- **Secure Notes** - Encrypted text notes
- **Identity Information** - Personal details
- **Generic Passwords** - Simple password storage
- **Custom Fields** - Additional encrypted key-value pairs

All password types support:
- TOTP/2FA codes (coming soon)
- Custom metadata fields
- Timestamps (created/updated)
- Search and filtering

## MFA Support

OpenPasswd supports multiple authentication methods:

- **Master Passphrase** - Simple password protection
- **TOTP (Time-based OTP)** - Google Authenticator, Authy, etc.
- **YubiKey** - Hardware key authentication

Configure MFA:
```bash
openpasswd settings set-passphrase  # Set master passphrase
openpasswd settings set-totp         # Enable TOTP
openpasswd settings set-yubikey      # Enable YubiKey
```

## Development

### Building from Source

Using Make (recommended):
```bash
make build          # Build with version information
make install        # Build and install to /usr/local/bin
make version        # Show version info
make clean          # Clean build artifacts
```

Using Go directly:
```bash
# Simple build
go build -o openpasswd ./cmd/openpasswd

# Build with version information
VERSION="0.1.0"
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

go build -ldflags="-X 'github.com/r2unit/openpasswd/pkg/version.Version=${VERSION}' \
                    -X 'github.com/r2unit/openpasswd/pkg/version.GitCommit=${GIT_COMMIT}' \
                    -X 'github.com/r2unit/openpasswd/pkg/version.BuildDate=${BUILD_DATE}'" \
         -o openpasswd ./cmd/openpasswd
```

### Cross-Compilation

Build for multiple platforms:
```bash
make cross-compile  # Builds for Linux, macOS, and Windows
```

Binaries will be created in the `./dist/` directory.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

r2unit - https://github.com/r2unit
