# OpenPasswd User Guide

Complete guide to using OpenPasswd for managing your passwords securely.

## Table of Contents

- [Getting Started](#getting-started)
- [Basic Operations](#basic-operations)
- [Password Types](#password-types)
- [Multi-Factor Authentication](#multi-factor-authentication)
- [Search and Navigation](#search-and-navigation)
- [Security Best Practices](#security-best-practices)
- [Backup and Recovery](#backup-and-recovery)
- [Advanced Usage](#advanced-usage)

## Getting Started

### Installation

Choose your preferred installation method:

**Quick Install (Recommended):**
```bash
curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
```

**From Pre-built Binary:**
1. Download from [releases page](https://github.com/r2unit/openpasswd/releases/latest)
2. Extract the archive
3. Move the binary to your PATH: `sudo mv openpasswd /usr/local/bin/`
4. Make it executable: `sudo chmod +x /usr/local/bin/openpasswd`

**Build from Source:**
```bash
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd
make install
```

### First-Time Setup

Run the initialization command:

```bash
openpasswd init
```

You'll be guided through:

1. **Creating a master passphrase** - This encrypts all your passwords. Choose something strong but memorable.
2. **Receiving a recovery key** - A 24-word phrase that can restore access if you forget your passphrase.

**Important:** Write down your recovery key on paper and store it somewhere safe. Don't save it digitally!

Example recovery key format:
```
1. abandon-ability-able-about
2. above-absent-absorb-abstract
3. absurd-abuse-access-accident
...
6. busy-butter-buyer-buzz
```

### Verifying Installation

Check that OpenPasswd is installed correctly:

```bash
openpasswd version
```

You should see version information. If you get "command not found", check your PATH.

## Basic Operations

### Adding Passwords

**Interactive Mode:**
```bash
openpasswd add
```

This opens a menu where you can choose the password type and fill in details.

**Quick Add:**
```bash
openpasswd add login      # Add login credentials
openpasswd add card       # Add credit card
openpasswd add note       # Add secure note
openpasswd add identity   # Add identity info
```

**What You Can Store:**
- Name (required) - A descriptive label like "GitHub" or "Gmail"
- Username/Email
- Password
- URL
- Notes
- Custom fields (any additional data you need)

### Viewing Passwords

```bash
openpasswd list
```

This opens an interactive terminal interface where you can:
- Browse all passwords
- Search by name, username, or URL
- View password details
- Copy passwords to clipboard
- Edit or delete entries

### Searching

In the list view, press `/` to search. Type any part of:
- Password name
- Username
- URL

The list filters in real-time as you type.

### Copying Passwords

In the list view:
1. Navigate to the password you want
2. Press `c` to copy the password to clipboard
3. The password is automatically cleared from clipboard after 45 seconds

### Editing Passwords

In the list view:
1. Navigate to the password
2. Press `e` to edit
3. Modify any fields
4. Save changes

### Deleting Passwords

In the list view:
1. Navigate to the password
2. Press `d` to delete
3. Confirm deletion

**Warning:** Deletion is permanent. Make sure you have backups if you're unsure.

## Password Types

OpenPasswd supports different password types to organize your data:

### Login Credentials

For website and app accounts.

**Fields:**
- Name (e.g., "GitHub Account")
- Username or email
- Password
- URL (e.g., "https://github.com")
- Notes
- Custom fields (e.g., security questions)

**Add:**
```bash
openpasswd add login
```

### Credit Cards

For payment information.

**Fields:**
- Name (e.g., "Visa Ending 1234")
- Cardholder name
- Card number
- CVV/CVC
- Expiration date
- PIN (optional)
- Notes

**Add:**
```bash
openpasswd add card
```

### Secure Notes

For any text you want to keep encrypted.

**Fields:**
- Name (e.g., "WiFi Password")
- Content (the actual note)

**Add:**
```bash
openpasswd add note
```

### Identity Information

For personal details.

**Fields:**
- Name (e.g., "Passport Info")
- Full name
- Email
- Phone number
- Address
- Custom fields

**Add:**
```bash
openpasswd add identity
```

### Custom Fields

All password types support custom fields. Use these for:
- Security questions and answers
- Account numbers
- PINs
- Recovery codes
- Any other encrypted data

## Multi-Factor Authentication

Add extra security layers to your password vault.

### Master Passphrase

Your master passphrase is always required. It's the primary protection for your passwords.

**Tips for a Strong Passphrase:**
- At least 16 characters
- Mix uppercase, lowercase, numbers, symbols
- Use a memorable phrase, not a single word
- Don't reuse passwords from other services

**Example Good Passphrases:**
- `Coffee!Morning@2025#Sunshine`
- `MyDog-Loves-Pizza-1234!`
- `Tr0ub4dor&3-but-longer`

### TOTP (Time-based One-Time Password)

TOTP adds a second factor using authenticator apps like Google Authenticator, Authy, or 1Password.

**Enable TOTP:**
```bash
openpasswd settings set-totp
```

You'll see:
1. A QR code in your terminal
2. A secret key (if you can't scan the QR code)

**Scan with your authenticator app:**
- Open Google Authenticator / Authy / etc.
- Tap "Add account" or "+"
- Scan the QR code
- Enter the 6-digit code to verify

**Using TOTP:**

After enabling, you'll need to enter a 6-digit code from your authenticator app when accessing passwords:

```bash
openpasswd list
Enter master passphrase: ****
Enter 6-digit TOTP code: 123456
```

**Disable TOTP:**
```bash
openpasswd settings remove-totp
```

**Warning:** Make sure you have access to your authenticator app before enabling TOTP. If you lose access, you'll need your recovery key.

### YubiKey (Coming Soon)

Hardware key authentication will be available in a future release.

## Search and Navigation

### Keyboard Shortcuts

**In List View:**
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` - View password details
- `/` - Start search
- `c` - Copy password to clipboard
- `e` - Edit password
- `d` - Delete password
- `Esc` - Clear search / Go back
- `:q` or `Ctrl+C` - Quit

**In Detail View:**
- `c` - Copy password
- `u` - Copy username
- `l` - Copy URL
- `Esc` - Go back
- `:q` or `Ctrl+C` - Quit

### Search Tips

Search is fuzzy and case-insensitive. It matches:
- Password names
- Usernames
- URLs

**Examples:**
- Type `git` to find "GitHub", "GitLab", "Gitea"
- Type `@gmail` to find all Gmail accounts
- Type `bank` to find all banking sites

### Filtering by Type

Currently, you can't filter by type in the UI, but you can search for common patterns:
- Login credentials: Search by URL or username
- Cards: Search by card name or last 4 digits
- Notes: Search by note title

## Security Best Practices

### Passphrase Security

1. **Never share your master passphrase** - Not even with family or friends
2. **Don't write it digitally** - No text files, emails, or notes apps
3. **Use a unique passphrase** - Don't reuse from other services
4. **Make it strong** - At least 16 characters, mixed case, numbers, symbols

### Recovery Key Security

1. **Write it down on paper** - Don't store it digitally
2. **Store it safely** - Safe, safety deposit box, or trusted location
3. **Keep it separate** - Don't store it with your computer
4. **Consider multiple copies** - Store in different secure locations
5. **Never share it** - It's as powerful as your master passphrase

### Database Security

1. **Backup regularly** - Copy `~/.config/openpasswd/` to secure storage
2. **Encrypt backups** - Use encrypted USB drives or cloud storage
3. **Protect your device** - Use full disk encryption
4. **Lock your screen** - When stepping away from your computer
5. **Keep software updated** - Run `openpasswd upgrade` regularly

### Password Hygiene

1. **Use unique passwords** - Never reuse passwords across sites
2. **Use strong passwords** - At least 12 characters, random when possible
3. **Change compromised passwords** - If a site is breached, change immediately
4. **Enable 2FA everywhere** - Use TOTP on all services that support it
5. **Review regularly** - Check for old or unused accounts

## Backup and Recovery

### Backing Up Your Passwords

Your password database is stored in `~/.config/openpasswd/`. Back up this entire directory:

```bash
# Create a backup
tar -czf openpasswd-backup-$(date +%Y%m%d).tar.gz ~/.config/openpasswd/

# Store the backup securely
# - Encrypted USB drive
# - Encrypted cloud storage (Proton Drive, Cryptomator, etc.)
# - External hard drive in a safe
```

**What to backup:**
- `passwords.db` - Your encrypted passwords
- `salt` - Required for decryption
- `kdf_version` - KDF configuration
- `recovery_key` - Encrypted recovery key
- `recovery_hash` - Recovery key verification
- `totp_secret` - If you use TOTP
- `config.toml` - Your preferences

### Restoring from Backup

```bash
# Extract backup
tar -xzf openpasswd-backup-20250111.tar.gz

# Copy to config directory
cp -r openpasswd/* ~/.config/openpasswd/

# Verify
openpasswd list
```

### Using Your Recovery Key

If you forget your master passphrase:

1. You'll need your 24-word recovery key
2. Run the recovery process (feature coming soon)
3. You'll be able to set a new master passphrase

**Current workaround:** The recovery key feature is still in development. For now, make sure you never forget your master passphrase!

### Migrating to a New Computer

1. **On old computer:**
   ```bash
   # Create backup
   tar -czf openpasswd-backup.tar.gz ~/.config/openpasswd/
   ```

2. **Transfer the backup** to your new computer (USB drive, secure file transfer)

3. **On new computer:**
   ```bash
   # Install OpenPasswd
   curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
   
   # Extract backup
   tar -xzf openpasswd-backup.tar.gz
   cp -r openpasswd/* ~/.config/openpasswd/
   
   # Verify
   openpasswd list
   ```

## Advanced Usage

### Upgrading Encryption Strength

OpenPasswd supports multiple KDF (Key Derivation Function) versions. Newer versions are more secure but slower.

**Check current version:**
```bash
cat ~/.config/openpasswd/kdf_version
```

**Versions:**
- `1` - PBKDF2 with 100k iterations (legacy)
- `2` - PBKDF2 with 600k iterations (current default)
- `3` - Argon2id (future default)

**Upgrade to 600k iterations:**
```bash
openpasswd migrate upgrade-kdf
```

This re-encrypts all passwords with stronger key derivation. It takes a few seconds and makes your passwords 6x harder to crack if your database is stolen.

### Customizing Keybindings

Edit `~/.config/openpasswd/config.toml`:

```toml
[keybindings]
quit = ":q"           # Vim-style quit
quit_alt = "ctrl+c"   # Alternative quit
back = "esc"          # Go back
up = "up"             # Move up
up_alt = "k"          # Vim-style up
down = "down"         # Move down
down_alt = "j"        # Vim-style down
select = "enter"      # Select item
```

Change any binding to your preference. Restart OpenPasswd to apply changes.

### Disabling Auto-Update Checks

OpenPasswd checks for updates once per day (cached). To disable:

```bash
openpasswd version --disable-checking
```

To re-enable:
```bash
openpasswd version --enable-checking
```

You can still manually check for updates:
```bash
openpasswd version --check
```

### Using Aliases

The installer creates convenient aliases:

```bash
openpasswd list   # Full command
openpass list     # Short alias
pw list           # Ultra-short alias
```

All three work identically. Use whichever you prefer.

### Shell Completions

The installer sets up completions for Bash and Zsh. If they're not working:

**Bash:**
```bash
source /etc/bash_completion.d/openpass
```

**Zsh:**
```bash
# Add to ~/.zshrc
fpath=(/usr/local/share/zsh/site-functions $fpath)
autoload -Uz compinit && compinit
```

### Viewing Detailed Version Info

```bash
openpasswd version --verbose
```

Shows:
- Version number
- Git commit hash
- Build date
- Go version
- Platform

## Troubleshooting

For common issues and solutions, see the [Troubleshooting Guide](TROUBLESHOOTING.md).

## Getting Help

- **Built-in help:** `openpasswd help`
- **Command help:** `openpasswd <command> --help`
- **GitHub Issues:** [Report bugs](https://github.com/r2unit/openpasswd/issues)
- **Discussions:** [Ask questions](https://github.com/r2unit/openpasswd/discussions)

## Next Steps

- Read the [Security Architecture](SECURITY.md) to understand how OpenPasswd protects your data
- Check out the [CLI Reference](CLI_REFERENCE.md) for all commands and options
- Learn about [importing passwords](IMPORT_GUIDE.md) from other password managers
- Explore [configuration options](CONFIGURATION.md) to customize OpenPasswd
