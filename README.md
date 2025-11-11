# OpenPasswd

> A secure, terminal-based password manager built for developers

[![CI/CD Pipeline](https://github.com/r2unit/openpasswd/actions/workflows/ci.yml/badge.svg)](https://github.com/r2unit/openpasswd/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/r2unit/openpasswd)](https://github.com/r2unit/openpasswd/releases/latest)

> [!WARNING]
> OpenPasswd is still under development and currently in a pre-alpha phase. Do not use it in production.

## Overview

OpenPasswd is a command-line password manager that keeps your passwords safe on your own device. No cloud sync, no third-party servers, just you and your encrypted data. Built for developers and terminal enthusiasts who value privacy and control.

**Why OpenPasswd?**

- **Your data stays local** - Everything lives on your device, encrypted with your master passphrase
- **Strong security** - AES-256-GCM encryption, Argon2id key derivation, BLAKE2b integrity checks
- **Terminal native** - Beautiful TUI (Terminal User Interface) that feels right at home in your workflow
- **Zero dependencies** - Single binary, no external requirements
- **Open source** - Fully transparent code you can audit yourself

## Features

### Password Management
- **Multiple password types**: Login credentials, credit cards, secure notes, identity info
- **Custom fields**: Add any extra encrypted data you need
- **Rich metadata**: URLs, notes, timestamps
- **Fuzzy search**: Find passwords fast by name, username, or URL
- **Secure clipboard**: Copy passwords with automatic clearing

### Security
- **AES-256-GCM encryption** for all stored data
- **Argon2id key derivation** (current default) or PBKDF2-HMAC-SHA256 with 600k iterations
- **BLAKE2b integrity verification** to detect tampering
- **Recovery keys** - Encrypted backup if you forget your passphrase
- **No plaintext storage** - Master passphrase never saved to disk
- **Secure memory handling** - Minimal exposure of sensitive data

### Multi-Factor Authentication
- **Master passphrase** (required)
- **TOTP** - Works with Google Authenticator, Authy, and other authenticator apps
- **YubiKey** - Hardware key support (coming soon)

### Import/Export
- **Proton Pass** - Import from JSON, CSV, or encrypted exports
- **More coming** - Bitwarden, 1Password, LastPass, KeePass (planned)

### User Experience
- **Interactive TUI** - Navigate with keyboard shortcuts (vim-style supported)
- **Color customization** - Personalize your interface
- **Auto-updates** - Check for new versions and upgrade with one command
- **Shell completions** - Bash and Zsh support included

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
```

This automatically downloads, builds, and installs OpenPasswd with shell completions. Creates handy aliases: `openpass` and `pw`.

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/r2unit/openpasswd/releases/latest).

### Build from Source

```bash
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd
make install
```

**Requirements:**
- Go 1.21 or later
- Git

## Quick Start

### First-Time Setup

```bash
# Initialize OpenPasswd (creates config and master passphrase)
openpasswd init

# Add your first password
openpasswd add

# List all passwords
openpasswd list
```

During `init`, you'll create a master passphrase and receive a 24-word recovery key. **Write down your recovery key** and store it somewhere safe - it's your backup if you forget your passphrase.

### Basic Usage

```bash
# Add a new password
openpasswd add                    # Interactive mode
openpasswd add login              # Add login credentials
openpasswd add card               # Add credit card
openpasswd add note               # Add secure note

# List and search passwords
openpasswd list                   # Opens interactive TUI

# Configure MFA
openpasswd settings set-totp      # Enable TOTP authentication
openpasswd settings remove-totp   # Disable TOTP

# Version management
openpasswd version                # Show version
openpasswd version --check        # Check for updates
openpasswd upgrade                # Upgrade to latest version

# Migration
openpasswd migrate upgrade-kdf    # Upgrade to stronger encryption (600k iterations)
```

### Keyboard Shortcuts (in TUI)

- `↑/↓` or `k/j` - Navigate up/down
- `Enter` - Select/view password
- `/` - Search
- `c` - Copy password to clipboard
- `e` - Edit entry
- `d` - Delete entry
- `Esc` - Go back
- `:q` or `Ctrl+C` - Quit

## Documentation

Comprehensive documentation is available in the [docs/](docs/) directory:

- **[User Guide](docs/USER_GUIDE.md)** - Complete guide for using OpenPasswd
- **[Security Architecture](docs/SECURITY.md)** - Detailed security information
- **[CLI Reference](docs/CLI_REFERENCE.md)** - All commands and options
- **[Configuration Guide](docs/CONFIGURATION.md)** - Customize OpenPasswd
- **[Import Guide](docs/IMPORT_GUIDE.md)** - Import from other password managers
- **[Developer Guide](docs/DEVELOPER_GUIDE.md)** - Contributing and development
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

## Architecture

OpenPasswd uses a simple, secure architecture:

```
~/.config/openpasswd/
├── passwords.db          # Encrypted password database (JSON)
├── salt                  # Encryption salt (base64)
├── kdf_version           # KDF version number
├── recovery_key          # Encrypted recovery key
├── recovery_hash         # Recovery key verification hash
├── totp_secret           # TOTP secret (if enabled)
├── config.toml           # User preferences (colors, keybindings)
└── disable_version_check # Flag to disable auto-update checks
```

All sensitive data is encrypted with AES-256-GCM using a key derived from your master passphrase via Argon2id or PBKDF2-HMAC-SHA256.

## Security

### Cryptography

- **Encryption**: AES-256-GCM (Galois/Counter Mode)
- **Key Derivation**: 
  - Argon2id (default) - 3 iterations, 64 MiB memory, 4 parallel lanes
  - PBKDF2-HMAC-SHA256 (legacy) - 600,000 iterations (OWASP recommended)
- **Integrity**: BLAKE2b-256 hashing
- **Random Generation**: `crypto/rand` from Go standard library

### Security Features

- **Zero-knowledge architecture** - Your data never leaves your device
- **No cloud sync** - Eliminates remote attack vectors
- **No telemetry** - We don't collect any data
- **Secure defaults** - Strong encryption out of the box
- **Recovery mechanism** - Encrypted recovery key for passphrase loss
- **Database integrity checks** - Detect corruption or tampering

### Security Notice

OpenPasswd is in **pre-alpha** and hasn't undergone a formal security audit. While we use industry-standard cryptography, use at your own risk. For production use, wait for a stable release and security audit.

### Best Practices

1. **Use a strong master passphrase** - At least 16 characters, mix of letters, numbers, symbols
2. **Store your recovery key safely** - Write it down, keep it offline
3. **Enable TOTP** - Add an extra layer of security
4. **Keep backups** - Regularly backup `~/.config/openpasswd/`
5. **Keep OpenPasswd updated** - Run `openpasswd upgrade` regularly

## Configuration

OpenPasswd stores configuration in `~/.config/openpasswd/config.toml`:

```toml
[keybindings]
quit = ":q"
quit_alt = "ctrl+c"
back = "esc"
up = "up"
up_alt = "k"
down = "down"
down_alt = "j"
select = "enter"

[colors]
# Color customization (coming soon)
```

## Importing from Other Password Managers

OpenPasswd can import passwords from:

- **Proton Pass** - JSON, CSV, or encrypted ZIP exports
- **More coming** - Bitwarden, 1Password, LastPass, KeePass (planned)

See the [Import Guide](docs/IMPORT_GUIDE.md) for detailed instructions.

## Contributing

We welcome contributions! Here's how to get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Format code (`go fmt ./...`)
6. Run linter (`golangci-lint run`)
7. Commit (`git commit -m 'feat: add amazing feature'`)
8. Push (`git push origin feature/amazing-feature`)
9. Open a Pull Request

See the [Developer Guide](docs/DEVELOPER_GUIDE.md) for more details.

### Development Setup

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

OpenPasswd uses industry-standard cryptography (AES-256-GCM, Argon2id, BLAKE2b). However, as the project is in **pre-alpha**, it hasn't undergone a formal security audit. Use at your own risk.

### Can I sync my passwords across devices?

Currently, OpenPasswd is designed for local storage only. You can manually copy the `~/.config/openpasswd/` directory to sync, but there's no built-in sync feature. This is intentional to maintain zero-knowledge architecture and maximum security.

### What if I forget my master passphrase?

Use your 24-word recovery key to regain access. This is why it's critical to write it down and store it safely during initial setup.

### Can I change my master passphrase?

Not yet, but this feature is planned. For now, you'd need to export your passwords, reinitialize OpenPasswd, and re-import them.

### Does it work on Windows?

Yes! OpenPasswd supports Linux, macOS, and Windows. The terminal experience is best on Unix-like systems, but it works on Windows PowerShell and Command Prompt too.

## Roadmap

### v0.1.0 (Current - Pre-Alpha)
- [x] Basic password management (add, list, delete)
- [x] AES-256-GCM encryption
- [x] PBKDF2 key derivation
- [x] Interactive TUI
- [x] TOTP support
- [x] Recovery keys
- [x] Proton Pass import

### v0.2.0 (Planned)
- [ ] Argon2id as default KDF
- [ ] Password generation
- [ ] Password strength meter
- [ ] Breach detection (Have I Been Pwned integration)
- [ ] Bitwarden import
- [ ] 1Password import

### v0.3.0 (Planned)
- [ ] YubiKey support
- [ ] Passphrase change
- [ ] Export functionality
- [ ] Database backup/restore
- [ ] Audit logging

### v1.0.0 (Future)
- [ ] Security audit
- [ ] Stable API
- [ ] Plugin system
- [ ] Browser extension (optional)

## License

OpenPasswd is released under the MIT License. See the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2025 r2unit

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
```

## Support

- **Issues**: [GitHub Issues](https://github.com/r2unit/openpasswd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/r2unit/openpasswd/discussions)
- **Author**: [@r2unit](https://github.com/r2unit)

## Acknowledgments

OpenPasswd is built with these excellent open-source libraries:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- Go standard library - Cryptography and core functionality

Special thanks to the password manager community for inspiration and best practices.

---

**Made with ❤️ for the terminal**
