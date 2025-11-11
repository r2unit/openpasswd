# OpenPasswd Documentation

Welcome to the OpenPasswd documentation! OpenPasswd is a secure, terminal-based password manager built for developers and terminal enthusiasts who value privacy, security, and simplicity.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Quick Start](#quick-start)
- [Documentation](#documentation)
- [Security](#security)
- [Contributing](#contributing)

## Overview

OpenPasswd is a command-line password manager that stores all your sensitive data locally with strong encryption. Unlike cloud-based password managers, OpenPasswd follows a zero-knowledge architecture where your data never leaves your device.

### Key Principles

- **Privacy First**: All data is stored locally on your device
- **Strong Encryption**: Industry-standard cryptography (AES-256-GCM, Argon2id/PBKDF2)
- **Zero Dependencies**: Self-contained single binary with no external requirements
- **Terminal Native**: Beautiful TUI (Terminal User Interface) built with Bubbletea
- **Open Source**: Fully transparent and auditable code

## Features

### Password Storage

- **Multiple Password Types**: Login credentials, credit cards, secure notes, identity information
- **Custom Fields**: Add additional encrypted key-value pairs to any entry
- **Rich Metadata**: URLs, notes, timestamps, and more

### Security Features

- **AES-256-GCM Encryption**: Military-grade encryption for all stored data
- **Argon2id Key Derivation**: Current default KDF with configurable iterations
- **PBKDF2 Support**: Legacy support with 600,000 iterations (OWASP recommended)
- **BLAKE2b Integrity Checks**: Detect tampering and corruption
- **Secure Memory Handling**: Minimal exposure of sensitive data in memory
- **Recovery Keys**: Encrypted recovery mechanism if you forget your passphrase

### Multi-Factor Authentication

- **Master Passphrase**: Required for accessing your password vault
- **TOTP (Time-based One-Time Password)**: Google Authenticator, Authy, etc.
- **YubiKey Support**: Hardware key authentication (coming soon)

### User Experience

- **Interactive TUI**: Beautiful terminal interface with search, filter, and navigation
- **Fuzzy Search**: Quickly find passwords by name, username, or URL
- **Copy to Clipboard**: Securely copy passwords with automatic clearing
- **Color Customization**: Personalize the interface with custom color schemes

### Import/Export

- **Proton Pass**: Import from Proton Pass (implementation in progress)
- **Other Providers**: Planned support for Bitwarden, 1Password, LastPass, KeePass

## Quick Start

### Installation

```bash
# Quick install (recommended)
curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash

# Or download from releases
# https://github.com/r2unit/openpasswd/releases/latest

# Or build from source
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd
make install
```

### First-Time Setup

```bash
# Initialize OpenPasswd
openpasswd init

# Add your first password
openpasswd add

# List your passwords
openpasswd list
```

### Basic Commands

```bash
# Add a new password
openpasswd add

# List and search passwords
openpasswd list

# Manage settings
openpasswd settings set-totp       # Enable TOTP
openpasswd settings set-yubikey    # Enable YubiKey

# Check version
openpasswd version

# Upgrade to latest version
openpasswd upgrade
```

## Documentation

- **[User Guide](USER_GUIDE.md)**: Comprehensive guide for end users
- **[Architecture](ARCHITECTURE.md)**: System design and technical architecture
- **[Developer Guide](DEVELOPER_GUIDE.md)**: Contributing and development setup
- **[Security](SECURITY.md)**: Security features and best practices
- **[API Reference](API.md)**: Package and module documentation

## Security

OpenPasswd uses industry-standard cryptography:

- **Encryption**: AES-256-GCM (Galois/Counter Mode)
- **Key Derivation**: Argon2id (default) or PBKDF2-HMAC-SHA256 (legacy)
- **Integrity**: BLAKE2b hashing
- **Random Number Generation**: `crypto/rand` from Go standard library

### Security Notice

OpenPasswd is currently in **pre-alpha** and has not undergone a formal security audit. While we use industry-standard cryptography, use at your own risk for production data.

## Contributing

We welcome contributions! Please see our [Developer Guide](DEVELOPER_GUIDE.md) for details on:

- Setting up your development environment
- Code style and conventions
- Testing requirements
- Submitting pull requests

## License

OpenPasswd is released under the MIT License. See the [LICENSE](../LICENSE) file for details.

## Support

- **GitHub Issues**: https://github.com/r2unit/openpasswd/issues
- **Discussions**: https://github.com/r2unit/openpasswd/discussions
- **Author**: [@r2unit](https://github.com/r2unit)
