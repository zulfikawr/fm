#!/bin/bash
set -e

# Gruvbox Theme Colors (matching internal/tui/theme/theme.go)
COLOR_DIR='\033[38;5;208m'      # Orange
COLOR_EXEC='\033[38;5;142m'     # Green
COLOR_FILE='\033[38;5;223m'     # Light Cream
COLOR_SUBTLE='\033[38;5;243m'   # Dark Gray
COLOR_PRIMARY='\033[38;5;208m'  # Orange
COLOR_SECONDARY='\033[38;5;142m' # Green
COLOR_ACCENT='\033[38;5;109m'   # Blue-gray
COLOR_MUTED='\033[38;5;245m'    # Light Gray
COLOR_HIGHLIGHT='\033[38;5;214m' # Bright Yellow/Orange
COLOR_INFO='\033[38;5;66m'      # Blue-teal
COLOR_SUCCESS='\033[38;5;142m'  # Green
COLOR_WARNING='\033[38;5;214m'  # Yellow/Orange
COLOR_ERROR='\033[38;5;167m'    # Red
COLOR_BOLD='\033[1m'
NC='\033[0m' # No Color

# Header
echo -e "\n${COLOR_PRIMARY}${COLOR_BOLD}FM - Terminal File Manager Installer${NC}"
echo -e "${COLOR_SUBTLE}──────────────────────────────────────────────────${NC}\n"

REPO="zulfikawr/fm"
LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"

# 1. Dependency Check
echo -e "${COLOR_INFO}::${NC} ${COLOR_BOLD}Validating system dependencies...${NC}"
for cmd in curl grep cut mktemp; do
    if ! command -v $cmd &> /dev/null; then
        echo -e "   ${COLOR_ERROR}✗${NC} ${COLOR_ERROR}Error:${NC} ${COLOR_FILE}$cmd${NC} is required but not installed."
        exit 1
    fi
done
echo -e "   ${COLOR_SUCCESS}✓${NC} All dependencies found.\n"

# 2. Detect Architecture
echo -e "${COLOR_INFO}::${NC} ${COLOR_BOLD}Identifying system architecture...${NC}"
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)     OS="linux" ;;
    Darwin)    OS="darwin" ;;
    *)         echo -e "   ${COLOR_ERROR}✗${NC} ${COLOR_ERROR}Error:${NC} Unsupported OS: ${COLOR_FILE}$OS${NC}"; exit 1 ;;
esac

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo -e "   ${COLOR_ERROR}✗${NC} ${COLOR_ERROR}Error:${NC} Unsupported Architecture: ${COLOR_FILE}$ARCH${NC}"; exit 1 ;;
esac
echo -e "   ${COLOR_SUCCESS}✓${NC} Detected ${COLOR_ACCENT}$OS${NC} on ${COLOR_ACCENT}$ARCH${NC} platform.\n"

# 3. Fetch Release Info
echo -e "${COLOR_INFO}::${NC} ${COLOR_BOLD}Querying GitHub for latest release...${NC}"
RELEASE_DATA=$(curl -s $LATEST_URL)
VERSION=$(echo "$RELEASE_DATA" | grep "tag_name" | cut -d '"' -f 4)
DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep "browser_download_url" | grep "$OS-$ARCH" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo -e "   ${COLOR_ERROR}✗${NC} ${COLOR_ERROR}Error:${NC} Could not find binary for ${COLOR_FILE}$OS/$ARCH${NC}"
    exit 1
fi
echo -e "   ${COLOR_SUCCESS}✓${NC} Found ${COLOR_HIGHLIGHT}$VERSION${NC} as the most recent version.\n"

# 4. Path Selection
# If it's not a terminal (e.g. piped from curl), default to system-wide if root, else ask if possible
if [ -t 0 ]; then
    echo -e "${COLOR_INFO}::${NC} ${COLOR_BOLD}Select installation target:${NC}"
    echo -e "   ${COLOR_HIGHLIGHT}1)${NC} ${COLOR_FILE}System-wide${NC}  ${COLOR_SUBTLE}(/usr/local/bin/fm)${NC} ${COLOR_BOLD}${COLOR_WARNING}*${NC} ${COLOR_SUBTLE}requires sudo${NC}"
    echo -e "   ${COLOR_HIGHLIGHT}2)${NC} ${COLOR_FILE}User-only${NC}    ${COLOR_SUBTLE}(~/.local/bin/fm)${NC}"
    echo -ne "   ${COLOR_BOLD}${COLOR_INFO}>>${NC} ${COLOR_FILE}Pick an option [1-2, default 1]: ${NC}"
    read choice < /dev/tty || choice=1
    choice=${choice:-1}
    echo ""
