#!/usr/bin/env bash
# Groq CLI Uninstaller
# Usage: ./uninstall.sh

set -e

BINARY="groq"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/groq-cli"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

echo ""
echo -e "${RED}${BOLD}"
echo "  ██████╗ ██████╗  ██████╗  ██████╗     ██████╗██╗     ██╗"
echo " ██╔════╝ ██╔══██╗██╔═══██╗██╔═══██╗   ██╔════╝██║     ██║"
echo " ██║  ███╗██████╔╝██║   ██║██║   ██║   ██║     ██║     ██║"
echo " ██║   ██║██╔══██╗██║   ██║██║▄▄ ██║   ██║     ██║     ██║"
echo " ╚██████╔╝██║  ██║╚██████╔╝╚██████╔╝   ╚██████╗███████╗██║"
echo "  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝  ╚══▀▀═╝    ╚═════╝╚══════╝╚═╝"
echo -e "${RESET}"
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo -e "  ${RED}${BOLD}🗑  Groq CLI Uninstaller${RESET}"
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo ""

# Check if installed
BINARY_PATH="$INSTALL_DIR/$BINARY"
if [ ! -f "$BINARY_PATH" ]; then
    # Try to find it elsewhere
    BINARY_PATH=$(which "$BINARY" 2>/dev/null || echo "")
    if [ -z "$BINARY_PATH" ]; then
        echo -e "${YELLOW}  ⚠  Groq CLI binary not found in PATH.${RESET}"
        echo -e "${DIM}  It may have already been removed or installed elsewhere.${RESET}"
        echo ""
    else
        echo -e "${CYAN}  Found binary at: ${BINARY_PATH}${RESET}"
    fi
else
    echo -e "${CYAN}  Found binary at: ${BINARY_PATH}${RESET}"
fi

# Show what will be removed
echo ""
echo -e "${YELLOW}  The following will be removed:${RESET}"
echo ""

if [ -n "$BINARY_PATH" ] && [ -f "$BINARY_PATH" ]; then
    echo -e "  ${DIM}[binary]${RESET}  $BINARY_PATH"
fi

if [ -d "$CONFIG_DIR" ]; then
    echo -e "  ${DIM}[config]${RESET}  $CONFIG_DIR"
    echo -e "  ${DIM}         ${RESET}  └─ config.json (API key, settings)"
else
    echo -e "  ${DIM}[config]${RESET}  $CONFIG_DIR ${DIM}(not found, skipping)${RESET}"
fi

echo ""

# Confirm
echo -ne "${RED}  Are you sure you want to uninstall Groq CLI? [y/N] ${RESET}"
read -r CONFIRM
echo ""

if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo -e "${GREEN}  ✓ Uninstall cancelled. Groq CLI is still installed.${RESET}"
    echo ""
    exit 0
fi

# Ask about config
REMOVE_CONFIG=false
if [ -d "$CONFIG_DIR" ]; then
    echo -ne "${YELLOW}  Also remove config and saved API key? [y/N] ${RESET}"
    read -r CONFIRM_CONFIG
    echo ""
    if [[ "$CONFIRM_CONFIG" =~ ^[Yy]$ ]]; then
        REMOVE_CONFIG=true
    fi
fi

echo -e "${DIM}  Removing...${RESET}"
echo ""

# Remove binary
if [ -n "$BINARY_PATH" ] && [ -f "$BINARY_PATH" ]; then
    if [ -w "$(dirname "$BINARY_PATH")" ]; then
        rm -f "$BINARY_PATH"
    else
        sudo rm -f "$BINARY_PATH"
    fi
    echo -e "${GREEN}  ✓ Removed binary:${RESET} $BINARY_PATH"
else
    echo -e "${DIM}  - Binary not found, skipping.${RESET}"
fi

# Remove config
if [ "$REMOVE_CONFIG" = true ] && [ -d "$CONFIG_DIR" ]; then
    rm -rf "$CONFIG_DIR"
    echo -e "${GREEN}  ✓ Removed config:${RESET} $CONFIG_DIR"
elif [ -d "$CONFIG_DIR" ]; then
    echo -e "${DIM}  - Config kept at: $CONFIG_DIR${RESET}"
fi

# Verify
echo ""
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
if command -v "$BINARY" &> /dev/null; then
    echo -e "${YELLOW}  ⚠  Warning: 'groq' still found in PATH at: $(which groq)${RESET}"
    echo -e "${DIM}     You may need to remove it manually.${RESET}"
else
    echo -e "${GREEN}${BOLD}  ✅ Groq CLI successfully uninstalled!${RESET}"
fi
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo ""

if [ "$REMOVE_CONFIG" = false ] && [ -d "$CONFIG_DIR" ]; then
    echo -e "${DIM}  Your config was kept at: $CONFIG_DIR${RESET}"
    echo -e "${DIM}  To remove it manually:   rm -rf $CONFIG_DIR${RESET}"
    echo ""
fi

echo -e "${DIM}  Thanks for using Groq CLI! Come back anytime 👋${RESET}"
echo ""
