#!/usr/bin/env bash
#
# OpenPasswd Installation Script
# Install with: curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
# Or clone and run: ./install.sh
# Install specific branch: ./install.sh --branch devel

set -e

BINARY_NAME="openpasswd"
INSTALL_DIR="/usr/local/bin"
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="/usr/local/share/zsh/site-functions"
REPO_URL="https://github.com/r2unit/openpasswd"
REPO_API_URL="https://api.github.com/repos/r2unit/openpasswd"
TEMP_DIR=""
BRANCH="main"  # Default branch

# Cleanup function
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        echo "Cleaning up temporary files..."
        rm -rf "$TEMP_DIR"
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            echo "Error: Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            echo "Error: Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    if [ "$OS" = "windows" ]; then
        BINARY_EXT=".exe"
    else
        BINARY_EXT=""
    fi
    
    PLATFORM_BINARY="openpasswd-${OS}-${ARCH}${BINARY_EXT}"
}

# Get latest release version from GitHub
get_latest_version() {
    echo "Fetching latest release information..."
    
    RELEASE_JSON=$(curl -sL "${REPO_API_URL}/releases/latest")
    
    if echo "$RELEASE_JSON" | grep -q "Not Found"; then
        echo "No releases found. Will build from source..."
        return 1
    fi
    
    LATEST_VERSION=$(echo "$RELEASE_JSON" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    DOWNLOAD_URL="${REPO_URL}/releases/download/${LATEST_VERSION}/${PLATFORM_BINARY}"
    CHECKSUM_URL="${REPO_URL}/releases/download/${LATEST_VERSION}/${PLATFORM_BINARY}.sha256"
    
    if [ -z "$LATEST_VERSION" ]; then
        echo "Could not determine latest version. Will build from source..."
        return 1
    fi
    
    echo "Latest version: ${LATEST_VERSION}"
    return 0
}

# Download pre-built binary
download_binary() {
    TEMP_DIR=$(mktemp -d)
    echo "Downloading ${PLATFORM_BINARY}..."
    
    if ! curl -sSL -f "$DOWNLOAD_URL" -o "${TEMP_DIR}/${BINARY_NAME}"; then
        echo "Failed to download pre-built binary"
        echo "This usually means the binary for your platform hasn't been released yet"
        return 1
    fi
    
    # Download and verify checksum if available
    if curl -sSL -f "$CHECKSUM_URL" -o "${TEMP_DIR}/${BINARY_NAME}.sha256" 2>/dev/null; then
        echo "Verifying checksum..."
        cd "${TEMP_DIR}"
        if sha256sum -c "${BINARY_NAME}.sha256" 2>/dev/null; then
            echo "Checksum verified"
        else
            echo "Warning: Checksum verification failed"
        fi
        cd - > /dev/null
    fi
    
    chmod +x "${TEMP_DIR}/${BINARY_NAME}"
    echo "Download complete"
    return 0
}

# Build from source
build_from_source() {
    echo "Building from source (branch: ${BRANCH})..."
    
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed"
        echo "Please install Go from https://golang.org/dl/"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        echo "Error: Git is not installed"
        echo "Please install Git and try again"
        exit 1
    fi
    
    TEMP_DIR=$(mktemp -d)
    echo "Cloning repository (branch: ${BRANCH})..."
    if ! git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$TEMP_DIR" >/dev/null 2>&1; then
        echo "Failed to clone repository from branch '${BRANCH}'"
        echo "Make sure the branch exists"
        exit 1
    fi
    
    # Get version info for build
    if [ -f "$TEMP_DIR/pkg/version/version.go" ]; then
        VERSION=$(grep 'Version = ' "$TEMP_DIR/pkg/version/version.go" | sed 's/.*"\(.*\)".*/\1/' || echo "dev")
    else
        VERSION="dev"
    fi
    GIT_COMMIT=$(cd "$TEMP_DIR" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
    
    echo "Building version ${VERSION} (commit: ${GIT_COMMIT})..."
    
    # Build in the temp directory
    if (cd "$TEMP_DIR" && go build -ldflags="-X 'github.com/r2unit/openpasswd/pkg/version.Version=${VERSION}' \
                        -X 'github.com/r2unit/openpasswd/pkg/version.GitCommit=${GIT_COMMIT}' \
                        -X 'github.com/r2unit/openpasswd/pkg/version.BuildDate=${BUILD_DATE}'" \
             -o "$BINARY_NAME" ./cmd/client); then
        echo "Build successful"
    else
        echo "Build failed"
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --branch)
            if [ -z "$2" ]; then
                echo "Error: --branch requires a branch name"
                echo "Usage: $0 --branch <branch-name>"
                echo "Example: $0 --branch devel"
                exit 1
            fi
            BRANCH="$2"
            shift 2
            ;;
        -h|--help)
            echo "OpenPasswd Installer"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --branch <name>    Install from specific branch (default: main)"
            echo "  -h, --help         Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                 # Install from main branch"
            echo "  $0 --branch devel  # Install from devel branch"
            echo "  $0 --branch main   # Install from main branch (explicit)"
            echo ""
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Run '$0 --help' for usage information"
            exit 1
            ;;
    esac
done

echo "OpenPasswd Installer"
echo "===================="
echo ""

if [ "$BRANCH" != "main" ]; then
    echo "WARNING: Installing from branch: ${BRANCH}"
    echo "         (This may be an unstable development branch)"
    echo ""
fi

if [ "$EUID" -ne 0 ]; then
    echo "Note: This script requires sudo privileges to install to $INSTALL_DIR"
    echo "You will be prompted for your password..."
    echo ""
fi

echo "[1/4] Detecting platform..."
detect_platform
echo "Platform: ${OS}/${ARCH}"

echo ""
echo "[2/4] Getting OpenPasswd binary..."

# If a specific branch is requested, always build from source
if [ "$BRANCH" != "main" ]; then
    echo "Building from branch '${BRANCH}' (pre-built binaries only available for main branch)"
    build_from_source
else
    # Try to download pre-built binary first
    if get_latest_version && download_binary; then
        echo "Using pre-built binary"
    else
        # Fall back to building from source
        build_from_source
    fi
fi

echo ""
echo "[3/4] Installing binary to $INSTALL_DIR..."
sudo cp "${TEMP_DIR}/${BINARY_NAME}" "$INSTALL_DIR/$BINARY_NAME"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "Installed: $INSTALL_DIR/$BINARY_NAME"

# Create aliases
if sudo ln -sf "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/openpass" 2>/dev/null; then
    echo "Created alias: $INSTALL_DIR/openpass"
else
    echo "Warning: Could not create 'openpass' alias"
fi

if sudo ln -sf "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/pw" 2>/dev/null; then
    echo "Created alias: $INSTALL_DIR/pw"
else
    echo "Warning: Could not create 'pw' alias"
fi

echo ""
echo "[4/4] Installing shell completions..."

cat > /tmp/openpass-completion.bash << 'EOF'
_openpass_completions()
{
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="init add list settings version upgrade help"
    
    case "${prev}" in
        openpass|openpasswd|pw)
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
            return 0
            ;;
        add)
            COMPREPLY=( $(compgen -W "login card note identity password other" -- ${cur}) )
            return 0
            ;;
        settings)
            COMPREPLY=( $(compgen -W "set-passphrase remove-passphrase set-totp remove-totp show-totp-qr set-yubikey remove-yubikey help" -- ${cur}) )
            return 0
            ;;
        version)
            COMPREPLY=( $(compgen -W "--verbose --check --disable-checking --enable-checking" -- ${cur}) )
            return 0
            ;;
    esac
}

