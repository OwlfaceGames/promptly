#!/bin/bash

set -e

# Promptly installer script
REPO="owlfacegames/promptly"
BINARY_NAME="promptly"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)
            os="linux"
            ;;
        Darwin*)
            os="darwin"
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            os="windows"
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Get the latest release version
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
main() {
    print_status "Installing Promptly..."
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    print_status "Detected platform: ${platform}"
    
    # Get latest version
    print_status "Fetching latest version..."
    local version
    version=$(get_latest_version)
    if [ -z "$version" ]; then
        print_error "Failed to get latest version"
        exit 1
    fi
    print_status "Latest version: ${version}"
    
    # Set binary name based on OS
    local binary_name="${BINARY_NAME}"
    if [[ "$platform" == *"windows"* ]]; then
        binary_name="${BINARY_NAME}.exe"
    fi
    
    # Construct download URL
    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}"
    if [[ "$platform" == *"windows"* ]]; then
        download_url="${download_url}.exe"
    fi
    
    # Create temporary directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    local tmp_file="${tmp_dir}/${binary_name}"
    
    # Download binary
    print_status "Downloading from ${download_url}..."
    if ! curl -L "$download_url" -o "$tmp_file"; then
        print_error "Failed to download binary"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    # Make executable
    chmod +x "$tmp_file"
    
    # Install binary
    print_status "Installing to ${INSTALL_DIR}..."
    if [[ "$EUID" -eq 0 ]]; then
        # Running as root
        mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        # Need sudo
        if command -v sudo >/dev/null 2>&1; then
            sudo mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
        else
            print_warning "sudo not available. Please move ${tmp_file} to ${INSTALL_DIR}/${BINARY_NAME} manually"
            print_warning "Or run: cp ${tmp_file} ${INSTALL_DIR}/${BINARY_NAME}"
            exit 0
        fi
    fi
    
    # Cleanup
    rm -rf "$tmp_dir"
    
    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        print_success "Promptly installed successfully!"
        print_success "Run 'promptly' to get started"
    else
        print_warning "Installation completed, but 'promptly' not found in PATH"
        print_warning "You may need to add ${INSTALL_DIR} to your PATH or restart your terminal"
    fi
}

# Check dependencies
if ! command -v curl >/dev/null 2>&1; then
    print_error "curl is required but not installed"
    exit 1
fi

main "$@"
