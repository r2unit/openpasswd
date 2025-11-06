# OpenPasswd

A secure, terminal-based password manager built with Go. Store and manage your passwords locally with end-to-end encryption and sync from external password managers.

## Features

- **Terminal UI**: Beautiful interactive terminal interface powered by Bubble Tea
- **End-to-End Encryption**: AES-256-GCM encryption with PBKDF2 key derivation
- **Local Storage**: SQLite database with encrypted passwords
- **Provider Integration**: Connect to Proton Pass and other password managers
- **Import/Export**: Support for multiple password manager formats
- **MFA Support**: TOTP and YubiKey authentication
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Using Install Script (Recommended)

```bash
# Clone the repository
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd

# Run install script
./install.sh
```

This will install `openpasswd` to `/usr/local/bin` and create convenient aliases (`openpass` and `pw`).

### Manual Installation

```bash
# Build the binary
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

# Connect to Proton Pass
openpasswd auth login

# Configure MFA
openpasswd settings set-totp
```

## Usage

### Commands

- `openpasswd init` - Initialize configuration and database
- `openpasswd add` - Add a new password entry
- `openpasswd list` - List and search passwords
- `openpasswd auth login` - Connect to password providers (Proton Pass, etc.)
- `openpasswd import` - Import passwords from files
- `openpasswd settings` - Manage settings (passphrase, MFA, etc.)
- `openpasswd help` - Show help message

### Aliases

You can use these shorter commands:
- `openpass` = `openpasswd` (short alias)
- `pw` = `openpasswd` (ultra-short alias)

## Proton Pass Integration

### Important: No Public API Available

Proton Pass **does not provide a public API** for third-party integrations. The only supported method for integration is via export files.

### How to Sync from Proton Pass

1. **Export from Proton Pass**:
   - Open Proton Pass (browser extension or web app)
   - Go to Settings → Export
   - Choose your export format:
     - **PGP-encrypted ZIP** (recommended, most secure)
     - **Unencrypted ZIP**
     - **CSV**

2. **Import to OpenPasswd**:
   ```bash
   openpasswd auth login
   # Select "Proton Pass" from the provider list
   # Enter the path to your export file
   # Enter passphrase (if encrypted)
   ```

### Supported Export Formats

- **JSON** - Proton Pass JSON export (unencrypted)
- **CSV** - Simple CSV export
- **ZIP** - ZIP archive containing JSON or encrypted data
- **PGP** - PGP-encrypted export (requires `gpg` installed)

### Why No Live API Sync?

Proton Pass uses end-to-end encryption where all encryption happens client-side. They deliberately do not expose a public API to maintain security and privacy. The export feature is the official and recommended way to migrate or backup your passwords.

### Research References

- [proton-pass-common repository](https://github.com/protonpass/proton-pass-common) - Internal library, not a public API
- [Proton WebClients repository](https://github.com/ProtonMail/WebClients) - Official clients source code
- All Proton Pass clients use private, undocumented APIs for internal use only

## Configuration

Configuration is stored in `~/.config/openpasswd/`:

- `salt` - Encryption salt (base64 encoded)
- `passwords.db` - Encrypted password database (SQLite)
- `passphrase` - Master passphrase (optional, encrypted)
- `totp_secret` - TOTP secret for 2FA (optional)
- `yubikey_challenge` - YubiKey challenge (optional)
- `config.toml` - Color configuration

## Security

- **Encryption**: All passwords are encrypted with AES-256-GCM
- **Key Derivation**: PBKDF2 with 100,000 iterations and SHA-256
- **Secure Storage**: Database files stored with 0600 permissions
- **Token-Based Auth**: JWT-like tokens with 24-hour expiration
- **No Third-Party Deps**: Reduces attack surface

## Supported Password Managers

### Currently Supported ✅

- **Proton Pass** - Via export file only (no public API available)
  - Formats: JSON, CSV, ZIP, PGP-encrypted
  - Supports: Logins, Secure Notes, Credit Cards, Identities, TOTP
  - Export from: Settings → Export → Choose format
  - See [Proton Pass Integration](#proton-pass-integration) section below

### Future Integrations 🚧

The following password managers are planned for future implementation:

- **Bitwarden** - Has public API with OAuth support
  - Live sync possible
  - API docs: https://bitwarden.com/help/public-api/
  
- **1Password** - Has CLI and Connect API
  - Official Go SDK available
  - API docs: https://developer.1password.com/
  
- **LastPass** - CSV export only
  - No official API for third-party apps
  
- **KeePass** - File-based, no API
  - Could sync via cloud storage

**Note**: Currently, OpenPasswd focuses on local password storage with Proton Pass import capability. Additional integrations will be added based on demand and API availability.

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

### Project Structure

```
openpasswd/
├── cmd/
│   └── openpasswd/          # Main application
├── pkg/
│   ├── auth/                # Authentication & providers
│   ├── config/              # Configuration management
│   ├── crypto/              # Encryption (AES-256-GCM)
│   ├── database/            # SQLite operations
│   ├── models/              # Data models
│   ├── mfa/                 # TOTP & YubiKey
│   ├── proton/              # Proton services integration
│   │   └── pass/            # Proton Pass provider & importer
│   ├── sources/             # Password manager importers
│   ├── tui/                 # Terminal UI (Bubble Tea)
│   └── qrcode/              # QR code generation
├── install.sh               # Installation script
├── uninstall.sh             # Uninstallation script
├── go.mod
└── README.md
```

### Building

```bash
# Build for current platform
go build -o openpasswd ./cmd/openpasswd

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o openpasswd-linux ./cmd/openpasswd

# Build for macOS  
GOOS=darwin GOARCH=amd64 go build -o openpasswd-macos ./cmd/openpasswd

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o openpasswd.exe ./cmd/openpasswd
```

### Dependencies

Key dependencies:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Modernc SQLite](https://gitlab.com/cznic/sqlite) - Pure Go SQLite
- [TOTP](https://github.com/pquerna/otp) - TOTP/2FA support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

r2unit - https://github.com/r2unit
