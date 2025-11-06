#!/usr/bin/env bash

set -e

BINARY_NAME="openpass"
INSTALL_DIR="/usr/local/bin"
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="/usr/local/share/zsh/site-functions"

COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

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

echo -e "${COLOR_BLUE}[1/4]${COLOR_RESET} Building OpenPasswd..."
if ! command -v go &> /dev/null; then
    echo -e "${COLOR_RED}✗ Error: Go is not installed${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}  Please install Go from https://golang.org/dl/${COLOR_RESET}"
    exit 1
fi

go build -o "$BINARY_NAME" cmd/openpass/main.go cmd/openpass/terminal.go
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
if sudo ln -sf "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/openpasswd" 2>/dev/null; then
    echo -e "${COLOR_GREEN}✓ Created alias: $INSTALL_DIR/openpasswd${COLOR_RESET}"
    alias_created=1
else
    echo -e "${COLOR_YELLOW}⚠  Could not create 'openpasswd' alias${COLOR_RESET}"
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

    commands="init add list settings server auth help"
    
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
        auth)
            COMPREPLY=( $(compgen -W "login logout" -- ${cur}) )
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
        'server:Start headless server mode'
        'auth:Authentication commands'
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

    local -a auth_commands
    auth_commands=(
        'login:Login to remote server'
        'logout:Logout from remote server'
    )

    case $words[2] in
        add)
            _describe 'password types' add_types
            ;;
        settings)
            _describe 'settings commands' settings_commands
            ;;
        auth)
            _describe 'auth commands' auth_commands
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
if command -v openpass &> /dev/null; then
    version=$(openpass help | grep -i openpasswd | head -1)
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
echo -e "  ${COLOR_GREEN}openpass init${COLOR_RESET}              Initialize the password manager"
echo -e "  ${COLOR_GREEN}openpass add${COLOR_RESET}               Add a new password"
echo -e "  ${COLOR_GREEN}openpass list${COLOR_RESET}              List all passwords"
echo -e "  ${COLOR_GREEN}openpass settings${COLOR_RESET}          Configure MFA/passphrase"
echo -e "  ${COLOR_GREEN}openpass help${COLOR_RESET}              Show all commands"
echo ""
echo -e "${COLOR_BLUE}Aliases:${COLOR_RESET}"
echo -e "  ${COLOR_GREEN}openpasswd${COLOR_RESET} = ${COLOR_GREEN}openpass${COLOR_RESET}    Full name alias"
echo -e "  ${COLOR_GREEN}pw${COLOR_RESET} = ${COLOR_GREEN}openpass${COLOR_RESET}           Short alias"
echo ""
echo -e "${COLOR_YELLOW}Note:${COLOR_RESET} Restart your shell or run ${COLOR_GREEN}source ~/.bashrc${COLOR_RESET} to enable completions"
echo ""
