# Testing Instructions for Passphrase Fix

## The Debug Build

The current build includes debug output that will show exactly what bytes are being captured for your passphrase.

## Test Procedure

1. **Clean slate:**
   ```bash
   rm -rf ~/.config/openpasswd
   ```

2. **Initialize (this uses the TUI):**
   ```bash
   ./openpasswd init
   ```
   - Type a simple passphrase like `test123`
   - You'll see: `[SETUP DEBUG] Passphrase set: len=7, bytes=[116 101 115 116 49 50 51]`
   - Complete the setup

3. **Try to add a password (this uses terminal prompt):**
   ```bash
   ./openpasswd add
   ```
   - Type the SAME passphrase `test123`
   - You'll see: `[TERM DEBUG] Password read: len=7, bytes=[116 101 115 116 49 50 51]`
   
4. **Compare the bytes:**
   - If the byte arrays match exactly, the passphrase is being captured correctly
   - If they don't match, we found the bug!

## What to look for

- The byte arrays should be IDENTICAL
- `test123` should be: `[116 101 115 116 49 50 51]`
- Any difference (extra bytes, missing bytes, different values) indicates where the bug is

Please run this test and share the debug output!
