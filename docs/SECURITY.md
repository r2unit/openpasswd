# Security Architecture

Detailed documentation of OpenPasswd's security features and cryptographic implementation.

## Table of Contents

- [Overview](#overview)
- [Cryptographic Primitives](#cryptographic-primitives)
- [Encryption Process](#encryption-process)
- [Key Derivation](#key-derivation)
- [Data Storage](#data-storage)
- [Threat Model](#threat-model)
- [Security Features](#security-features)
- [Known Limitations](#known-limitations)
- [Security Audit Status](#security-audit-status)

## Overview

OpenPasswd uses industry-standard cryptography to protect your passwords. All sensitive data is encrypted with AES-256-GCM, keys are derived using Argon2id or PBKDF2, and integrity is verified with BLAKE2b hashing.

**Core Security Principles:**

1. **Zero-knowledge architecture** - Your data never leaves your device
2. **Defense in depth** - Multiple layers of security
3. **Secure by default** - Strong encryption out of the box
4. **Transparent implementation** - Open source, auditable code
5. **Standard cryptography** - No custom crypto, only proven algorithms

## Cryptographic Primitives

### Encryption: AES-256-GCM

**Algorithm:** Advanced Encryption Standard with 256-bit keys in Galois/Counter Mode

**Why AES-256-GCM?**
- Industry standard, widely vetted
- Authenticated encryption (provides both confidentiality and integrity)
- Resistant to padding oracle attacks
- Fast on modern hardware with AES-NI support
- NIST approved for classified information

**Implementation:**
- Uses Go's `crypto/aes` and `crypto/cipher` packages
- 12-byte random nonce per encryption operation
- Authentication tag prevents tampering
- No key reuse (each password has unique nonce)

**Security Margin:**
- 256-bit keys provide ~2^256 security (effectively unbreakable)
- GCM mode provides authenticated encryption
- Nonce collision probability: negligible with crypto/rand

### Key Derivation: Argon2id

**Algorithm:** Argon2id (hybrid of Argon2i and Argon2d)

**Why Argon2id?**
- Winner of Password Hashing Competition (2015)
- Memory-hard (resistant to GPU/ASIC attacks)
- Configurable time, memory, and parallelism
- Recommended by OWASP for password storage
- Resistant to side-channel attacks

**Parameters (Default):**
- Time cost: 3 iterations
- Memory cost: 64 MiB
- Parallelism: 4 threads
- Output: 32 bytes (256 bits)

**Implementation:**
- Custom Argon2id implementation following RFC 9106
- Uses BLAKE2b for internal hashing
- Secure memory handling

**Security Margin:**
- 64 MiB memory makes GPU attacks expensive
- 3 iterations balance security and usability
- Estimated ~0.3 seconds on modern hardware

### Legacy KDF: PBKDF2-HMAC-SHA256

**Algorithm:** Password-Based Key Derivation Function 2 with HMAC-SHA256

**Why PBKDF2?**
- Backward compatibility with older databases
- NIST approved (SP 800-132)
- Simple, well-understood algorithm
- Widely supported

**Parameters:**
- Iterations: 600,000 (OWASP recommended, 2023)
- Hash: SHA-256
- Output: 32 bytes (256 bits)

**Legacy Version:**
- Iterations: 100,000 (deprecated, can be upgraded)

**Security Margin:**
- 600k iterations provides ~0.5 seconds on modern hardware
- Resistant to brute-force attacks
- Can be upgraded to Argon2id via migration

### Integrity: BLAKE2b

**Algorithm:** BLAKE2b-256

**Why BLAKE2b?**
- Faster than SHA-256
- More secure than MD5/SHA-1
- Finalist in SHA-3 competition
- No known vulnerabilities

**Usage:**
- Recovery key hashing
- Integrity verification
- Internal Argon2id operations

**Implementation:**
- Custom BLAKE2b implementation
- 256-bit output
- Constant-time operations

### Random Number Generation

**Source:** Go's `crypto/rand` package

**Properties:**
- Cryptographically secure random number generator (CSPRNG)
- Uses OS-provided entropy:
  - Linux: `/dev/urandom`
  - macOS: `/dev/urandom`
  - Windows: `CryptGenRandom`
- Automatically seeded by OS
- No user-provided seed required

## Encryption Process

### Adding a Password

1. **User Input**
   - User enters password details (name, username, password, URL, etc.)
   - Master passphrase is prompted (not stored)

2. **Key Derivation**
   ```
   Salt (32 bytes, stored) + Master Passphrase
   ↓
   Argon2id (3 iterations, 64 MiB, 4 threads)
   ↓
   Encryption Key (32 bytes)
   ```

3. **Encryption**
   ```
   For each field (name, username, password, URL, notes):
   ↓
   Generate random nonce (12 bytes)
   ↓
   AES-256-GCM encryption
   ↓
   Ciphertext = Nonce || Encrypted Data || Auth Tag
   ↓
   Base64 encode for storage
   ```

4. **Storage**
   - Encrypted fields stored in JSON database
   - Database saved to `~/.config/openpasswd/passwords.db`
   - File permissions: 0600 (owner read/write only)

### Retrieving a Password

1. **Authentication**
   - User enters master passphrase
   - Optional: TOTP code verification

2. **Key Derivation**
   ```
   Salt (loaded from disk) + Master Passphrase
   ↓
   Argon2id (same parameters)
   ↓
   Encryption Key (32 bytes)
   ```

3. **Decryption**
   ```
   For each encrypted field:
   ↓
   Base64 decode
   ↓
   Extract nonce, ciphertext, auth tag
   ↓
   AES-256-GCM decryption
   ↓
   Verify authentication tag
   ↓
   Plaintext data
   ```

4. **Display**
   - Decrypted data shown in TUI
   - Can be copied to clipboard
   - Clipboard auto-clears after 45 seconds

## Key Derivation

### Argon2id Process

```
Input: Master Passphrase + Salt (32 bytes)

Step 1: Initial Hash (H0)
  H0 = BLAKE2b(
    parallelism || key_length || memory || iterations ||
    version || type || password_length || password ||
    salt_length || salt
  )

Step 2: Memory Initialization
  For each lane (4 parallel lanes):
    Block[0] = BLAKE2b(H0 || 0 || lane_id)
    Block[1] = BLAKE2b(H0 || 1 || lane_id)

Step 3: Memory Filling (3 iterations)
  For each iteration:
    For each segment:
      For each lane:
        Compute new block using:
          - Previous block
          - Reference block (data-dependent or independent)
          - Blake2b mixing function

Step 4: Finalization
  XOR all last blocks of all lanes
  ↓
  BLAKE2b-Long (variable output)
  ↓
  Encryption Key (32 bytes)
```

### PBKDF2 Process

```
Input: Master Passphrase + Salt (32 bytes)

Step 1: HMAC-SHA256 Setup
  PRF = HMAC-SHA256(password)

Step 2: Iteration
  For i = 1 to 600,000:
    U_i = PRF(U_{i-1})
    Result = Result XOR U_i

Step 3: Output
  Encryption Key (32 bytes)
```

### Salt Generation

```
Salt Generation:
  crypto/rand.Read(32 bytes)
  ↓
  Base64 encode
  ↓
  Store in ~/.config/openpasswd/salt
```

**Properties:**
- 32 bytes (256 bits) of entropy
- Unique per installation
- Never changes (tied to database)
- Stored in plaintext (not secret, just unique)

## Data Storage

### File Structure

```
~/.config/openpasswd/
├── passwords.db          # Encrypted password database
├── salt                  # Encryption salt (base64)
├── kdf_version           # KDF version number (1, 2, or 3)
├── recovery_key          # Encrypted recovery key
├── recovery_hash         # Recovery key verification hash
├── totp_secret           # TOTP secret (base64, if enabled)
├── config.toml           # User preferences (not sensitive)
└── disable_version_check # Flag file (if update checks disabled)
```

### Database Format

**File:** `passwords.db` (JSON)

```json
{
  "next_id": 5,
  "passwords": {
    "1": {
      "ID": 1,
      "Type": "login",
      "Name": "base64_encrypted_name",
      "Username": "base64_encrypted_username",
      "Password": "base64_encrypted_password",
      "URL": "base64_encrypted_url",
      "Notes": "base64_encrypted_notes",
      "Fields": {
        "custom_field": "base64_encrypted_value"
      },
      "CreatedAt": "2025-01-11T10:30:00Z",
      "UpdatedAt": "2025-01-11T10:30:00Z"
    }
  }
}
```

**Encryption:**
- All sensitive fields encrypted with AES-256-GCM
- Each field has unique nonce
- Base64 encoded for JSON storage
- Timestamps stored in plaintext (not sensitive)

**File Permissions:**
- `0600` - Owner read/write only
- No group or world access
- Enforced by OpenPasswd on save

### Recovery Key Storage

**Recovery Key Generation:**
```
Generate 32 bytes of entropy
↓
Map to 24 words from BIP39-style wordlist
↓
Format: word1-word2-word3-word4 (6 groups of 4)
```

**Storage:**
```
Recovery Key (plaintext, user writes down)
↓
Encrypt with master passphrase + salt
↓
Store in ~/.config/openpasswd/recovery_key (encrypted)

Also:
Recovery Key
↓
BLAKE2b-256 hash
↓
Store in ~/.config/openpasswd/recovery_hash (for verification)
```

## Threat Model

### What OpenPasswd Protects Against

✅ **Offline Attacks**
- Stolen database file
- Backup theft
- Disk imaging
- Cold boot attacks (if disk is encrypted)

✅ **Brute-Force Attacks**
- Dictionary attacks on master passphrase
- Rainbow table attacks
- GPU-accelerated cracking (Argon2id memory-hard)

✅ **Data Tampering**
- Modified database entries
- Corrupted data
- Malicious modifications

✅ **Unauthorized Access**
- Other users on same system (file permissions)
- Malware reading database file (encrypted)
- Network interception (no network communication)

### What OpenPasswd Does NOT Protect Against

❌ **Active Attacks on Running System**
- Keyloggers capturing master passphrase
- Screen capture while viewing passwords
- Memory dumps while passwords are decrypted
- Malware with root/admin privileges

❌ **Physical Access to Unlocked System**
- Someone using your computer while unlocked
- Shoulder surfing while entering passphrase
- Physical keyloggers on keyboard

❌ **Social Engineering**
- Phishing for master passphrase
- Tricking user into revealing passwords
- Impersonation attacks

❌ **Weak Master Passphrase**
- Short passphrases (< 12 characters)
- Common words or patterns
- Reused passphrases from other services

### Assumptions

OpenPasswd's security relies on:

1. **Secure Operating System**
   - OS is not compromised
   - No malware with elevated privileges
   - File permissions are enforced

2. **Strong Master Passphrase**
   - User chooses a strong, unique passphrase
   - Passphrase is not written down insecurely
   - Passphrase is not reused

3. **Physical Security**
   - Computer is physically secure
   - No unauthorized physical access
   - Disk encryption is enabled (recommended)

4. **Trusted Execution Environment**
   - Binary is authentic (not tampered)
   - Go runtime is secure
   - Crypto libraries are not compromised

## Security Features

### 1. Zero-Knowledge Architecture

**What it means:**
- Your data never leaves your device
- No cloud sync, no remote servers
- No telemetry or analytics
- Complete offline operation

**Benefits:**
- No remote attack surface
- No data breaches from server compromises
- No government surveillance via service providers
- Full control over your data

### 2. Encrypted Recovery Keys

**How it works:**
- Recovery key generated during initialization
- Encrypted with master passphrase before storage
- Hash stored separately for verification
- Can restore access if passphrase is forgotten

**Security:**
- Recovery key itself is 256 bits of entropy
- Encrypted storage prevents unauthorized use
- User must write down plaintext copy
- Hash allows verification without decryption

### 3. No Plaintext Storage

**Previous versions** stored master passphrase on disk (bad idea!). This has been removed.

**Current behavior:**
- Master passphrase never stored
- User must enter passphrase each time
- Passphrase only exists in memory during use
- Memory is cleared after use (best effort)

### 4. TOTP Multi-Factor Authentication

**How it works:**
- TOTP secret generated during setup
- Stored encrypted in config directory
- User scans QR code with authenticator app
- 6-digit code required in addition to passphrase

**Security:**
- Protects against stolen database + passphrase
- Requires physical access to authenticator device
- Time-based codes expire every 30 seconds
- Standard TOTP algorithm (RFC 6238)

### 5. Database Integrity Verification

**How it works:**
- AES-GCM provides authenticated encryption
- Authentication tag verified on decryption
- Detects any tampering or corruption

**Benefits:**
- Prevents malicious modifications
- Detects accidental corruption
- Fails safely (won't decrypt corrupted data)

### 6. Secure File Permissions

**Enforced permissions:**
- `passwords.db`: 0600 (owner read/write only)
- `salt`: 0600
- `recovery_key`: 0600
- `totp_secret`: 0600
- `config.toml`: 0644 (not sensitive)

**Benefits:**
- Other users can't read your passwords
- Prevents accidental exposure
- Standard Unix security model

### 7. Clipboard Auto-Clear

**How it works:**
- Password copied to clipboard
- Timer starts (45 seconds)
- Clipboard automatically cleared

**Benefits:**
- Prevents passwords lingering in clipboard
- Reduces exposure window
- Automatic, no user action required

### 8. KDF Versioning and Migration

**How it works:**
- KDF version stored in config
- Supports multiple KDF algorithms
- Migration command re-encrypts with newer KDF
- Backward compatible with old databases

**Benefits:**
- Can upgrade security without losing data
- Smooth transition to stronger algorithms
- Future-proof design

## Known Limitations

### 1. Memory Security

**Issue:** Decrypted passwords exist in memory while in use.

**Risk:** Memory dumps or swap files could expose passwords.

**Mitigation:**
- Minimize time passwords are in memory
- Clear sensitive data when done (best effort)
- Use disk encryption to protect swap
- Lock screen when away from computer

**Future:** Implement secure memory pages (mlock) and memory wiping.

### 2. Terminal Emulator Security

**Issue:** Terminal emulators may log output or store scrollback.

**Risk:** Passwords could be captured in terminal logs.

**Mitigation:**
- Passwords are hidden by default
- Use secure terminal emulators
- Disable terminal logging
- Clear scrollback after use

**Future:** Investigate terminal security best practices.

### 3. Clipboard Security

**Issue:** Clipboard is accessible to all applications.

**Risk:** Malware could read clipboard contents.

**Mitigation:**
- Auto-clear after 45 seconds
- Minimize clipboard usage
- Use paste immediately
- Consider using keyboard typing instead

**Future:** Explore secure clipboard alternatives.

### 4. No Hardware Security Module (HSM) Support

**Issue:** Encryption keys are derived in software.

**Risk:** Software-based key derivation is slower than hardware.

**Mitigation:**
- Use strong KDF (Argon2id)
- High iteration counts
- Memory-hard algorithms

**Future:** Consider TPM/Secure Enclave integration.

### 5. Single-Device Only

**Issue:** No built-in sync across devices.

**Risk:** Manual sync could expose database during transfer.

**Mitigation:**
- Use encrypted file transfer
- Encrypt backups
- Use secure cloud storage if needed

**Future:** End-to-end encrypted sync (maybe).

## Security Audit Status

### Current Status: Pre-Alpha, Not Audited

OpenPasswd has **not** undergone a formal security audit. While we use industry-standard cryptography and follow best practices, there may be implementation bugs or design flaws.

**What this means:**
- Use at your own risk
- Not recommended for critical production data
- Suitable for personal use and testing
- Community review is welcome

### Planned Audits

Before v1.0 release, we plan to:
1. Complete internal security review
2. Engage professional security auditors
3. Conduct penetration testing
4. Implement bug bounty program
5. Publish audit results publicly

### Community Review

We welcome security researchers to review OpenPasswd:
- Code is fully open source
- Report vulnerabilities via GitHub Security Advisories
- Responsible disclosure appreciated
- Credit given for valid findings

## Best Practices

### For Users

1. **Use a strong master passphrase** (16+ characters)
2. **Enable TOTP** for additional security
3. **Write down recovery key** and store safely
4. **Enable disk encryption** on your device
5. **Keep OpenPasswd updated** (run `openpasswd upgrade`)
6. **Backup regularly** to encrypted storage
7. **Lock your screen** when away
8. **Use unique passwords** for each service

### For Developers

1. **Never store secrets in code** (use environment variables)
2. **Clear sensitive data** after use
3. **Use constant-time comparisons** for secrets
4. **Validate all inputs** before processing
5. **Follow Go security best practices**
6. **Keep dependencies updated**
7. **Run security linters** (gosec, etc.)
8. **Write security tests**

## Reporting Security Issues

Found a security vulnerability? Please report it responsibly:

1. **Do NOT** open a public GitHub issue
2. **Use** GitHub Security Advisories
3. **Email** security@openpasswd.dev (coming soon)
4. **Include** detailed reproduction steps
5. **Allow** reasonable time for fix before disclosure

We take security seriously and will respond promptly to valid reports.

## References

- [NIST SP 800-132](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-132.pdf) - PBKDF Recommendations
- [RFC 9106](https://www.rfc-editor.org/rfc/rfc9106.html) - Argon2 Specification
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [RFC 6238](https://www.rfc-editor.org/rfc/rfc6238) - TOTP Specification
- [NIST FIPS 197](https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.197.pdf) - AES Specification
- [BLAKE2 Paper](https://www.blake2.net/blake2.pdf)

## Conclusion

OpenPasswd uses well-established cryptographic algorithms and follows security best practices. However, as a pre-alpha project, it hasn't been formally audited. Use it for personal projects and testing, but wait for a stable release and security audit before trusting it with critical data.

Security is an ongoing process. We continuously review and improve OpenPasswd's security posture. Feedback and contributions are welcome!
