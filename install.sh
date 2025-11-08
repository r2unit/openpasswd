#!/usr/bin/env bash
#
# OpenPasswd Installation Script
# Install with: curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/master/install.sh | bash
# Or clone and run: ./install.sh

set -e

BINARY_NAME="openpasswd"
INSTALL_DIR="/usr/local/bin"
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="/usr/local/share/zsh/site-functions"
REPO_URL="https://github.com/r2unit/openpasswd"
REPO_API_URL="https://api.github.com/repos/r2unit/openpasswd"
TEMP_DIR=""

COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

# Cleanup function
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        echo -e "${COLOR_BLUE}Cleaning up...${COLOR_RESET}"
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
            echo -e "${COLOR_RED}✗ Unsupported operating system: $OS${COLOR_RESET}"
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
            echo -e "${COLOR_RED}✗ Unsupported architecture: $ARCH${COLOR_RESET}"
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
    echo -e "${COLOR_BLUE}Fetching latest release information...${COLOR_RESET}"
    
    RELEASE_JSON=$(curl -sL "${REPO_API_URL}/releases/latest")
    
    if echo "$RELEASE_JSON" | grep -q "Not Found"; then
        echo -e "${COLOR_YELLOW}⚠  No releases found. Building from source...${COLOR_RESET}"
        return 1
    fi
    
    LATEST_VERSION=$(echo "$RELEASE_JSON" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    DOWNLOAD_URL="${REPO_URL}/releases/download/${LATEST_VERSION}/${PLATFORM_BINARY}"
    CHECKSUM_URL="${REPO_URL}/releases/download/${LATEST_VERSION}/${PLATFORM_BINARY}.sha256"
    
    if [ -z "$LATEST_VERSION" ]; then
        echo -e "${COLOR_YELLOW}⚠  Could not determine latest version. Building from source...${COLOR_RESET}"
        return 1
    fi
    
    echo -e "${COLOR_GREEN}Latest version: ${LATEST_VERSION}${COLOR_RESET}"
    return 0
}

# Download pre-built binary
download_binary() {
    TEMP_DIR=$(mktemp -d)
    echo -e "${COLOR_BLUE}Downloading ${PLATFORM_BINARY}...${COLOR_RESET}"
    
    if ! curl -sSL -f "$DOWNLOAD_URL" -o "${TEMP_DIR}/${BINARY_NAME}"; then
        echo -e "${COLOR_YELLOW}⚠  Failed to download pre-built binary${COLOR_RESET}"
        return 1
    fi
    
    # Download and verify checksum if available
    if curl -sSL -f "$CHECKSUM_URL" -o "${TEMP_DIR}/${BINARY_NAME}.sha256" 2>/dev/null; then
        echo -e "${COLOR_BLUE}Verifying checksum...${COLOR_RESET}"
        cd "${TEMP_DIR}"
        if sha256sum -c "${BINARY_NAME}.sha256" 2>/dev/null; then
            echo -e "${COLOR_GREEN}✓ Checksum verified${COLOR_RESET}"
        else
            echo -e "${COLOR_YELLOW}⚠  Checksum verification failed${COLOR_RESET}"
        fi
        cd - > /dev/null
    fi
    
    chmod +x "${TEMP_DIR}/${BINARY_NAME}"
    echo -e "${COLOR_GREEN}✓ Download complete${COLOR_RESET}"
    return 0
}

# Build from source
build_from_source() {
    echo -e "${COLOR_BLUE}Building from source...${COLOR_RESET}"
    
    if ! command -v go &> /dev/null; then
        echo -e "${COLOR_RED}✗ Error: Go is not installed${COLOR_RESET}"
        echo -e "${COLOR_YELLOW}  Please install Go from https://golang.org/dl/${COLOR_RESET}"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        echo -e "${COLOR_RED}✗ Error: Git is not installed${COLOR_RESET}"
        echo -e "${COLOR_YELLOW}  Please install Git and try again${COLOR_RESET}"
        exit 1
    fi
    
    TEMP_DIR=$(mktemp -d)
    echo -e "${COLOR_BLUE}Cloning repository...${COLOR_RESET}"
    if ! git clone --depth 1 "$REPO_URL" "$TEMP_DIR" 2>&1 | grep -v "Cloning into" || true; then
        echo -e "${COLOR_RED}✗ Failed to clone repository${COLOR_RESET}"
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
    
    echo -e "${COLOR_BLUE}Building version ${VERSION} (commit: ${GIT_COMMIT})...${COLOR_RESET}"
    
    # Build in the temp directory
    if (cd "$TEMP_DIR" && go build -ldflags="-X 'github.com/r2unit/openpasswd/pkg/version.Version=${VERSION}' \
                        -X 'github.com/r2unit/openpasswd/pkg/version.GitCommit=${GIT_COMMIT}' \
                        -X 'github.com/r2unit/openpasswd/pkg/version.BuildDate=${BUILD_DATE}'" \
             -o "$BINARY_NAME" ./cmd/client); then
        echo -e "${COLOR_GREEN}✓ Build successful${COLOR_RESET}"
    else
        echo -e "${COLOR_RED}✗ Build failed${COLOR_RESET}"
        exit 1
    fi
}

echo -e "${COLOR_BLUE}╔══════════════════════════════════════════════════════════════╗${COLOR_RESET}"
echo -e "${COLOR_BLUE}║                                                              ║${COLOR_RESET}"
echo -e "${COLOR_BLUE}║                  OpenPasswd Installer                        ║${COLOR_RESET}"
echo -e "${COLOR_BLUE}║          Secure Terminal Password Manager                    ║${COLOR_RESET}"
echo -e "${COLOR_BLUE}║                                                              ║${COLOR_RESET}"
echo -e "${COLOR_BLUE}╚══════════════════════════════════════════════════════════════╝${COLOR_RESET}"
echo ""

