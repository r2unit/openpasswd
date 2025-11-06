#!/usr/bin/env bash

set -e

BINARY_NAME="openpasswd"
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
echo -e "${COLOR_BLUE}║                OpenPasswd Uninstaller                        ║${COLOR_RESET}"
echo -e "${COLOR_BLUE}║                                                              ║${COLOR_RESET}"
echo -e "${COLOR_BLUE}╚══════════════════════════════════════════════════════════════╝${COLOR_RESET}"
echo ""

if [ "$EUID" -ne 0 ]; then
    echo -e "${COLOR_YELLOW}⚠  This script requires sudo privileges${COLOR_RESET}"
    echo ""
fi

echo -e "${COLOR_YELLOW}This will remove OpenPasswd from your system.${COLOR_RESET}"
echo -e "${COLOR_YELLOW}Your password data in ~/.config/openpasswd will NOT be deleted.${COLOR_RESET}"
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Uninstall cancelled"
    exit 0
fi

echo ""
echo -e "${COLOR_BLUE}[1/3]${COLOR_RESET} Removing binary..."
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    sudo rm -f "$INSTALL_DIR/$BINARY_NAME"
    echo -e "${COLOR_GREEN}✓ Removed $INSTALL_DIR/$BINARY_NAME${COLOR_RESET}"
else
    echo -e "${COLOR_YELLOW}⚠  Binary not found${COLOR_RESET}"
fi

if [ -L "$INSTALL_DIR/openpass" ]; then
    sudo rm -f "$INSTALL_DIR/openpass"
    echo -e "${COLOR_GREEN}✓ Removed $INSTALL_DIR/openpass${COLOR_RESET}"
fi

if [ -L "$INSTALL_DIR/pw" ]; then
    sudo rm -f "$INSTALL_DIR/pw"
    echo -e "${COLOR_GREEN}✓ Removed $INSTALL_DIR/pw${COLOR_RESET}"
fi

echo ""
echo -e "${COLOR_BLUE}[2/3]${COLOR_RESET} Removing bash completion..."
if [ -f "$COMPLETION_DIR_BASH/openpass" ]; then
    sudo rm -f "$COMPLETION_DIR_BASH/openpass"
    echo -e "${COLOR_GREEN}✓ Removed bash completion${COLOR_RESET}"
else
    echo -e "${COLOR_YELLOW}⚠  Bash completion not found${COLOR_RESET}"
fi

echo ""
echo -e "${COLOR_BLUE}[3/3]${COLOR_RESET} Removing zsh completion..."
if [ -f "$COMPLETION_DIR_ZSH/_openpass" ]; then
    sudo rm -f "$COMPLETION_DIR_ZSH/_openpass"
    echo -e "${COLOR_GREEN}✓ Removed zsh completion${COLOR_RESET}"
else
    echo -e "${COLOR_YELLOW}⚠  Zsh completion not found${COLOR_RESET}"
fi

echo ""
echo -e "${COLOR_GREEN}╔══════════════════════════════════════════════════════════════╗${COLOR_RESET}"
echo -e "${COLOR_GREEN}║                                                              ║${COLOR_RESET}"
echo -e "${COLOR_GREEN}║              OpenPasswd Uninstalled Successfully             ║${COLOR_RESET}"
echo -e "${COLOR_GREEN}║                                                              ║${COLOR_RESET}"
echo -e "${COLOR_GREEN}╚══════════════════════════════════════════════════════════════╝${COLOR_RESET}"
echo ""
echo -e "${COLOR_BLUE}Your password data is preserved in:${COLOR_RESET}"
echo -e "  ${COLOR_GREEN}~/.config/openpasswd/${COLOR_RESET}"
echo ""
echo -e "${COLOR_YELLOW}To completely remove all data:${COLOR_RESET}"
echo -e "  ${COLOR_RED}rm -rf ~/.config/openpasswd${COLOR_RESET}"
echo ""
