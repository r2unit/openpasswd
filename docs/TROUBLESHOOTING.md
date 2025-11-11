# Troubleshooting Guide

Solutions to common issues and problems with OpenPasswd.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Authentication Problems](#authentication-problems)
- [Database Issues](#database-issues)
- [TUI Problems](#tui-problems)
- [MFA Issues](#mfa-issues)
- [Performance Problems](#performance-problems)
- [Platform-Specific Issues](#platform-specific-issues)
- [Getting Help](#getting-help)

## Installation Issues

### "Command not found" after installation

**Problem:** Running `openpasswd` shows "command not found"

**Solutions:**

1. **Check if binary is in PATH:**
   ```bash
   which openpasswd
   ```
   If nothing is returned, the binary isn't in your PATH.

2. **Check installation location:**
   ```bash
   ls -l /usr/local/bin/openpasswd
   ```

3. **Add to PATH (if needed):**
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH="/usr/local/bin:$PATH"
   
   # Reload shell
   source ~/.bashrc  # or source ~/.zshrc
   ```

4. **Reinstall:**
   ```bash
   curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
   ```

### Installation script fails

**Problem:** Install script exits with error

**Common causes:**

1. **No internet connection:**
   ```bash
   # Test connection
   ping github.com
   ```

2. **Go not installed (for source build):**
   ```bash
   # Check Go installation
   go version
   
   # Install Go if needed
   # https://golang.org/dl/
   ```

3. **Permission denied:**
   ```bash
   # Install script needs sudo for /usr/local/bin
   # You'll be prompted for password
   ```

4. **Git not installed:**
   ```bash
   # Check Git
   git --version
   
   # Install Git
   sudo apt-get install git  # Debian/Ubuntu
   sudo dnf install git      # Fedora
   brew install git          # macOS
   ```

### Build from source fails

**Problem:** `make install` or `go build` fails

**Solutions:**

1. **Check Go version:**
   ```bash
   go version
   # Requires Go 1.21 or later
   ```

2. **Update dependencies:**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Clean and rebuild:**
   ```bash
   make clean
   make build
   ```

4. **Check for errors:**
   ```bash
   go build -v ./cmd/client
   # -v shows verbose output
   ```

## Authentication Problems

### "Wrong passphrase" error

**Problem:** Can't access passwords, getting wrong passphrase error

**Solutions:**

1. **Double-check passphrase:**
   - Passphrases are case-sensitive
   - Check Caps Lock
   - Try typing slowly

2. **Check keyboard layout:**
   - Ensure correct keyboard layout is active
   - Special characters may be different

3. **Use recovery key:**
   - If you've forgotten your passphrase
   - Recovery key feature (coming soon)
   - For now, you'll need to reinitialize (loses all passwords)

4. **Check database integrity:**
   ```bash
   # Verify database file exists
   ls -l ~/.config/openpasswd/passwords.db
   
   # Check if it's corrupted
   file ~/.config/openpasswd/passwords.db
   # Should show: ASCII text or data
   ```

5. **Restore from backup:**
   ```bash
   # If you have a backup
   cp ~/backup/openpasswd/* ~/.config/openpasswd/
   ```

### TOTP code not working

**Problem:** 6-digit TOTP code is rejected

**Solutions:**

1. **Check time synchronization:**
   ```bash
   # TOTP requires accurate time
   date
   
   # Sync time (Linux)
   sudo ntpdate pool.ntp.org
   # or
   sudo timedatectl set-ntp true
   
   # Sync time (macOS)
   sudo sntp -sS time.apple.com
   ```

2. **Wait for new code:**
   - TOTP codes expire every 30 seconds
   - Wait for the next code
   - Don't reuse old codes

3. **Check authenticator app:**
   - Ensure app is on correct account
   - Check if app time is synced
   - Try removing and re-adding TOTP

4. **Disable and re-enable TOTP:**
   ```bash
   openpasswd settings remove-totp
   openpasswd settings set-totp
   # Scan new QR code
   ```

### Can't remember which MFA methods are enabled

**Problem:** Not sure if TOTP or YubiKey is enabled

**Solution:**

```bash
# Check for TOTP
ls ~/.config/openpasswd/totp_secret
# If exists, TOTP is enabled

# Check for YubiKey
ls ~/.config/openpasswd/yubikey_challenge
# If exists, YubiKey is enabled
```

## Database Issues

### Database file corrupted

**Problem:** Error messages about corrupted database

**Solutions:**

1. **Check file integrity:**
   ```bash
   # View database file
   cat ~/.config/openpasswd/passwords.db
   # Should be valid JSON
   ```

2. **Restore from backup:**
   ```bash
   cp ~/backup/passwords.db ~/.config/openpasswd/passwords.db
   ```

3. **Try to repair (manual):**
   ```bash
   # Backup first!
   cp ~/.config/openpasswd/passwords.db ~/passwords.db.broken
   
   # Try to fix JSON
   # Open in text editor and fix syntax errors
   nano ~/.config/openpasswd/passwords.db
   ```

4. **Reinitialize (last resort):**
   ```bash
   # WARNING: This deletes all passwords!
   rm -rf ~/.config/openpasswd/
   openpasswd init
   ```

### "Permission denied" errors

**Problem:** Can't read or write database

**Solutions:**

1. **Check file permissions:**
   ```bash
   ls -la ~/.config/openpasswd/
   ```

2. **Fix permissions:**
   ```bash
   chmod 700 ~/.config/openpasswd/
   chmod 600 ~/.config/openpasswd/passwords.db
   chmod 600 ~/.config/openpasswd/salt
   ```

3. **Check ownership:**
   ```bash
   ls -l ~/.config/openpasswd/passwords.db
   # Should be owned by you
   
   # Fix ownership if needed
   sudo chown $USER:$USER ~/.config/openpasswd/*
   ```

### Database not found

**Problem:** "Configuration not initialized" error

**Solutions:**

1. **Check if initialized:**
   ```bash
   ls ~/.config/openpasswd/
   # Should have: salt, passwords.db, etc.
   ```

2. **Initialize:**
   ```bash
   openpasswd init
   ```

3. **Check config directory:**
   ```bash
   # Ensure directory exists
   mkdir -p ~/.config/openpasswd/
   ```

## TUI Problems

### Terminal display issues

**Problem:** Garbled text, weird characters, broken layout

**Solutions:**

1. **Check terminal type:**
   ```bash
   echo $TERM
   # Should be: xterm-256color, screen-256color, etc.
   
   # Set if needed
   export TERM=xterm-256color
   ```

2. **Update terminal:**
   - Use a modern terminal emulator
   - Recommended: iTerm2 (macOS), GNOME Terminal (Linux), Windows Terminal (Windows)

3. **Check terminal size:**
   ```bash
   # Minimum recommended: 80x24
   tput cols  # Width
   tput lines # Height
   ```

4. **Resize terminal:**
   - Make terminal window larger
   - OpenPasswd needs at least 80 columns

5. **Clear screen:**
   ```bash
   clear
   # or
   reset
   ```

### Colors not working

**Problem:** No colors or wrong colors

**Solutions:**

1. **Check color support:**
   ```bash
   # Test colors
   tput colors
   # Should return 256 or more
   ```

2. **Enable 256 colors:**
   ```bash
   export TERM=xterm-256color
   
   # Add to ~/.bashrc or ~/.zshrc
   echo 'export TERM=xterm-256color' >> ~/.bashrc
   ```

3. **Disable colors (if needed):**
   ```bash
   # Future feature
   export OPENPASSWD_NO_COLOR=1
   ```

### Keyboard shortcuts not working

**Problem:** Keys don't do what they should

**Solutions:**

1. **Check keybindings:**
   ```bash
   cat ~/.config/openpasswd/config.toml
   ```

2. **Reset to defaults:**
   ```bash
   rm ~/.config/openpasswd/config.toml
   openpasswd list
   ```

3. **Check terminal key mapping:**
   ```bash
   # Some terminals don't support all key combinations
   # Try alternative keybindings
   ```

4. **Use alternative keys:**
   - Instead of `:q`, use `Ctrl+C`
   - Instead of `k/j`, use arrow keys

### TUI freezes or hangs

**Problem:** Interface becomes unresponsive

**Solutions:**

1. **Force quit:**
   ```bash
   # Press Ctrl+C
   # or
   # Press Ctrl+Z, then: kill %1
   ```

2. **Check for long operations:**
   - Large password lists may take time to load
   - Wait a few seconds

3. **Check system resources:**
   ```bash
   top
   # Check if OpenPasswd is using excessive CPU/memory
   ```

4. **Restart:**
   ```bash
   # Close and reopen
   openpasswd list
   ```

## MFA Issues

### Lost authenticator app

**Problem:** Can't access TOTP codes

**Solutions:**

1. **Use recovery key (future):**
   - Recovery key feature coming soon
   - Will allow disabling TOTP

2. **Current workaround:**
   ```bash
   # Remove TOTP manually
   rm ~/.config/openpasswd/totp_secret
   
   # Now you can access without TOTP
   openpasswd list
   ```

3. **Prevention:**
   - Backup TOTP secret when setting up
   - Use multiple authenticator apps
   - Store backup codes

### QR code not scanning

**Problem:** Can't scan TOTP QR code

**Solutions:**

1. **Increase terminal size:**
   - QR code needs space to display
   - Make terminal window larger

2. **Adjust terminal zoom:**
   - Zoom out to make QR code smaller
   - Zoom in if it's too small

3. **Use manual entry:**
   - QR code displays secret key below
   - Manually enter in authenticator app

4. **Take screenshot:**
   - Screenshot the QR code
   - Scan from photo

5. **Use different terminal:**
   - Some terminals render QR codes better
   - Try iTerm2 (macOS) or Windows Terminal

## Performance Problems

### Slow startup

**Problem:** OpenPasswd takes long to start

**Causes:**

1. **KDF iterations:**
   - Higher iterations = slower but more secure
   - 600k iterations takes ~0.5 seconds

2. **Large database:**
   - Many passwords take time to decrypt

**Solutions:**

1. **This is normal:**
   - Security requires time
   - 0.5-1 second is expected

2. **Check KDF version:**
   ```bash
   cat ~/.config/openpasswd/kdf_version
   # 2 = 600k iterations (current)
   # 1 = 100k iterations (legacy, less secure)
   ```

3. **Upgrade hardware:**
   - Faster CPU helps
   - More RAM helps with Argon2id

### Slow search

**Problem:** Searching passwords is slow

**Solutions:**

1. **Reduce database size:**
   - Remove old/unused passwords
   - Archive rarely-used passwords

2. **Wait for optimizations:**
   - Search performance improvements planned

3. **Use shorter search terms:**
   - Type fewer characters
   - Be more specific

### High memory usage

**Problem:** OpenPasswd uses too much RAM

**Causes:**

1. **Argon2id (future):**
   - Uses 64 MiB by design
   - Memory-hard for security

2. **Large database:**
   - All passwords loaded in memory

**Solutions:**

1. **This is normal:**
   - 64-100 MiB is expected
   - Required for security

2. **Close other applications:**
   - Free up RAM
   - OpenPasswd needs priority

3. **Upgrade RAM:**
   - 4 GB minimum recommended
   - 8 GB or more ideal

## Platform-Specific Issues

### Linux

**Issue: Clipboard not working**

```bash
# Install xclip or xsel
sudo apt-get install xclip  # Debian/Ubuntu
sudo dnf install xclip      # Fedora
sudo pacman -S xclip        # Arch

# Or use xsel
sudo apt-get install xsel
```

**Issue: Permission denied on /usr/local/bin**

```bash
# Use sudo for installation
sudo make install

# Or install to user directory
make build
mkdir -p ~/bin
cp openpasswd ~/bin/
export PATH="$HOME/bin:$PATH"
```

### macOS

**Issue: "openpasswd" cannot be opened**

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/openpasswd

# Or allow in System Preferences
# Security & Privacy → Allow anyway
```

**Issue: Clipboard not clearing**

```bash
# macOS clipboard behavior is different
# Auto-clear may not work in all apps
# Manually clear: pbcopy < /dev/null
```

### Windows

**Issue: Terminal encoding problems**

```powershell
# Set UTF-8 encoding
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# Or use Windows Terminal (recommended)
```

**Issue: Colors not working in Command Prompt**

```powershell
# Use PowerShell or Windows Terminal instead
# Command Prompt has limited color support
```

**Issue: Paths with spaces**

```powershell
# Use quotes
openpasswd --config "C:\Users\My Name\openpasswd"
```

## Getting Help

### Diagnostic Information

When reporting issues, include:

```bash
# Version information
openpasswd version --verbose

# System information
uname -a  # Linux/macOS
systeminfo | findstr /B /C:"OS Name" /C:"OS Version"  # Windows

# Terminal information
echo $TERM
tput colors

# Config directory contents (no sensitive data)
ls -la ~/.config/openpasswd/

# Error messages
# Copy the exact error message
```

### Debug Mode (Future)

```bash
# Will be available in future release
export OPENPASSWD_DEBUG=1
openpasswd list 2>&1 | tee debug.log

# Share debug.log (remove sensitive data first!)
```

### Common Error Messages

**"Configuration not initialized"**
- Run: `openpasswd init`

**"Wrong passphrase"**
- Check passphrase carefully
- Try recovery key (future)

**"Database file not found"**
- Check: `ls ~/.config/openpasswd/passwords.db`
- Restore from backup or reinitialize

**"Permission denied"**
- Fix permissions: `chmod 600 ~/.config/openpasswd/*`

**"TOTP code invalid"**
- Sync system time
- Wait for new code

**"Failed to decrypt"**
- Database may be corrupted
- Restore from backup

### Reporting Bugs

1. **Check existing issues:**
   - https://github.com/r2unit/openpasswd/issues

2. **Create new issue:**
   - Include version info
   - Describe steps to reproduce
   - Include error messages
   - Remove sensitive data!

3. **Security issues:**
   - Don't open public issue
   - Use GitHub Security Advisories
   - Email: security@openpasswd.dev (coming soon)

### Community Support

- **GitHub Discussions:** https://github.com/r2unit/openpasswd/discussions
- **Issues:** https://github.com/r2unit/openpasswd/issues
- **Documentation:** https://github.com/r2unit/openpasswd/tree/master/docs

### Emergency Recovery

**Lost all passwords:**

1. **Check backups:**
   ```bash
   ls ~/backup/openpasswd/
   ls ~/.config/openpasswd.backup/
   ```

2. **Check Time Machine (macOS):**
   - Restore ~/.config/openpasswd/

3. **Check system backups:**
   - Linux: check /var/backups/
   - Windows: check File History

4. **Recovery key (future):**
   - Use 24-word recovery key
   - Restore access to database

**Database corrupted:**

1. **Try backup:**
   ```bash
   cp ~/backup/passwords.db ~/.config/openpasswd/
   ```

2. **Try manual repair:**
   - Open database in text editor
   - Fix JSON syntax errors
   - Save and retry

3. **Last resort:**
   - Reinitialize (loses all passwords)
   - Manually re-enter passwords

## Prevention Tips

### Regular Backups

```bash
# Daily backup script
#!/bin/bash
tar -czf ~/backups/openpasswd-$(date +%Y%m%d).tar.gz ~/.config/openpasswd/

# Run daily with cron
0 2 * * * /path/to/backup-script.sh
```

### Test Backups

```bash
# Regularly test restore process
tar -xzf ~/backups/openpasswd-20250111.tar.gz -C /tmp/
ls /tmp/openpasswd/
```

### Keep Software Updated

```bash
# Check for updates
openpasswd version --check

# Upgrade
openpasswd upgrade
```

### Monitor Disk Space

```bash
# Check available space
df -h ~/.config/

# Clean old backups if needed
```

### Document Your Setup

- Write down your master passphrase (securely!)
- Store recovery key safely
- Document any custom configuration
- Note which MFA methods are enabled

## Still Having Issues?

If you've tried everything and still have problems:

1. **Search documentation:** Check all docs in `docs/` directory
2. **Search issues:** Someone may have had the same problem
3. **Ask community:** GitHub Discussions
4. **Report bug:** GitHub Issues (with details)
5. **Start fresh:** Backup, reinitialize, restore

Remember: Your password security is worth the effort!
