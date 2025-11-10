#!/usr/bin/env bash

set -e

CLIENT_BINARY="openpasswd"
SERVER_BINARY="openpasswd-server"
INSTALL_DIR="/usr/local/bin"
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="/usr/local/share/zsh/site-functions"

echo "OpenPasswd Uninstaller"
echo "======================"
echo ""

if [ "$EUID" -ne 0 ]; then
    echo "Note: This script requires sudo privileges"
    echo ""
fi

echo "This will remove OpenPasswd from your system."
echo "Your password data in ~/.config/openpasswd will NOT be deleted."
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Uninstall cancelled"
    exit 0
fi

echo ""
echo "[1/4] Removing binaries..."
if [ -f "$INSTALL_DIR/$CLIENT_BINARY" ]; then
    sudo rm -f "$INSTALL_DIR/$CLIENT_BINARY"
    echo "Removed $INSTALL_DIR/$CLIENT_BINARY"
else
    echo "Client binary not found"
fi

if [ -f "$INSTALL_DIR/$SERVER_BINARY" ]; then
    sudo rm -f "$INSTALL_DIR/$SERVER_BINARY"
    echo "Removed $INSTALL_DIR/$SERVER_BINARY"
else
    echo "Server binary not found"
fi

echo ""
echo "[2/4] Removing aliases..."
if [ -L "$INSTALL_DIR/openpass" ]; then
    sudo rm -f "$INSTALL_DIR/openpass"
    echo "Removed $INSTALL_DIR/openpass"
fi

if [ -L "$INSTALL_DIR/pw" ]; then
    sudo rm -f "$INSTALL_DIR/pw"
    echo "Removed $INSTALL_DIR/pw"
fi

echo ""
echo "[3/4] Removing bash completion..."
if [ -f "$COMPLETION_DIR_BASH/openpass" ]; then
    sudo rm -f "$COMPLETION_DIR_BASH/openpass"
    echo "Removed bash completion"
else
    echo "Bash completion not found"
fi

echo ""
echo "[4/4] Removing zsh completion..."
if [ -f "$COMPLETION_DIR_ZSH/_openpass" ]; then
    sudo rm -f "$COMPLETION_DIR_ZSH/_openpass"
    echo "Removed zsh completion"
else
    echo "Zsh completion not found"
fi

echo ""
echo "OpenPasswd Uninstalled Successfully"
echo "==================================="
echo ""
echo "Your password data is preserved in:"
echo "  ~/.config/openpasswd/"
echo ""
echo "To completely remove all data:"
echo "  rm -rf ~/.config/openpasswd"
echo ""
