# Import Guide

Guide to importing passwords from other password managers into OpenPasswd.

## Table of Contents

- [Overview](#overview)
- [Supported Password Managers](#supported-password-managers)
- [Proton Pass](#proton-pass)
- [Coming Soon](#coming-soon)
- [General Import Tips](#general-import-tips)
- [Troubleshooting](#troubleshooting)

## Overview

OpenPasswd can import passwords from various password managers. All imported passwords are re-encrypted with your OpenPasswd master passphrase before storage.

**Import Process:**
1. Export passwords from your current password manager
2. Run OpenPasswd import command (when available)
3. Enter your OpenPasswd master passphrase
4. Passwords are decrypted, converted, and re-encrypted
5. Original export file can be securely deleted

**Security Notes:**
- Export files contain your passwords in plaintext or weakly encrypted
- Delete export files immediately after import
- Use encrypted exports when available
- Never email or upload export files

## Supported Password Managers

### Currently Supported
- ✅ Proton Pass (JSON, CSV, encrypted ZIP)

### Coming Soon
- 🔄 Bitwarden (JSON export)
- 🔄 1Password (1PIF, CSV)
- 🔄 LastPass (CSV export)
- 🔄 KeePass (XML, KDBX)
- 🔄 Chrome/Firefox (CSV)

## Proton Pass

Proton Pass is a privacy-focused password manager from the makers of ProtonMail.

### Export from Proton Pass

**Step 1: Open Proton Pass**
- Browser extension or web app (pass.proton.me)

**Step 2: Access Settings**
- Click the gear icon (⚙️) in the top right
- Navigate to "Export" tab

**Step 3: Choose Export Format**

**Option A: Encrypted ZIP (Recommended)**
- Most secure option
- Requires passphrase to decrypt
- Best for importing into OpenPasswd

**Option B: Unencrypted ZIP**
- Easier to work with
- Less secure (passwords in plaintext)
- Delete immediately after import

**Option C: CSV**
- Simple format
- No encryption
- Limited metadata

**Step 4: Download Export**
- Enter passphrase if using encrypted export
- Save file to secure location
- Remember the passphrase (you'll need it for import)

### Import into OpenPasswd

**Note:** Import functionality is currently disabled in the CLI but the importer code is ready. This will be enabled in a future release.

**When available, the command will be:**

```bash
openpasswd import protonpass /path/to/export.zip
```

**For encrypted exports:**
```bash
openpasswd import protonpass /path/to/export.zip --passphrase
# You'll be prompted for the export passphrase
```

**For JSON exports:**
```bash
openpasswd import protonpass /path/to/export.json
```

**For CSV exports:**
```bash
openpasswd import protonpass /path/to/export.csv
```

### What Gets Imported

**Login Credentials:**
- Name/Title
- Username/Email
- Password
- URL(s) - First URL is used
- Notes
- TOTP URI (stored in custom fields)
- Custom fields

**Secure Notes:**
- Name/Title
- Content (stored as notes)

**Credit Cards:**
- Name/Title
- Cardholder name
- Card number
- CVV/CVC
- Expiration date
- PIN
- Notes

**Identity Information:**
- Name/Title
- Full name
- Email
- Phone number
- Custom fields

### Supported Export Formats

| Format | Encryption | Metadata | Custom Fields | Notes |
|--------|-----------|----------|---------------|-------|
| Encrypted ZIP | ✅ PGP | ✅ Full | ✅ Yes | Recommended |
| Unencrypted ZIP | ❌ None | ✅ Full | ✅ Yes | Delete after import |
| JSON | ❌ None | ✅ Full | ✅ Yes | Delete after import |
| CSV | ❌ None | ⚠️ Limited | ❌ No | Basic data only |

### PGP-Encrypted Exports

If you export with PGP encryption, you'll need GPG installed:

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
- Install [GPG4Win](https://www.gpg4win.org/)

**Import Process:**
1. OpenPasswd extracts the ZIP
2. Finds the PGP-encrypted file
3. Calls `gpg` to decrypt with your passphrase
4. Parses the decrypted JSON
5. Imports passwords

### Example: Complete Proton Pass Migration

```bash
# 1. Export from Proton Pass (encrypted ZIP)
#    - Open Proton Pass
#    - Settings → Export
#    - Choose "Encrypted ZIP"
#    - Enter passphrase: MySecureExportPass123!
#    - Download: protonpass-export-2025-01-11.zip

# 2. Import into OpenPasswd (when available)
openpasswd import protonpass protonpass-export-2025-01-11.zip

# You'll be prompted for:
# - OpenPasswd master passphrase
# - Proton Pass export passphrase
# - TOTP code (if enabled)

# 3. Verify import
openpasswd list
# Search for a password you know exists

# 4. Securely delete export file
shred -u protonpass-export-2025-01-11.zip
# Or on macOS:
rm -P protonpass-export-2025-01-11.zip
```

### Troubleshooting Proton Pass Import

**"Failed to decrypt PGP file"**
- Check that GPG is installed: `gpg --version`
- Verify export passphrase is correct
- Try unencrypted export instead

**"No supported data file found in ZIP"**
- Extract ZIP manually and check contents
- Look for `data.json` or `data.pgp`
- Try exporting again from Proton Pass

**"Invalid JSON format"**
- Proton Pass may have changed export format
- Open an issue on GitHub with format details
- Try CSV export as fallback

**"Some passwords failed to import"**
- Check OpenPasswd logs for details
- Unsupported item types are skipped
- Custom fields may not import perfectly

## Coming Soon

### Bitwarden

**Export Format:** JSON

**How to Export:**
1. Open Bitwarden (web vault or app)
2. Tools → Export Vault
3. Choose "JSON" format
4. Enter master password
5. Download export file

**Import Command (Future):**
```bash
openpasswd import bitwarden /path/to/bitwarden-export.json
```

**What Gets Imported:**
- Login credentials
- Secure notes
- Cards
- Identities
- Custom fields
- TOTP secrets
- Attachments (future)

### 1Password

**Export Format:** 1PIF or CSV

**How to Export:**
1. Open 1Password
2. File → Export → All Items
3. Choose format (1PIF recommended)
4. Save export file

**Import Command (Future):**
```bash
openpasswd import 1password /path/to/1password.1pif
```

**What Gets Imported:**
- Login credentials
- Secure notes
- Credit cards
- Identities
- Software licenses
- Custom fields

### LastPass

**Export Format:** CSV

**How to Export:**
1. Open LastPass web vault
2. More Options → Advanced → Export
3. Enter master password
4. Copy CSV data or download file

**Import Command (Future):**
```bash
openpasswd import lastpass /path/to/lastpass-export.csv
```

**What Gets Imported:**
- Login credentials
- Secure notes
- Basic metadata
- Limited custom fields

### KeePass

**Export Format:** XML or KDBX

**How to Export:**
1. Open KeePass database
2. File → Export
3. Choose "KeePass XML (2.x)" format
4. Save export file

**Import Command (Future):**
```bash
openpasswd import keepass /path/to/keepass-export.xml
```

**What Gets Imported:**
- All entry types
- Groups/folders (converted to tags)
- Custom fields
- Attachments (future)

### Browser Password Managers

**Chrome/Edge:**
1. Settings → Passwords
2. ⋮ (three dots) → Export passwords
3. Save CSV file

**Firefox:**
1. about:logins
2. ⋮ (three dots) → Export Logins
3. Save CSV file

**Import Command (Future):**
```bash
openpasswd import chrome /path/to/chrome-passwords.csv
openpasswd import firefox /path/to/firefox-passwords.csv
```

## General Import Tips

### Before Importing

1. **Backup your current OpenPasswd database**
   ```bash
   tar -czf openpasswd-backup-$(date +%Y%m%d).tar.gz ~/.config/openpasswd/
   ```

2. **Verify export file integrity**
   - Open in text editor to check format
   - Ensure all passwords are present
   - Check for corruption

3. **Prepare your passphrases**
   - OpenPasswd master passphrase
   - Export file passphrase (if encrypted)
   - TOTP code (if enabled)

### During Import

1. **Don't interrupt the process**
   - Import can take several minutes for large databases
   - Progress is shown in terminal
   - Ctrl+C will abort (safe, but incomplete)

2. **Watch for errors**
   - Some entries may fail to import
   - Errors are logged
   - You can retry failed entries

3. **Verify as you go**
   - Check a few passwords during import
   - Ensure data is correct
   - Stop if something looks wrong

### After Importing

1. **Verify all passwords**
   ```bash
   openpasswd list
   # Browse through and spot-check
   ```

2. **Test a few logins**
   - Copy passwords and try logging in
   - Ensure URLs are correct
   - Check custom fields

3. **Securely delete export files**
   ```bash
   # Linux
   shred -u export-file.zip
   
   # macOS
   rm -P export-file.zip
   
   # Windows
   # Use File Shredder or similar tool
   ```

4. **Update your old password manager**
   - Don't delete your old vault yet
   - Keep it as backup for a few weeks
   - Verify everything works in OpenPasswd first

5. **Enable MFA in OpenPasswd**
   ```bash
   openpasswd settings set-totp
   ```

## Troubleshooting

### Common Issues

**"Import command not found"**
- Import functionality is currently disabled
- Will be enabled in a future release
- Use manual entry for now

**"Failed to parse export file"**
- Check file format is supported
- Verify file isn't corrupted
- Try exporting again

**"Some passwords missing after import"**
- Check import logs for errors
- Some entry types may not be supported yet
- Custom fields may need manual entry

**"Duplicate passwords after import"**
- OpenPasswd doesn't automatically deduplicate
- Manually remove duplicates using `openpasswd list`
- Future versions will detect duplicates

**"TOTP secrets not imported"**
- TOTP import is supported for Proton Pass
- Other managers may store TOTP differently
- May need to re-scan QR codes

### Getting Help

If you encounter issues:

1. **Check the logs**
   - Import errors are displayed in terminal
   - Note the error message

2. **Verify export format**
   - Open export file in text editor
   - Check if format matches expected structure

3. **Try alternative export format**
   - If JSON fails, try CSV
   - If encrypted fails, try unencrypted

4. **Report bugs**
   - Open issue on GitHub
   - Include error message (redact sensitive data)
   - Mention export format and source manager

## Security Considerations

### Export File Security

**Risks:**
- Export files contain all your passwords
- Often in plaintext or weakly encrypted
- Can be read by anyone with file access

**Best Practices:**
- Use encrypted exports when available
- Store export files on encrypted drives
- Delete immediately after import
- Never email or upload to cloud
- Use secure file transfer if needed

### Import Process Security

**What OpenPasswd Does:**
- Reads export file
- Decrypts if necessary (using your passphrase)
- Converts to OpenPasswd format
- Re-encrypts with your master passphrase
- Stores in local database
- Clears sensitive data from memory

**What You Should Do:**
- Run import on trusted computer
- Ensure no one is watching your screen
- Use private network (not public WiFi)
- Verify OpenPasswd binary is authentic
- Check file permissions after import

### After Import

**Cleanup:**
```bash
# Securely delete export file
shred -u export-file.zip

# Verify it's gone
ls -la export-file.zip
# Should show "No such file or directory"

# Clear shell history if it contains passphrases
history -c
```

**Verify Security:**
```bash
# Check database permissions
ls -la ~/.config/openpasswd/passwords.db
# Should show: -rw------- (0600)

# Check config directory permissions
ls -ld ~/.config/openpasswd/
# Should show: drwx------ (0700)
```

## Future Enhancements

Planned improvements for import functionality:

- **Duplicate detection** - Warn about duplicate entries
- **Merge conflicts** - Handle existing passwords
- **Batch import** - Import from multiple files
- **Import preview** - Review before importing
- **Selective import** - Choose which passwords to import
- **Tag preservation** - Maintain folder/group structure
- **Attachment import** - Import file attachments
- **Import history** - Track what was imported when

## Contributing

Want to add support for another password manager?

1. Check if it's already planned
2. Open an issue to discuss
3. Implement the importer (see `pkg/sources/`)
4. Follow existing patterns (Proton Pass example)
5. Add tests
6. Update this documentation
7. Submit pull request

See [Developer Guide](DEVELOPER_GUIDE.md) for details.

## References

- [Proton Pass Export Guide](https://proton.me/support/pass-export)
- [Bitwarden Export Guide](https://bitwarden.com/help/export-your-data/)
- [1Password Export Guide](https://support.1password.com/export/)
- [LastPass Export Guide](https://support.lastpass.com/help/how-do-i-nbsp-export-stored-data-from-lastpass-using-a-generic-csv-file)
- [KeePass Export Guide](https://keepass.info/help/base/importexport.html)
