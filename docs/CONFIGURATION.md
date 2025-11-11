# Configuration Guide

Complete guide to configuring and customizing OpenPasswd.

## Table of Contents

- [Configuration Files](#configuration-files)
- [Keybindings](#keybindings)
- [Colors](#colors)
- [Security Settings](#security-settings)
- [Advanced Configuration](#advanced-configuration)
- [Environment Variables](#environment-variables)

## Configuration Files

OpenPasswd stores all configuration in `~/.config/openpasswd/`:

```
~/.config/openpasswd/
├── config.toml           # User preferences (keybindings, colors)
├── passwords.db          # Encrypted password database
├── salt                  # Encryption salt (base64)
├── kdf_version           # KDF version (1, 2, or 3)
├── recovery_key          # Encrypted recovery key
├── recovery_hash         # Recovery key verification hash
├── totp_secret           # TOTP secret (if enabled)
├── yubikey_challenge     # YubiKey challenge (if enabled)
└── disable_version_check # Flag to disable auto-updates
```

### File Permissions

OpenPasswd automatically sets secure permissions:

| File | Permissions | Description |
|------|-------------|-------------|
| `config.toml` | 0644 | Not sensitive, readable by owner |
| `passwords.db` | 0600 | Encrypted, owner only |
| `salt` | 0600 | Required for decryption, owner only |
| `kdf_version` | 0600 | Security configuration, owner only |
| `recovery_key` | 0600 | Encrypted backup, owner only |
| `recovery_hash` | 0600 | Verification hash, owner only |
| `totp_secret` | 0600 | MFA secret, owner only |

**Verify permissions:**
```bash
ls -la ~/.config/openpasswd/
```

**Fix permissions if needed:**
```bash
chmod 700 ~/.config/openpasswd/
chmod 600 ~/.config/openpasswd/passwords.db
chmod 600 ~/.config/openpasswd/salt
chmod 600 ~/.config/openpasswd/*_key
chmod 600 ~/.config/openpasswd/*_secret
```

## Keybindings

Customize keyboard shortcuts in `config.toml`.

### Default Keybindings

```toml
[keybindings]
quit = ":q"           # Quit application
quit_alt = "ctrl+c"   # Alternative quit
back = "esc"          # Go back / Cancel
up = "up"             # Move up
up_alt = "k"          # Vim-style up
down = "down"         # Move down
down_alt = "j"        # Vim-style down
select = "enter"      # Select / Confirm
```

### Customizing Keybindings

Edit `~/.config/openpasswd/config.toml`:

```toml
[keybindings]
# Use Emacs-style bindings
up = "ctrl+p"
down = "ctrl+n"
quit = "ctrl+x ctrl+c"

# Or keep vim-style
up_alt = "k"
down_alt = "j"
quit = ":q"
```

**Available Key Names:**
- Letters: `a`, `b`, `c`, etc.
- Numbers: `1`, `2`, `3`, etc.
- Special: `enter`, `esc`, `tab`, `space`, `backspace`, `delete`
- Arrows: `up`, `down`, `left`, `right`
- Function: `f1`, `f2`, ..., `f12`
- Modifiers: `ctrl+`, `alt+`, `shift+`

**Examples:**
```toml
# Emacs-style
quit = "ctrl+x ctrl+c"
back = "ctrl+g"
up = "ctrl+p"
down = "ctrl+n"

# Custom
quit = "ctrl+q"
select = "space"
back = "backspace"

# Disable vim-style alternatives
up_alt = ""
down_alt = ""
```

**Notes:**
- Changes take effect on next run
- Invalid bindings fall back to defaults
- Can't bind single letters (reserved for search)
- Some keys may not work in all terminals

### Keybinding Reference

**In List View:**

| Action | Default | Alternative | Description |
|--------|---------|-------------|-------------|
| Move up | `↑` | `k` | Navigate up |
| Move down | `↓` | `j` | Navigate down |
| Select | `Enter` | - | View details |
| Search | `/` | - | Start search |
| Copy password | `c` | - | Copy to clipboard |
| Copy username | `u` | - | Copy username |
| Copy URL | `l` | - | Copy URL |
| Edit | `e` | - | Edit entry |
| Delete | `d` | - | Delete entry |
| Back | `Esc` | - | Clear search |
| Quit | `:q` | `Ctrl+C` | Exit |

**In Detail View:**

| Action | Key | Description |
|--------|-----|-------------|
| Copy password | `c` | Copy to clipboard |
| Copy username | `u` | Copy username |
| Copy URL | `l` | Copy URL |
| Back | `Esc` | Return to list |
| Quit | `:q` or `Ctrl+C` | Exit |

## Colors

Color customization is planned for a future release.

### Planned Color Configuration

```toml
[colors]
# Foreground colors
primary = "#00FF00"
secondary = "#0088FF"
accent = "#FF00FF"
error = "#FF0000"
warning = "#FFAA00"
success = "#00FF00"
info = "#0088FF"

# Background colors
background = "#000000"
background_alt = "#1A1A1A"

# UI elements
border = "#444444"
selected = "#0088FF"
cursor = "#00FF00"

# Syntax highlighting
keyword = "#FF00FF"
string = "#00FF00"
number = "#0088FF"
comment = "#666666"
```

### Current Color Scheme

OpenPasswd uses a default color scheme optimized for dark terminals:

- **Primary**: Green (success, selected items)
- **Secondary**: Blue (info, links)
- **Accent**: Cyan (highlights)
- **Error**: Red (errors, warnings)
- **Warning**: Yellow (cautions)

**Terminal Compatibility:**
- Works with 16-color terminals
- Enhanced with 256-color support
- Best with true color (24-bit) terminals

**Recommended Terminals:**
- **Linux**: GNOME Terminal, Konsole, Alacritty, kitty
- **macOS**: iTerm2, Terminal.app, Alacritty
- **Windows**: Windows Terminal, ConEmu

## Security Settings

### KDF Configuration

Key Derivation Function settings are stored in `kdf_version`:

**Versions:**
- `1` - PBKDF2 with 100,000 iterations (legacy)
- `2` - PBKDF2 with 600,000 iterations (current default)
- `3` - Argon2id (future default)

**Check current version:**
```bash
cat ~/.config/openpasswd/kdf_version
```

**Upgrade KDF:**
```bash
openpasswd migrate upgrade-kdf
```

**Manual configuration:**
```bash
# Not recommended - use migrate command instead
echo "2" > ~/.config/openpasswd/kdf_version
```

### TOTP Configuration

TOTP secret is stored encrypted in `totp_secret`.

**Enable TOTP:**
```bash
openpasswd settings set-totp
```

**Disable TOTP:**
```bash
openpasswd settings remove-totp
```

**Manual backup:**
```bash
# Backup TOTP secret (encrypted)
cp ~/.config/openpasswd/totp_secret ~/backup/totp_secret.backup
```

**Restore TOTP:**
```bash
# Restore from backup
cp ~/backup/totp_secret.backup ~/.config/openpasswd/totp_secret
```

### Recovery Key

Recovery key is stored encrypted in `recovery_key`.

**View recovery key location:**
```bash
ls -l ~/.config/openpasswd/recovery_key
```

**Backup recovery key:**
```bash
# Backup encrypted recovery key
cp ~/.config/openpasswd/recovery_key ~/backup/recovery_key.backup
cp ~/.config/openpasswd/recovery_hash ~/backup/recovery_hash.backup
```

**Note:** The recovery key itself (24 words) should be written down on paper, not stored digitally.

## Advanced Configuration

### Custom Config Directory

By default, OpenPasswd uses `~/.config/openpasswd/`. To use a different location:

**Option 1: Symlink**
```bash
# Move config to custom location
mv ~/.config/openpasswd ~/my-passwords

# Create symlink
ln -s ~/my-passwords ~/.config/openpasswd
```

**Option 2: Environment Variable (Future)**
```bash
# Will be supported in future release
export OPENPASSWD_CONFIG_DIR=~/my-passwords
openpasswd list
```

### Multiple Profiles

You can maintain multiple password databases:

```bash
# Create profile directories
mkdir -p ~/.config/openpasswd-work
mkdir -p ~/.config/openpasswd-personal

# Initialize each profile
ln -s ~/.config/openpasswd-work ~/.config/openpasswd
openpasswd init
# Set up work passwords

# Switch to personal profile
rm ~/.config/openpasswd
ln -s ~/.config/openpasswd-personal ~/.config/openpasswd
openpasswd init
# Set up personal passwords

# Switch between profiles
rm ~/.config/openpasswd
ln -s ~/.config/openpasswd-work ~/.config/openpasswd
openpasswd list  # Work passwords

rm ~/.config/openpasswd
ln -s ~/.config/openpasswd-personal ~/.config/openpasswd
openpasswd list  # Personal passwords
```

**Better approach (future):**
```bash
# Will be supported in future release
openpasswd --profile work list
openpasswd --profile personal list
```

### Database Backup

**Automated backup script:**

```bash
#!/bin/bash
# backup-openpasswd.sh

BACKUP_DIR=~/backups/openpasswd
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/openpasswd-$DATE.tar.gz"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create encrypted backup
tar -czf "$BACKUP_FILE" ~/.config/openpasswd/

# Keep only last 10 backups
ls -t "$BACKUP_DIR"/openpasswd-*.tar.gz | tail -n +11 | xargs -r rm

echo "Backup created: $BACKUP_FILE"
```

**Run backup:**
```bash
chmod +x backup-openpasswd.sh
./backup-openpasswd.sh
```

**Automated backups with cron:**
```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /path/to/backup-openpasswd.sh
```

### Database Encryption

**Re-encrypt with new passphrase (future):**
```bash
# Will be supported in future release
openpasswd change-passphrase
```

**Current workaround:**
1. Export passwords (when export is available)
2. Reinitialize OpenPasswd with new passphrase
3. Re-import passwords

### Performance Tuning

**KDF Performance:**

Argon2id parameters affect login speed:

```go
// Default parameters (in code)
Time:        3,     // 3 iterations (~0.3s)
Memory:      65536, // 64 MiB
Parallelism: 4,     // 4 threads
```

**Trade-offs:**
- Higher values = More secure, slower login
- Lower values = Less secure, faster login

**Current:** Not user-configurable (requires code change)

**Future:** Will be configurable in `config.toml`

## Environment Variables

### Current (None)

OpenPasswd doesn't currently use environment variables.

### Planned

Future releases will support:

```bash
# Custom config directory
export OPENPASSWD_CONFIG_DIR=~/my-passwords

# Disable colors
export OPENPASSWD_NO_COLOR=1

# Debug logging
export OPENPASSWD_DEBUG=1

# Custom log file
export OPENPASSWD_LOG_FILE=~/openpasswd.log

# Clipboard timeout (seconds)
export OPENPASSWD_CLIPBOARD_TIMEOUT=45

# Disable auto-update checks
export OPENPASSWD_NO_UPDATE_CHECK=1
```

## Configuration Examples

### Minimal Configuration

```toml
# ~/.config/openpasswd/config.toml
# Minimal config with defaults

[keybindings]
quit = ":q"
back = "esc"
```

### Vim-Style Configuration

```toml
# ~/.config/openpasswd/config.toml
# Vim-style keybindings

[keybindings]
quit = ":q"
quit_alt = "ZZ"
back = "esc"
up = "k"
up_alt = "ctrl+p"
down = "j"
down_alt = "ctrl+n"
select = "enter"
```

### Emacs-Style Configuration

```toml
# ~/.config/openpasswd/config.toml
# Emacs-style keybindings

[keybindings]
quit = "ctrl+x ctrl+c"
back = "ctrl+g"
up = "ctrl+p"
down = "ctrl+n"
select = "enter"
```

### Custom Configuration

```toml
# ~/.config/openpasswd/config.toml
# Custom keybindings

[keybindings]
quit = "ctrl+q"
back = "esc"
up = "up"
down = "down"
select = "space"

# Future: Color customization
[colors]
primary = "#00FF00"
error = "#FF0000"
```

## Troubleshooting Configuration

### Config File Not Loading

**Check file exists:**
```bash
ls -la ~/.config/openpasswd/config.toml
```

**Check file permissions:**
```bash
chmod 644 ~/.config/openpasswd/config.toml
```

**Check TOML syntax:**
```bash
# Install TOML linter
go install github.com/pelletier/go-toml/v2/cmd/tomll@latest

# Validate config
tomll ~/.config/openpasswd/config.toml
```

### Keybindings Not Working

**Common issues:**
- Invalid key names
- Conflicting bindings
- Terminal doesn't support key combination

**Debug:**
```bash
# Test keybinding
# Run OpenPasswd and try the key
# If it doesn't work, try alternative

# Check terminal capabilities
infocmp | grep key
```

**Reset to defaults:**
```bash
# Remove config file
rm ~/.config/openpasswd/config.toml

# OpenPasswd will use defaults
openpasswd list
```

### Permission Errors

**Fix all permissions:**
```bash
# Config directory
chmod 700 ~/.config/openpasswd/

# Sensitive files
chmod 600 ~/.config/openpasswd/passwords.db
chmod 600 ~/.config/openpasswd/salt
chmod 600 ~/.config/openpasswd/recovery_key
chmod 600 ~/.config/openpasswd/totp_secret

# Config file
chmod 644 ~/.config/openpasswd/config.toml
```

## Best Practices

### Configuration Management

1. **Backup your config:**
   ```bash
   cp ~/.config/openpasswd/config.toml ~/config.toml.backup
   ```

2. **Version control (optional):**
   ```bash
   # Only config.toml, NOT sensitive files!
   git init ~/openpasswd-config
   cp ~/.config/openpasswd/config.toml ~/openpasswd-config/
   cd ~/openpasswd-config
   git add config.toml
   git commit -m "My OpenPasswd config"
   ```

3. **Document your changes:**
   ```toml
   # ~/.config/openpasswd/config.toml
   # Custom configuration for OpenPasswd
   # Last updated: 2025-01-11
   # Changes: Added Emacs-style keybindings
   
   [keybindings]
   quit = "ctrl+x ctrl+c"
   # ... rest of config
   ```

4. **Test changes:**
   - Make small changes
   - Test immediately
   - Keep backup of working config

### Security

1. **Never commit sensitive files:**
   - Don't version control `passwords.db`, `salt`, `recovery_key`, etc.
   - Only `config.toml` is safe to version control

2. **Protect config directory:**
   ```bash
   chmod 700 ~/.config/openpasswd/
   ```

3. **Regular backups:**
   - Backup entire config directory
   - Store backups encrypted
   - Test restore process

4. **Audit permissions:**
   ```bash
   # Check permissions regularly
   ls -la ~/.config/openpasswd/
   ```

## Further Reading

- [User Guide](USER_GUIDE.md) - Complete usage guide
- [Security Architecture](SECURITY.md) - Security details
- [CLI Reference](CLI_REFERENCE.md) - All commands
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues
