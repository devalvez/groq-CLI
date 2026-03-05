#!/usr/bin/env bash
# Groq CLI Installer
# Usage: ./install.sh

set -e

BINARY="groq"
INSTALL_DIR="/usr/local/bin"
BUILD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

echo ""
echo -e "${CYAN}${BOLD}"
echo " ______  ______  ______  ______     ______  __      __   "
echo "/\  ___\/\  == \/\  __ \/\  __ \   /\  ___\/\ \    /\ \  "
echo "\ \ \__ \ \  __<\ \ \/\ \ \ \/\_\  \ \ \___\ \ \___\ \ \ "
echo " \ \_____\ \_\ \_\ \_____\ \___\_\  \ \_____\ \_____\ \_\\"
echo "  \/_____/\/_/ /_/\/_____/\/___/_/   \/_____/\/_____/\/_/"
echo -e "${RESET}"
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo -e "  ${CYAN}  Groq CLI Installer${RESET} ${DIM}· Powered by ⚡Groq AI${RESET}"
echo -e "  %{CYAN}           Developed by Wesley Alves (Devalvez)                 "
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo ""

# Check Go installation
echo -e "${DIM}  Checking dependencies...${RESET}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}  ✗ Go is not installed!${RESET}"
    echo ""
    echo -e "${YELLOW}  Install Go from: https://go.dev/doc/install${RESET}"
    echo ""
    echo -e "${DIM}  Quick install on Ubuntu/Debian:${RESET}"
    echo -e "${CYAN}    sudo apt-get install golang-go${RESET}"
    echo ""
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}  ✓ Go found: ${GO_VERSION}${RESET}"

# Navigate to source directory
cd "$BUILD_DIR"

# Download dependencies
echo -e "${DIM}  Downloading dependencies...${RESET}"
go mod tidy 2>&1 | sed 's/^/    /'
echo -e "${GREEN}  ✓ Dependencies ready${RESET}"

# Build
echo -e "${DIM}  Building binary...${RESET}"
go build -ldflags="-s -w -X main.version=1.0.0" -o "$BINARY" . 2>&1 | sed 's/^/    /'

if [ ! -f "$BINARY" ]; then
    echo -e "${RED}  ✗ Build failed!${RESET}"
    exit 1
fi

echo -e "${GREEN}  ✓ Build successful${RESET}"

# Install
echo -e "${DIM}  Installing to ${INSTALL_DIR}...${RESET}"
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY" "$INSTALL_DIR/$BINARY"
else
    sudo mv "$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo -e "${GREEN}  ✓ Installed to ${INSTALL_DIR}/${BINARY}${RESET}"

# Verify
echo ""
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo -e "${GREEN}${BOLD}  🎉 Installation complete!${RESET}"
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo ""
echo -e "${YELLOW}  Next steps:${RESET}"
echo ""
echo -e "  ${DIM}1. Get your free API key at:${RESET}"
echo -e "     ${CYAN}https://console.groq.com${RESET}"
echo ""
echo -e "  ${DIM}2. Set your API key:${RESET}"
echo -e "     ${CYAN}groq config set-key YOUR_API_KEY${RESET}"
echo -e "     ${DIM}or${RESET}"
echo -e "     ${CYAN}export GROQ_API_KEY=your_key${RESET}"
echo ""
echo -e "  ${DIM}3. Start using:${RESET}"
echo -e "     ${CYAN}groq${RESET}                         ${DIM}# Welcome screen${RESET}"
echo -e "     ${CYAN}groq chat${RESET}                    ${DIM}# Interactive chat${RESET}"
echo -e "     ${CYAN}groq create \"my project\"${RESET}     ${DIM}# Generate a project${RESET}"
echo -e "     ${CYAN}groq run \"list files by size\"${RESET} ${DIM}# Run a task${RESET}"
echo ""
echo -e "${DIM}  ─────────────────────────────────────────────────────────${RESET}"
echo ""
