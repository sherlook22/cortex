#!/bin/bash
set -e

REPO="sherlook22/cortex"
BINARY="cortex"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { printf "${GREEN}[info]${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}[warn]${NC} %s\n" "$1"; }
error() { printf "${RED}[error]${NC} %s\n" "$1"; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       error "Unsupported OS: $(uname -s). Only Linux and macOS are supported." ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)             error "Unsupported architecture: $(uname -m). Only amd64 and arm64 are supported." ;;
    esac
}

# Get latest release tag from GitHub API
get_latest_version() {
    if command -v curl > /dev/null 2>&1; then
        curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget > /dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "curl or wget is required for installation."
    fi
}

# Download file
download() {
    local url="$1" dest="$2"
    if command -v curl > /dev/null 2>&1; then
        curl -sL "$url" -o "$dest"
    elif command -v wget > /dev/null 2>&1; then
        wget -qO "$dest" "$url"
    fi
}

main() {
    local os arch version archive_name url tmp_dir

    os=$(detect_os)
    arch=$(detect_arch)

    info "Detected OS: ${os}, Arch: ${arch}"

    # Allow version override via argument
    if [ -n "$1" ]; then
        version="$1"
    else
        info "Fetching latest version..."
        version=$(get_latest_version)
    fi

    if [ -z "$version" ]; then
        error "Could not determine latest version. Check https://github.com/${REPO}/releases"
    fi

    info "Installing ${BINARY} ${version}..."

    archive_name="${BINARY}_${os}_${arch}.tar.gz"
    url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"

    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    info "Downloading ${url}..."
    download "$url" "${tmp_dir}/${archive_name}"

    if [ ! -f "${tmp_dir}/${archive_name}" ] || [ ! -s "${tmp_dir}/${archive_name}" ]; then
        error "Download failed. Check that version ${version} exists at https://github.com/${REPO}/releases"
    fi

    info "Extracting..."
    tar -xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"

    if [ ! -f "${tmp_dir}/${BINARY}" ]; then
        error "Binary not found in archive."
    fi

    info "Installing to ${INSTALL_DIR}/${BINARY}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        warn "Requires sudo to install to ${INSTALL_DIR}"
        sudo mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi

    chmod +x "${INSTALL_DIR}/${BINARY}"

    info "Installed successfully!"
    info "Run '${BINARY} --help' to get started."
    "${INSTALL_DIR}/${BINARY}" version
}

main "$@"
