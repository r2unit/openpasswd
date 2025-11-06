# OpenPasswd

A secure, terminal-based password manager built with Go. Store and manage your passwords locally or remotely with end-to-end encryption.

## Features

- **Terminal UI**: Beautiful interactive terminal interface
- **End-to-End Encryption**: AES-256-GCM encryption with PBKDF2 key derivation
- **Local Storage**: JSON-based database with encrypted passwords
- **Headless Server Mode**: Run as HTTP API server for remote access
- **Remote Client**: Connect to remote server instances
- **Zero Third-Party Dependencies**: Built with Go standard library only
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

```bash
# Clone the repository
git clone https://github.com/r2unit/openpasswd.git
cd openpasswd

# Build the binary
go build -o openpass ./cmd/openpass

# Move to your PATH (optional)
sudo mv openpass /usr/local/bin/
```

## Quick Start

### Local Usage

```bash
# Initialize configuration
openpass init

# Launch the TUI
openpass
```

### Server Mode (Headless)

```bash
# Set environment variables
export OPENPASS_MASTER_KEY="your-secure-master-key"
export OPENPASS_PORT="8080"

# Start the server
openpass server
```

### Remote Client

```bash
# Login to remote server
openpass auth login http://localhost:8080

# Use TUI with remote server
openpass
```

## Usage

### Commands

- `openpass` - Launch interactive TUI
- `openpass init` - Initialize configuration
- `openpass server` - Start headless server
- `openpass auth login <url>` - Login to remote server
- `openpass auth logout` - Logout from remote server
- `openpass help` - Show help message

### TUI Operations

1. **List Passwords** - View all stored passwords
2. **Add Password** - Add a new password entry
3. **View Password** - View password details (decrypted)
4. **Search Passwords** - Search by name, username, or URL
5. **Update Password** - Modify existing password
6. **Delete Password** - Remove password entry
7. **Exit** - Quit the application

## Docker

### Build Image

```bash
docker build -t openpasswd .
```

### Run Server in Container

```bash
docker run -d \
  -p 8080:8080 \
  -e OPENPASS_MASTER_KEY="your-secure-master-key" \
  -v /path/to/data:/root/.config/passwd \
  openpasswd
```

### Connect from Local Client

```bash
openpass auth login http://localhost:8080
```

## Configuration

### Local Configuration

Configuration is stored in `~/.config/passwd/`:

- `salt` - Encryption salt (base64 encoded)
- `passwords.db` - Encrypted password database (JSON)
- `token.json` - Authentication token for remote access

### Environment Variables

- `OPENPASS_MASTER_KEY` - Master key for server authentication
- `OPENPASS_PORT` - Server port (default: 8080)

## Security

- **Encryption**: All passwords are encrypted with AES-256-GCM
- **Key Derivation**: PBKDF2 with 100,000 iterations and SHA-256
- **Secure Storage**: Database files stored with 0600 permissions
- **Token-Based Auth**: JWT-like tokens with 24-hour expiration
- **No Third-Party Deps**: Reduces attack surface

## API Endpoints

When running in server mode:

- `POST /api/auth/login` - Authenticate and get token
- `POST /api/auth/logout` - Invalidate token
- `GET /api/passwords` - List all passwords
- `POST /api/passwords` - Create new password
- `GET /api/passwords/:id` - Get password by ID
- `PUT /api/passwords/:id` - Update password
- `DELETE /api/passwords/:id` - Delete password
- `GET /api/passwords/search?q=query` - Search passwords
- `GET /api/health` - Health check

## Development

### Project Structure

```
openpasswd/
├── cmd/
│   └── openpass/       # Main application
├── pkg/
│   ├── auth/           # Authentication
│   ├── client/         # HTTP client
│   ├── config/         # Configuration
│   ├── crypto/         # Encryption
│   ├── database/       # Database operations
│   ├── models/         # Data models
│   ├── server/         # HTTP server
│   └── tui/            # Terminal UI
├── Dockerfile
├── go.mod
└── README.md
```

### Building

```bash
# Build for current platform
go build -o openpass ./cmd/openpass

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o openpass-linux ./cmd/openpass

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o openpass-macos ./cmd/openpass

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o openpass.exe ./cmd/openpass
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

r2unit - https://github.com/r2unit