complete -F _openpass_completions openpass
complete -F _openpass_completions openpasswd
complete -F _openpass_completions pw
EOF

if [ -d "$COMPLETION_DIR_BASH" ]; then
    sudo cp /tmp/openpass-completion.bash "$COMPLETION_DIR_BASH/openpass"
    echo "Bash completion installed"
else
    echo "Warning: Bash completion directory not found"
    echo "Manual install: sudo cp /tmp/openpass-completion.bash /etc/bash_completion.d/openpass"
fi

cat > /tmp/_openpass << 'EOF'
#compdef openpass openpasswd pw

_openpass() {
    local -a commands
    commands=(
        'init:Initialize configuration and database'
        'add:Add a new password entry'
        'list:List and search passwords'
        'settings:Manage settings'
        'version:Show version information'
        'upgrade:Upgrade to latest version'
        'help:Show help message'
    )

    local -a add_types
    add_types=(
        'login:Login credentials'
        'card:Credit/debit card'
        'note:Secure note'
        'identity:Personal identity'
        'password:Simple password'
        'other:Other credential'
    )

    local -a settings_commands
    settings_commands=(
        'set-passphrase:Set master passphrase'
        'remove-passphrase:Remove master passphrase'
        'set-totp:Enable TOTP authentication'
        'remove-totp:Disable TOTP authentication'
        'show-totp-qr:Show TOTP QR code'
        'set-yubikey:Enable YubiKey authentication'
        'remove-yubikey:Disable YubiKey authentication'
        'help:Show settings help'
    )

    local -a version_flags
    version_flags=(
        '--verbose:Show detailed version information'
        '--check:Check for updates'
        '--disable-checking:Disable automatic update checks'
        '--enable-checking:Enable automatic update checks'
    )

    case $words[2] in
        add)
            _describe 'password types' add_types
            ;;
        settings)
            _describe 'settings commands' settings_commands
            ;;
        version)
            _describe 'version options' version_flags
            ;;
        *)
            _describe 'commands' commands
            ;;
    esac
}

compdef _openpass openpass
compdef _openpass openpasswd
compdef _openpass pw
EOF

if [ -d "$COMPLETION_DIR_ZSH" ]; then
    sudo mkdir -p "$COMPLETION_DIR_ZSH"
    sudo cp /tmp/_openpass "$COMPLETION_DIR_ZSH/_openpass"
    echo "Zsh completion installed"
else
    echo "Warning: Zsh completion directory not found"
    echo "Manual install: sudo cp /tmp/_openpass $COMPLETION_DIR_ZSH/_openpass"
fi

echo ""
echo "Verifying installation..."
if command -v openpasswd &> /dev/null; then
    version=$(openpasswd version 2>/dev/null || echo "unknown")
    echo "OpenPasswd installed successfully!"
    echo "Version: ${version}"
    echo ""
else
    echo "Installation verification failed"
    exit 1
fi

echo "Installation Complete!"
echo "====================="
echo ""
echo "Quick Start:"
echo "  openpasswd init              Initialize the password manager"
echo "  openpasswd add               Add a new password"
echo "  openpasswd list              List all passwords"
echo "  openpasswd settings          Configure MFA/passphrase"
echo "  openpasswd version --check   Check for updates"
echo "  openpasswd upgrade           Upgrade to latest version"
echo "  openpasswd help              Show all commands"
echo ""
echo "Aliases:"
echo "  openpass = openpasswd    Short alias"
echo "  pw = openpasswd          Ultra-short alias"
echo ""
echo "Note: Restart your shell or run 'source ~/.bashrc' to enable completions"
echo ""
echo "Documentation: ${REPO_URL}"
echo ""