else
    # Non-interactive: default to /usr/local/bin if possible, else ~/.local/bin
    if [ "$EUID" -eq 0 ] || [ -w "/usr/local/bin" ]; then
        choice=1
    else
        choice=2
    fi
fi

if [ "$choice" == "1" ]; then
    INSTALL_DIR="/usr/local/bin"
    if [ "$EUID" -ne 0 ] && [ ! -w "$INSTALL_DIR" ]; then
        USE_SUDO=true
    else
        USE_SUDO=false
    fi
else
    INSTALL_DIR="$HOME/.local/bin"
    USE_SUDO=false
    mkdir -p "$INSTALL_DIR"
fi

# 5. Download
echo -e "${COLOR_INFO}::${NC} ${COLOR_BOLD}Downloading binary...${NC}"
TMP_BIN=$(mktemp)

# Use curl with a Gruvbox primary color progress bar
echo -ne "${COLOR_PRIMARY}"
curl -L -# -o "$TMP_BIN" "$DOWNLOAD_URL"
echo -ne "${NC}"
echo ""
chmod 755 "$TMP_BIN"

# 6. Install
echo -e "${COLOR_INFO}::${NC} ${COLOR_BOLD}Moving binary to installation directory...${NC}"
if [ "$USE_SUDO" = true ]; then
    sudo mv "$TMP_BIN" "$INSTALL_DIR/fm"
    sudo chown root:root "$INSTALL_DIR/fm"
else
    mv "$TMP_BIN" "$INSTALL_DIR/fm"
fi

echo -e "\n${COLOR_SUBTLE}──────────────────────────────────────────────────${NC}"
echo -e "${COLOR_SUCCESS}${COLOR_BOLD}Success!${NC} ${COLOR_FILE}fm${NC} ${COLOR_HIGHLIGHT}$VERSION${NC} has been deployed to ${COLOR_ACCENT}$INSTALL_DIR${NC}."

# Check if path is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "\n${COLOR_INFO}::${NC} ${COLOR_BOLD}Automatically adding ${COLOR_ACCENT}$INSTALL_DIR${NC} to your PATH...${NC}"
    
    ADDED_TO=""
    # For Bash
    if [ -f "$HOME/.bashrc" ]; then
        if ! grep -q "$INSTALL_DIR" "$HOME/.bashrc"; then
            echo -e "\n# fm - Terminal File Manager\nexport PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
            ADDED_TO="$ADDED_TO .bashrc"
        fi
    fi
    
    # For Zsh
    if [ -f "$HOME/.zshrc" ]; then
        if ! grep -q "$INSTALL_DIR" "$HOME/.zshrc"; then
            echo -e "\n# fm - Terminal File Manager\nexport PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.zshrc"
            ADDED_TO="$ADDED_TO .zshrc"
        fi
    fi

    # For Profile
    if [ -f "$HOME/.profile" ]; then
        if ! grep -q "$INSTALL_DIR" "$HOME/.profile"; then
            echo -e "\n# fm - Terminal File Manager\nexport PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.profile"
            ADDED_TO="$ADDED_TO .profile"
        fi
    fi

    if [ -n "$ADDED_TO" ]; then
        echo -e "   ${COLOR_SUCCESS}✓${NC} Added to:$ADDED_TO"
        echo -e "   ${COLOR_WARNING}Note:${NC} To use 'fm' immediately, run: ${COLOR_PRIMARY}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
    else
        echo -e "   ${COLOR_WARNING}⚠️  Warning:${NC} Could not automatically update PATH. Please add ${COLOR_ACCENT}$INSTALL_DIR${NC} manually."
    fi
fi

echo -e "\n${COLOR_INFO}Execute ${COLOR_BOLD}'fm'${NC} ${COLOR_INFO}to begin navigation.${NC}\n"
