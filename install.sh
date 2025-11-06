#!/usr/bin/env bash
#
# OpenPasswd Installation Script
# Install with: curl -sSL https://raw.githubusercontent.com/r2unit/openpasswd/main/install.sh | bash
# Or clone and run: ./install.sh

set -e

BINARY_NAME="openpasswd"
INSTALL_DIR="/usr/local/bin"
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="/usr/local/share/zsh/site-functions"
REPO_URL="https://github.com/r2unit/openpasswd"
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

# Check if we're in the repo directory or need to clone
if [ -f "./cmd/openpasswd/main.go" ]; then
    echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} Building from current directory..."
    BUILD_DIR="."
else
    echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} Cloning repository from GitHub..."
    
    if ! command -v git &> /dev/null; then
        echo -e "${COLOR_RED}✗ Error: Git is not installed${COLOR_RESET}"
        echo -e "${COLOR_YELLOW}  Please install Git and try again${COLOR_RESET}"
        exit 1
    fi
    
    TEMP_DIR=$(mktemp -d)
    echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} Cloning into $TEMP_DIR..."
    git clone --depth 1 "$REPO_URL" "$TEMP_DIR" 2>&1 | grep -v "Cloning into" || true
    BUILD_DIR="$TEMP_DIR"
fi

echo ""
echo -e "${COLOR_BLUE}[1/4]${COLOR_RESET} Building OpenPasswd..."
if ! command -v go &> /dev/null; then
    echo -e "${COLOR_RED}✗ Error: Go is not installed${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}  Please install Go from https://golang.org/dl/${COLOR_RESET}"
    exit 1
fi

cd "$BUILD_DIR"
go build -o "$BINARY_NAME" ./cmd/openpasswd
if [ $? -eq 0 ]; then
    echo -e "${COLOR_GREEN}✓ Build successful${COLOR_RESET}"
else
    echo -e "${COLOR_RED}✗ Build failed${COLOR_RESET}"
    exit 1
fi

echo ""
echo -e "${COLOR_BLUE}[2/4]${COLOR_RESET} Installing binary to $INSTALL_DIR..."
sudo cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
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
echo -e "${COLOR_BLUE}[3/4]${COLOR_RESET} Installing shell completions..."

cat > /tmp/openpass-completion.bash << 'EOF'
_openpass_completions()
{
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="init add list settings help"
    
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

    case $words[2] in
        add)
            _describe 'password types' add_types
            ;;
        settings)
            _describe 'settings commands' settings_commands
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
echo -e "${COLOR_BLUE}[4/4]${COLOR_RESET} Verifying installation..."
if command -v openpasswd &> /dev/null; then
    version=$(openpasswd help | grep -i openpasswd | head -1)
    echo -e "${COLOR_GREEN}✓ OpenPasswd installed successfully!${COLOR_RESET}"
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
