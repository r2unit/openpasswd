# OpenPasswd

A terminal-based password manager that keeps your data local and encrypted.

[![CI/CD Pipeline](https://github.com/r2unit/openpasswd/actions/workflows/ci.yml/badge.svg)](https://github.com/r2unit/openpasswd/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/r2unit/openpasswd)](https://github.com/r2unit/openpasswd/releases/latest)

> [!WARNING]
> **Pre-alpha software.** Don't use this for anything important yet.

## What is this?

OpenPasswd is a password manager for people who live in the terminal. It stores everything locally with AES-256 encryption, supports TOTP for 2FA, and doesn't phone home. Ever.

**What you can store:**
- Logins (username, password, URL)
- Credit cards
- Secure notes
- Identity info
- Whatever else you need encrypted

## Install

**Quick install:**
```bash
curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
```

**Or grab a binary:** [releases page](https://github.com/r2unit/openpasswd/releases/latest)

**Or build it yourself:**
```bash
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd
make install
```

## Usage

```bash
openpasswd init              # First-time setup
openpasswd add               # Add a password
openpasswd list              # View your passwords
openpasswd settings set-totp # Enable 2FA
```

That's it. The TUI will guide you through the rest.

## Security stuff

- **AES-256-GCM** encryption
- **Argon2id** key derivation
- **BLAKE2b** integrity checks
- **Local only** - no cloud, no sync, no tracking
- **Zero-knowledge** - your data stays on your machine

Since this is pre-alpha, there's no security audit yet. Use at your own risk.

## Why this exists

Most password managers want you to trust their cloud. OpenPasswd doesn't have a cloud. It's just you, your terminal, and an encrypted database on your disk.

If you like `pass` or `gopass` but want something that doesn't need GPG and has a nicer interface, this might be for you.

## Contributing

Found a bug? Want to add something? PRs are welcome.

1. Fork it
2. Make your changes
3. Run `go test ./...` and `go fmt ./...`
4. Send a PR

## License

MIT - see [LICENSE](LICENSE)

---

Made by [@r2unit](https://github.com/r2unit)
