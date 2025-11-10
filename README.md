# OpenPasswd

> A password manager built for the Terminal

[![CI/CD Pipeline](https://github.com/r2unit/openpasswd/actions/workflows/ci.yml/badge.svg)](https://github.com/r2unit/openpasswd/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/r2unit/openpasswd)](https://github.com/r2unit/openpasswd/releases/latest)

> [!WARNING]
> OpenPasswd is still under development and currently in a pre-alpha phase. Do not use it in production.

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
```

This will automatically download, build, and install OpenPasswd with shell completions.

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/r2unit/openpasswd/releases/latest).

### Build from Source

```bash
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd
make install
```

## Quick Start

```bash
# Initialize configuration
openpasswd init

# Add a password
openpasswd add

# List passwords
openpasswd list

# Configure MFA
openpasswd settings set-totp
```

## Documentation

### Commands

- `openpasswd init` - Initialize configuration and database
- `openpasswd add` - Add a new password entry
- `openpasswd list` - List and search passwords
- `openpasswd settings` - Manage settings (passphrase, MFA, etc.)
- `openpasswd version` - Show version information
- `openpasswd upgrade` - Upgrade to the latest version

### Supported Password Types

- **Login Credentials** - Username, password, URL, notes
- **Credit Cards** - Number, cardholder, expiry, CVV
- **Secure Notes** - Encrypted text notes
- **Identity Information** - Personal details
- **Custom Fields** - Additional encrypted key-value pairs

### MFA Support

- **Master Passphrase** - Simple password protection
- **TOTP (Time-based OTP)** - Google Authenticator, Authy, etc.
- **YubiKey** - Hardware key authentication (coming soon)

Configure MFA:
```bash
openpasswd settings set-passphrase  # Set master passphrase
openpasswd settings set-totp         # Enable TOTP
```

### Security

- **AES-256-GCM encryption** for all stored data
- **Argon2id** for key derivation
- **BLAKE2b** for integrity verification
- **Local storage only** - your data never leaves your device
- **Zero-knowledge architecture** - no cloud sync, no telemetry

For more details, see the [Security Architecture](docs/security.md) (coming soon).

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass (`go test ./...`)
- Code is properly formatted (`go fmt ./...`)
- Linter checks pass (`golangci-lint run`)

## FAQ

### Why another password manager?

OpenPasswd is designed for developers and terminal enthusiasts who prefer:
- **Command-line interface** over GUI applications
- **Local storage** over cloud synchronization
- **Full control** over their security setup
- **Open source** transparency and auditability
- **Zero dependencies** on third-party services

### How is this different from pass/gopass?

While inspired by the Unix philosophy, OpenPasswd offers:
- **Built-in encryption** without requiring GPG
- **Structured data** with support for multiple password types
- **Interactive TUI** for better user experience
- **Cross-platform** support with consistent behavior
- **Self-contained** single binary with no external dependencies

### Is it secure?

OpenPasswd uses industry-standard cryptography:
- AES-256-GCM for encryption
- Argon2id for password hashing
- BLAKE2b for integrity checks

However, as the project is in **pre-alpha**, it has not undergone a formal security audit. Use at your own risk.

### Can I sync my passwords across devices?

Currently, OpenPasswd is designed for local storage only. Cloud sync is not planned to maintain zero-knowledge architecture and maximum security.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Author

[@r2unit](https://github.com/r2unit)