if [ "$EUID" -ne 0 ]; then
    echo -e "${COLOR_YELLOW}⚠  This script requires sudo privileges to install to $INSTALL_DIR${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}   You will be prompted for your password...${COLOR_RESET}"
    echo ""
fi

echo -e "${COLOR_BLUE}[1/4]${COLOR_RESET} Detecting platform..."
detect_platform
echo -e "${COLOR_GREEN}✓ Platform: ${OS}/${ARCH}${COLOR_RESET}"

echo ""
echo -e "${COLOR_BLUE}[2/4]${COLOR_RESET} Getting OpenPasswd binary..."

# Try to download pre-built binary first
if get_latest_version && download_binary; then
    echo -e "${COLOR_GREEN}✓ Using pre-built binary${COLOR_RESET}"
else
    # Fall back to building from source
    build_from_source
fi

echo ""
echo -e "${COLOR_BLUE}[3/4]${COLOR_RESET} Installing binary to $INSTALL_DIR..."
sudo cp "${TEMP_DIR}/${BINARY_NAME}" "$INSTALL_DIR/$BINARY_NAME"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo -e "${COLOR_GREEN}✓ Installed: $INSTALL_DIR/$BINARY_NAME${COLOR_RESET}"

# Create aliases
alias_created=0
if sudo ln -sf "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/openpass" 2>/dev/null; then
    echo -e "${COLOR_GREEN}✓ Created alias: $INSTALL_DIR/openpass${COLOR_RESET}"
    alias_created=1
else
    echo -e "${COLOR_YELLOW}⚠  Could not create 'openpass' alias${COLOR_RESET}"
fi

if sudo ln -sf "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/pw" 2>/dev/null; then
    echo -e "${COLOR_GREEN}✓ Created alias: $INSTALL_DIR/pw${COLOR_RESET}"
    alias_created=1
else
    echo -e "${COLOR_YELLOW}⚠  Could not create 'pw' alias${COLOR_RESET}"
fi

echo ""
echo -e "${COLOR_BLUE}[4/4]${COLOR_RESET} Installing shell completions..."

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
            COMPREPLY=( $(compgen -W "--verbose --check" -- ${cur}) )
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
    echo -e "${COLOR_GREEN}✓ Bash completion installed${COLOR_RESET}"
else
    echo -e "${COLOR_YELLOW}⚠  Bash completion directory not found${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}   Manual install: sudo cp /tmp/openpass-completion.bash /etc/bash_completion.d/openpass${COLOR_RESET}"
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
    echo -e "${COLOR_GREEN}✓ Zsh completion installed${COLOR_RESET}"
else
    echo -e "${COLOR_YELLOW}⚠  Zsh completion directory not found${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}   Manual install: sudo cp /tmp/_openpass $COMPLETION_DIR_ZSH/_openpass${COLOR_RESET}"
fi

echo ""
echo -e "${COLOR_BLUE}Verifying installation...${COLOR_RESET}"
if command -v openpasswd &> /dev/null; then
    version=$(openpasswd version 2>/dev/null || echo "unknown")
    echo -e "${COLOR_GREEN}✓ OpenPasswd installed successfully!${COLOR_RESET}"
    echo -e "${COLOR_GREEN}  Version: ${version}${COLOR_RESET}"
    echo ""
else
    echo -e "${COLOR_RED}✗ Installation verification failed${COLOR_RESET}"
    exit 1
fi

echo -e "${COLOR_GREEN}╔══════════════════════════════════════════════════════════════╗${COLOR_RESET}"
echo -e "${COLOR_GREEN}║                                                              ║${COLOR_RESET}"
echo -e "${COLOR_GREEN}║                   Installation Complete!                     ║${COLOR_RESET}"
echo -e "${COLOR_GREEN}║                                                              ║${COLOR_RESET}"
echo -e "${COLOR_GREEN}╚══════════════════════════════════════════════════════════════╝${COLOR_RESET}"
echo ""
echo -e "${COLOR_BLUE}Quick Start:${COLOR_RESET}"
echo -e "  ${COLOR_GREEN}openpasswd init${COLOR_RESET}              Initialize the password manager"
echo -e "  ${COLOR_GREEN}openpasswd add${COLOR_RESET}               Add a new password"
echo -e "  ${COLOR_GREEN}openpasswd list${COLOR_RESET}              List all passwords"
echo -e "  ${COLOR_GREEN}openpasswd settings${COLOR_RESET}          Configure MFA/passphrase"
echo -e "  ${COLOR_GREEN}openpasswd version --check${COLOR_RESET}   Check for updates"
echo -e "  ${COLOR_GREEN}openpasswd upgrade${COLOR_RESET}           Upgrade to latest version"
echo -e "  ${COLOR_GREEN}openpasswd help${COLOR_RESET}              Show all commands"
echo ""
echo -e "${COLOR_BLUE}Aliases:${COLOR_RESET}"
echo -e "  ${COLOR_GREEN}openpass${COLOR_RESET} = ${COLOR_GREEN}openpasswd${COLOR_RESET}    Short alias"
echo -e "  ${COLOR_GREEN}pw${COLOR_RESET} = ${COLOR_GREEN}openpasswd${COLOR_RESET}           Ultra-short alias"
echo ""
echo -e "${COLOR_YELLOW}Note:${COLOR_RESET} Restart your shell or run ${COLOR_GREEN}source ~/.bashrc${COLOR_RESET} to enable completions"
echo ""
echo -e "${COLOR_BLUE}Documentation:${COLOR_RESET} ${REPO_URL}"
echo ""
