# OpenPasswd

A secure, terminal-based password manager built with Go. Store and manage your passwords locally with end-to-end encryption.

> [!WARNING]
> Openpasswd is still under development and currently in a pre‑alpha phase. Do not use it in production.

## Installation

### Quick Install (One-Liner)

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

### Using Install Script (Alternative)

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
```

## Usage

### Commands

- `openpasswd init` - Initialize configuration and database
- `openpasswd add` - Add a new password entry
- `openpasswd list` - List and search passwords
- `openpasswd settings` - Manage settings (passphrase, MFA, etc.)
- `openpasswd help` - Show help message

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

r2unit - https://github.com/r2unit
