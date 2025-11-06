# Proton Integration Package

This package provides integrations with various Proton services for OpenPasswd.

## Structure

```
pkg/proton/
├── proton.go           # Main package documentation
├── pass/               # Proton Pass integration
│   └── importer.go    # Password import from Proton Pass
├── mail/              # [Future] Proton Mail integration
├── drive/             # [Future] Proton Drive integration
└── vpn/               # [Future] Proton VPN integration
```

## Current Features

### Proton Pass (`pkg/proton/pass`)

The Proton Pass importer allows you to import your passwords from Proton Pass into OpenPasswd.

**Supported Export Formats:**
- **JSON** - Unencrypted JSON export
- **CSV** - Simple CSV export
- **ZIP** - ZIP archive containing JSON or PGP-encrypted data
- **PGP** - PGP-encrypted export (requires `gpg` installed)

**Supported Item Types:**
- Login credentials (with username, password, URL, TOTP)
- Secure notes
- Credit cards (with cardholder name, number, CVV, expiration, PIN)
- Identity information (name, email, phone)
- Custom fields

**Usage Example:**

```go
import (
    "github.com/r2unit/openpasswd/pkg/proton/pass"
    "github.com/r2unit/openpasswd/pkg/models"
)

func main() {
    importer := &pass.Importer{}
    
    // Import from unencrypted export
    passwords, err := importer.Import("/path/to/export.json", "")
    
    // Import from encrypted export
    passwords, err := importer.Import("/path/to/export.zip", "my-passphrase")
    
    // Process imported passwords
    for _, pwd := range passwords {
        // Save to database, etc.
    }
}
```

## How to Export from Proton Pass

1. Open Proton Pass (browser extension or web app)
2. Click the **gear icon** (Settings)
3. Go to the **Export** tab
4. Choose export format:
   - **PGP-encrypted ZIP** (recommended) - Most secure
   - **Unencrypted ZIP** - Easier to work with
   - **CSV** - Simple format

For encrypted exports, you'll need to enter a passphrase when importing.

## Future Integrations

The following integrations are planned for future releases:

### Proton Mail (`pkg/proton/mail`)
- Email encryption/decryption
- Message import/export
- Contact synchronization

### Proton Drive (`pkg/proton/drive`)
- Secure file storage for password backups
- Encrypted attachment storage

### Proton VPN (`pkg/proton/vpn`)
- VPN configuration management
- Connection credential storage

## Requirements

For encrypted PGP imports, you need GPG installed:

**Linux:**
```bash
# Debian/Ubuntu
sudo apt-get install gnupg

# Fedora/RHEL
sudo dnf install gnupg2

# Arch
sudo pacman -S gnupg
```

**macOS:**
```bash
brew install gnupg
```

**Windows:**
Install [GPG4Win](https://www.gpg4win.org/)

## Contributing

When adding new Proton service integrations:

1. Create a new subdirectory under `pkg/proton/` (e.g., `mail/`, `drive/`)
2. Implement the service-specific functionality
3. Follow the existing code structure and naming conventions
4. Add comprehensive documentation
5. Update this README with the new integration

## API Documentation

Since Proton doesn't provide public APIs for many services, this package primarily works with exported data files. For services that do offer APIs, we'll implement proper API clients following Proton's terms of service.

## Security Notes

- All imported passwords are re-encrypted with OpenPasswd's encryption before storage
- Temporary files created during import are securely deleted
- Passphrases are never logged or stored in plain text
- PGP decryption happens in isolated temporary files

## License

This package is part of OpenPasswd and follows the same license.
