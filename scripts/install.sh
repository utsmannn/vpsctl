#!/bin/bash
#
# vpsctl - Modern VPS management tool using LXD
# Install script
#
# Usage:
#   # From cloned repo (recommended for private repo)
#   git clone git@github.com:utsmannn/vpsctl.git
#   cd vpsctl
#   ./scripts/install.sh
#
#   # Or with options
#   ./scripts/install.sh -v v1.0.0
#   ./scripts/install.sh -b /usr/local/bin
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
REPO="utsmannn/vpsctl"
VERSION="latest"
BIN_DIR="/usr/local/bin"
BINARY_NAME="vpsctl"

# Print functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -b|--bin-dir)
            BIN_DIR="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -v, --version VERSION   Install specific version (default: latest)"
            echo "  -b, --bin-dir DIR       Installation directory (default: /usr/local/bin)"
            echo "  -h, --help              Show this help message"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv7)
            ARCH="arm"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: ${PLATFORM}"
}

# Get latest version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        info "Fetching latest version..."
        VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            error "Failed to get latest version. Please specify with -v"
        fi
    fi
    info "Installing version: ${VERSION}"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"

    info "Downloading from: ${DOWNLOAD_URL}"

    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${BINARY_NAME}"

    if ! curl -fsSL --progress-bar -o "${TMP_FILE}" "${DOWNLOAD_URL}"; then
        error "Failed to download binary. Check if version ${VERSION} exists for platform ${PLATFORM}"
    fi

    chmod +x "${TMP_FILE}"
    success "Download complete"
}

# Verify checksum (optional)
verify_checksum() {
    CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}.sha256"

    if curl -fsSL -o "${TMP_DIR}/checksum.sha256" "${CHECKSUM_URL}" 2>/dev/null; then
        info "Verifying checksum..."
        cd "${TMP_DIR}"
        if sha256sum -c checksum.sha256 > /dev/null 2>&1; then
            success "Checksum verified"
        else
            warn "Checksum verification failed. Continuing anyway..."
        fi
        cd - > /dev/null
    else
        warn "Checksum file not available, skipping verification"
    fi
}

# Install binary
install_binary() {
    # Check if bin directory exists
    if [ ! -d "${BIN_DIR}" ]; then
        info "Creating directory: ${BIN_DIR}"
        sudo mkdir -p "${BIN_DIR}"
    fi

    # Check write permissions
    if [ ! -w "${BIN_DIR}" ]; then
        info "Need sudo to install to ${BIN_DIR}"
        sudo mv "${TMP_FILE}" "${BIN_DIR}/${BINARY_NAME}"
    else
        mv "${TMP_FILE}" "${BIN_DIR}/${BINARY_NAME}"
    fi

    success "Installed to ${BIN_DIR}/${BINARY_NAME}"
}

# Cleanup
cleanup() {
    if [ -d "${TMP_DIR}" ]; then
        rm -rf "${TMP_DIR}"
    fi
}

# Install systemd service (Linux only)
install_systemd_service() {
    if [ "$OS" != "linux" ]; then
        return
    fi

    echo ""
    read -p "Install systemd service for API server? (y/N): " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        return
    fi

    SERVICE_FILE="/etc/systemd/system/vpsctl.service"

    info "Creating systemd service..."

    sudo tee "${SERVICE_FILE}" > /dev/null << 'EOF'
[Unit]
Description=vpsctl API Server
After=network.target lxd.socket
Requires=lxd.socket

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/vpsctl serve --port 8080
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable vpsctl

    success "Systemd service installed"
    info "Start with: sudo systemctl start vpsctl"
    info "Status with: sudo systemctl status vpsctl"
}

# Print usage examples
print_examples() {
    echo ""
    echo -e "${GREEN}vpsctl installed successfully!${NC}"
    echo ""
    echo "Quick start:"
    echo "  vpsctl list                           # List all instances"
    echo "  vpsctl create myserver                # Create new instance"
    echo "  vpsctl create web --cpu 2 --memory 1GB"
    echo "  vpsctl shell myserver                 # Open shell"
    echo "  vpsctl dashboard                      # TUI dashboard"
    echo "  vpsctl serve --port 8080              # Start API server"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
}

# Main
main() {
    echo -e "${GREEN}"
    echo "                                                            "
    echo "                                                  ▄▄▄▄     "
    echo "                                          ██      ▀▀██     "
    echo "  ██▄  ▄██  ██▄███▄   ▄▄█████▄   ▄███████  ███████     ██     "
    echo "   ██  ██   ██▀  ▀██  ██▄▄▄▄ ▀  ██▀    ▀    ██        ██     "
    echo "   ▀█▄▄█▀   ██    ██   ▀▀▀▀██▄  ██          ██        ██     "
    echo "    ████    ███▄▄██▀  █▄▄▄▄▄██  ▀██▄▄▄▄█    ██▄▄▄     ██▄▄▄  "
    echo "     ▀▀     ██ ▀▀▀     ▀▀▀▀▀▀     ▀▀▀▀▀      ▀▀▀▀      ▀▀▀▀  "
    echo "            ██                                               "
    echo "                                                            "
    echo -e "${NC}"
    echo "  Modern VPS Management Tool"
    echo ""

    trap cleanup EXIT

    detect_platform
    get_latest_version
    download_binary
    verify_checksum
    install_binary
    install_systemd_service
    print_examples
}

main "$@"
