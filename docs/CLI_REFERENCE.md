# CLI Reference

Complete reference for all OpenPasswd commands and options.

## Table of Contents

- [Global Options](#global-options)
- [Commands](#commands)
  - [init](#init)
  - [add](#add)
  - [list](#list)
  - [settings](#settings)
  - [migrate](#migrate)
  - [version](#version)
  - [upgrade](#upgrade)
  - [help](#help)
- [Environment Variables](#environment-variables)
- [Exit Codes](#exit-codes)

## Global Options

These options work with any command:

```bash
--help, -h      Show help for the command
--version, -v   Show version information
```

## Commands

### init

Initialize OpenPasswd configuration and create your master passphrase.

**Usage:**
```bash
openpasswd init
```

**What it does:**
1. Creates configuration directory (`~/.config/openpasswd/`)
2. Prompts for master passphrase
3. Generates encryption salt
4. Creates recovery key (24-word phrase)
5. Encrypts and stores recovery key
6. Creates default configuration file

**Interactive Prompts:**
- Master passphrase (hidden input, confirmed)
- Recovery key display (you must write it down)

**Output Files:**
- `~/.config/openpasswd/salt` - Encryption salt
- `~/.config/openpasswd/kdf_version` - KDF version (2 for new installs)
- `~/.config/openpasswd/recovery_key` - Encrypted recovery key
- `~/.config/openpasswd/recovery_hash` - Recovery key verification hash
- `~/.config/openpasswd/config.toml` - Default configuration
- `~/.config/openpasswd/passwords.db` - Empty database (created on first add)

**Examples:**
```bash
# First-time setup
openpasswd init

# Reinitialize (overwrites existing config)
openpasswd init
# Choose "Override" when prompted
```

**Notes:**
- If config already exists, you'll be asked to ignore or override
- Override deletes ALL existing passwords
- Recovery key is shown only once - write it down!

---

### add

Add a new password entry to your vault.

**Usage:**
```bash
openpasswd add [type]
```

**Password Types:**
- `login` - Login credentials (username, password, URL)
- `card` - Credit/debit card information
- `note` - Secure note
- `identity` - Personal identity information
- `password` - Simple password entry
- `other` - Other credential type

**Interactive Mode:**
```bash
openpasswd add
```
Opens a menu to select password type and enter details.

**Direct Type:**
```bash
openpasswd add login      # Add login credentials
openpasswd add card       # Add credit card
openpasswd add note       # Add secure note
```

**Fields by Type:**

**Login:**
- Name (required)
- Username/Email
- Password
- URL
- Notes
- Custom fields

**Card:**
- Name (required)
- Cardholder name
- Card number
- CVV/CVC
- Expiration date
- PIN
- Notes

**Note:**
- Name (required)
- Content

**Identity:**
- Name (required)
- Full name
- Email
- Phone number
- Address
- Custom fields

**Examples:**
```bash
# Interactive mode
openpasswd add

# Add login directly
openpasswd add login

# Add credit card
openpasswd add card

# Add secure note
openpasswd add note
```

**Notes:**
- Master passphrase is always required
- TOTP code required if TOTP is enabled
- All fields are encrypted before storage
- Passwords are hidden during input

---

### list

List and manage your passwords in an interactive TUI.

**Usage:**
```bash
openpasswd list
```

**Features:**
- Browse all passwords
- Search by name, username, or URL
- View password details
- Copy passwords to clipboard
- Edit entries
- Delete entries

**Keyboard Shortcuts:**

**Navigation:**
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` - View details
- `Esc` - Go back
- `:q` or `Ctrl+C` - Quit

**Actions:**
- `/` - Search
- `c` - Copy password to clipboard
- `u` - Copy username to clipboard
- `l` - Copy URL to clipboard
- `e` - Edit entry
- `d` - Delete entry

**Search:**
- Press `/` to start searching
- Type to filter results in real-time
- Press `Esc` to clear search
- Search is case-insensitive and fuzzy

**Examples:**
```bash
# Open password list
openpasswd list

# You'll be prompted for:
# 1. Master passphrase
# 2. TOTP code (if enabled)
```

**Notes:**
- Passwords are decrypted on-demand
- Clipboard auto-clears after 45 seconds
- Changes are saved immediately
- Wrong passphrase shows error screen

---

### settings

Manage OpenPasswd settings and authentication methods.

**Usage:**
```bash
openpasswd settings <subcommand>
```

**Subcommands:**

#### set-totp

Enable TOTP (Time-based One-Time Password) authentication.

```bash
openpasswd settings set-totp
```

**What it does:**
1. Generates TOTP secret
2. Displays QR code in terminal
3. Shows secret key (for manual entry)
4. Prompts for verification code
5. Saves encrypted TOTP secret

**Usage:**
1. Run the command
2. Scan QR code with authenticator app (Google Authenticator, Authy, etc.)
3. Enter 6-digit code to verify
4. TOTP is now required for all password access

#### remove-totp

Disable TOTP authentication.

```bash
openpasswd settings remove-totp
```

**What it does:**
- Removes TOTP secret from config
- Disables TOTP requirement
- Requires confirmation

#### show-totp-qr

Re-display TOTP QR code (currently not supported).

```bash
openpasswd settings show-totp-qr
```

**Note:** For security reasons, QR code can't be re-displayed. To get a new QR code, remove and re-add TOTP.

#### set-yubikey

Enable YubiKey hardware authentication (coming soon).

```bash
openpasswd settings set-yubikey
```

**Status:** Not yet implemented.

#### remove-yubikey

Disable YubiKey authentication (coming soon).

```bash
openpasswd settings remove-yubikey
```

**Status:** Not yet implemented.

**Examples:**
```bash
# Enable TOTP
openpasswd settings set-totp

# Disable TOTP
openpasswd settings remove-totp

# Show help
openpasswd settings help
```

**Notes:**
- TOTP secret is stored encrypted
- Losing authenticator app requires recovery key
- Multiple MFA methods can be combined (future)

---

### migrate

Migrate database to newer security standards.

**Usage:**
```bash
openpasswd migrate <subcommand>
```

**Subcommands:**

#### upgrade-kdf

Upgrade key derivation function to stronger parameters.

```bash
openpasswd migrate upgrade-kdf
```

**What it does:**
1. Checks current KDF version
2. Prompts for master passphrase
3. Decrypts all passwords with old KDF
4. Re-encrypts with new KDF (600k iterations)
5. Updates KDF version in config

**Versions:**
- Version 1: PBKDF2 with 100k iterations (legacy)
- Version 2: PBKDF2 with 600k iterations (current)
- Version 3: Argon2id (future)

**Migration Path:**
- 1 → 2: Upgrade to 600k iterations (6x stronger)
- 2 → 3: Upgrade to Argon2id (memory-hard, future)

**Examples:**
```bash
# Upgrade from 100k to 600k iterations
openpasswd migrate upgrade-kdf

# Check current version
cat ~/.config/openpasswd/kdf_version
```

**Notes:**
- Migration is safe (doesn't lose data)
- Takes a few seconds (stronger = slower)
- Backup recommended before migrating
- Can't downgrade (one-way migration)

---

### version

Show version information and manage updates.

**Usage:**
```bash
openpasswd version [options]
```

**Options:**

#### (no options)

Show basic version information.

```bash
openpasswd version
```

**Output:**
```
OpenPasswd v0.1.0

Run 'openpasswd version --check' to check for updates
Run 'openpasswd version --verbose' for build information
```

#### --verbose, -v

Show detailed build information.

```bash
openpasswd version --verbose
```

**Output:**
```
OpenPasswd v0.1.0

Build Information:
  Git Commit:  a1b2c3d
  Build Date:  2025-01-11 10:30:00 UTC
  Go Version:  go1.21.5
  Platform:    linux/amd64
```

#### --check, -c

Check for updates interactively.

```bash
openpasswd version --check
```

**What it does:**
1. Fetches latest release from GitHub
2. Compares with current version
3. Shows release notes if update available
4. Prompts to upgrade

**Output (if update available):**
```
Checking for updates...

New version available: v0.2.0
Current version: v0.1.0

Release Notes:
─────────────────────────────────────────
- Added password generation
- Improved TUI performance
- Fixed clipboard issues
─────────────────────────────────────────

Do you want to upgrade? (yes/no):
```

#### --disable-checking

Disable automatic update checks on startup.

```bash
openpasswd version --disable-checking
```

**What it does:**
- Creates `~/.config/openpasswd/disable_version_check` flag file
- Prevents automatic update checks
- Manual checks still work

#### --enable-checking

Re-enable automatic update checks.

```bash
openpasswd version --enable-checking
```

**What it does:**
- Removes disable flag file
- Restores automatic update checks (once per 24 hours)

**Examples:**
```bash
# Show version
openpasswd version

# Show detailed info
openpasswd version --verbose

# Check for updates
openpasswd version --check

# Disable auto-checks
openpasswd version --disable-checking
```

**Notes:**
- Auto-checks are cached (max once per 24 hours)
- Checks are non-blocking (won't slow down startup)
- No telemetry or tracking
- Only checks GitHub releases API

---

### upgrade

Upgrade OpenPasswd to the latest version.

**Usage:**
```bash
openpasswd upgrade
```

**What it does:**
1. Checks for latest release on GitHub
2. Downloads install script
3. Runs installation
4. Replaces current binary
5. Preserves configuration and data

**Interactive Prompts:**
- Confirmation before upgrading
- Shows release notes
- Asks for sudo password (if needed)

**Examples:**
```bash
# Upgrade to latest version
openpasswd upgrade

# Check version after upgrade
openpasswd version
```

**Notes:**
- Requires internet connection
- Requires sudo for system-wide install
- Configuration and passwords are preserved
- Can't downgrade (install older version manually)

---

### help

Show help information.

**Usage:**
```bash
openpasswd help
openpasswd --help
openpasswd -h
```

**Output:**
```
OpenPasswd - A secure, terminal-based password manager

COMMANDS:
    openpasswd init              Initialize configuration and database
    openpasswd add               Add a new password entry
    openpasswd list              List and search passwords
    openpasswd settings          Manage settings (passphrase, MFA, etc.)
    openpasswd version           Show version information
    openpasswd upgrade           Upgrade to the latest version
    openpasswd help              Show this help message

OPTIONS:
    --help, -h                   Show this help message
    --version, -v                Show version number

EXAMPLES:
    openpasswd init                             # First-time setup
    openpasswd add                              # Add password interactively
    openpasswd add login                        # Add login password
    openpasswd list                             # List all passwords
    openpasswd settings set-totp                # Enable TOTP authentication
    openpasswd version --check                  # Check for updates
    openpasswd upgrade                          # Upgrade to latest version

For more information, visit: https://github.com/r2unit/openpasswd
```

**Command-Specific Help:**
```bash
openpasswd add --help        # Help for 'add' command
openpasswd settings --help   # Help for 'settings' command
```

---

## Environment Variables

OpenPasswd doesn't currently use environment variables, but these may be added in future versions:

**Planned:**
- `OPENPASSWD_CONFIG_DIR` - Custom config directory
- `OPENPASSWD_NO_COLOR` - Disable colors in output
- `OPENPASSWD_DEBUG` - Enable debug logging

---

## Exit Codes

OpenPasswd uses standard Unix exit codes:

- `0` - Success
- `1` - General error (wrong passphrase, file not found, etc.)
- `2` - Misuse of command (invalid arguments)
- `130` - Interrupted by user (Ctrl+C)

**Examples:**
```bash
# Check if command succeeded
openpasswd list
echo $?  # 0 if successful, 1 if error

# Use in scripts
if openpasswd list > /dev/null 2>&1; then
    echo "Success"
else
    echo "Failed"
fi
```

---

## Configuration Files

### config.toml

Location: `~/.config/openpasswd/config.toml`

**Format:**
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
```

**Customization:**
- Edit with any text editor
- Changes take effect on next run
- Invalid config falls back to defaults

---

## Shell Completions

### Bash

Completions are installed to `/etc/bash_completion.d/openpass`.

**Manual activation:**
```bash
source /etc/bash_completion.d/openpass
```

**Permanent activation:**
Add to `~/.bashrc`:
```bash
if [ -f /etc/bash_completion.d/openpass ]; then
    . /etc/bash_completion.d/openpass
fi
```

### Zsh

Completions are installed to `/usr/local/share/zsh/site-functions/_openpass`.

**Manual activation:**
```zsh
autoload -Uz compinit && compinit
```

**Permanent activation:**
Add to `~/.zshrc`:
```zsh
fpath=(/usr/local/share/zsh/site-functions $fpath)
autoload -Uz compinit && compinit
```

---

## Aliases

The installer creates convenient aliases:

```bash
openpasswd    # Full command
openpass      # Short alias
pw            # Ultra-short alias
```

All three work identically. Use whichever you prefer.

---

## Tips and Tricks

### Quick Password Lookup

```bash
# Open list and search immediately
openpasswd list
# Then press '/' and type search term
```

### Scripting with OpenPasswd

```bash
# Check if initialized
if [ -f ~/.config/openpasswd/salt ]; then
    echo "OpenPasswd is initialized"
fi

# Backup passwords
tar -czf backup.tar.gz ~/.config/openpasswd/

# Check version programmatically
version=$(openpasswd version | head -1 | awk '{print $2}')
echo "Running version: $version"
```

### Using with tmux/screen

OpenPasswd works great in tmux or screen sessions:

```bash
# In tmux
tmux new-session -s passwords
openpasswd list

# Detach: Ctrl+b, d
# Reattach: tmux attach -t passwords
```

---

## Troubleshooting

For common issues and solutions, see the [Troubleshooting Guide](TROUBLESHOOTING.md).

## Further Reading

- [User Guide](USER_GUIDE.md) - Complete usage guide
- [Security Architecture](SECURITY.md) - Security details
- [Configuration Guide](CONFIGURATION.md) - Customization options
- [Import Guide](IMPORT_GUIDE.md) - Import from other password managers
