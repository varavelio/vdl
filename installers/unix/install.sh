#!/bin/sh
set -e

# ======================================================================================
# VDL Installer Script (macOS/Linux)
#
# Usage:
#   curl -fsSL https://get.varavel.com/vdl | sh
#
# Options:
#   VERSION      : Specify a version (e.g., vx.x.x). Defaults to "latest".
#   INSTALL_DIR  : Directory to install the binary. Defaults to "/usr/local/bin".
#   QUIET        : Set to "true" to suppress all output (e.g., QUIET=true).
#
# Examples:
#   # Install or update to latest version
#   curl -fsSL https://get.varavel.com/vdl | sh
#
#   # Install specific version
#   curl -fsSL https://get.varavel.com/vdl | VERSION=vx.x.x sh
#
#   # Install to a custom directory quietly
#   curl -fsSL https://get.varavel.com/vdl | INSTALL_DIR=$HOME/.local/bin QUIET=true sh
# ======================================================================================

# Configuration
REPO="varavelio/vdl"
BINARY_NAME="vdl"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-}"
QUIET="${QUIET:-false}"

# Colors and formatting
setup_colors() {
  if [ -t 1 ] && [ "$QUIET" != "true" ]; then
    RED='\033[31m'
    GREEN='\033[32m'
    YELLOW='\033[33m'
    BLUE='\033[34m'
    BOLD='\033[1m'
    NC='\033[0m'
  else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    BOLD=''
    NC=''
  fi
}

log_info() {
  if [ "$QUIET" != "true" ]; then printf "${GREEN}[INFO]${NC} %s\n" "$1"; fi
}

log_warn() {
  if [ "$QUIET" != "true" ]; then printf "${YELLOW}[WARN]${NC} %s\n" "$1"; fi
}

log_error() {
  if [ "$QUIET" != "true" ]; then printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; fi
}

print_banner() {
  if [ "$QUIET" != "true" ]; then
    printf "${BLUE}${BOLD}"
    printf "██╗  ██╗█████╗ ██╗\n"
    printf "██║  ██║██╔═██╗██║\n"
    printf "╚██╗██╔╝██║ ██║██║\n"
    printf " ╚███╔╝ █████╔╝█████╗\n"
    printf "  ╚══╝  ╚════╝ ╚════╝"
    printf "${NC}\n"
  fi
}

cleanup() {
  if [ -d "$TMP_DIR" ]; then rm -rf "$TMP_DIR"; fi
}
trap cleanup EXIT

# Dependency Check
check_dependencies() {
  if ! command -v tar >/dev/null 2>&1; then
    log_error "Missing dependency: tar is required."
    exit 1
  fi
  if ! command -v grep >/dev/null 2>&1; then
    log_error "Missing dependency: grep is required."
    exit 1
  fi
  if ! command -v sed >/dev/null 2>&1; then
    log_error "Missing dependency: sed is required."
    exit 1
  fi
  if ! command -v uname >/dev/null 2>&1; then
    log_error "Missing dependency: uname is required."
    exit 1
  fi
}

# Detect OS and Architecture
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux) PLATFORM_OS="linux" ;;
    Darwin) PLATFORM_OS="darwin" ;;
    *) log_error "Unsupported OS: $OS"; exit 1 ;;
  esac

  case "$ARCH" in
    x86_64|amd64) PLATFORM_ARCH="amd64" ;;
    arm64|aarch64) PLATFORM_ARCH="arm64" ;;
    *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
  esac
}

# Determine Version
get_version() {
  if [ -z "$VERSION" ] || [ "$VERSION" = "latest" ]; then
    log_info "Fetching latest version..."
    VERSION=$(curl -sSfI "https://github.com/$REPO/releases/latest" | grep -i "location:" | sed -E 's/.*\/tag\/v?([^\r\n]+).*/v\1/')
  fi

  # Ensure version starts with 'v'
  case "$VERSION" in
    v*) ;;
    *) VERSION="v$VERSION" ;;
  esac

  if [ -z "$VERSION" ]; then
    log_error "Failed to determine version."
    exit 1
  fi
}

# Download and Verify
download_and_install() {
  TMP_DIR=$(mktemp -d)

  FILENAME="${BINARY_NAME}_${PLATFORM_OS}_${PLATFORM_ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"
  CHECKSUMS_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"

  log_info "Installing version: $VERSION"
  log_info "Downloading $FILENAME..."

  curl -sSfL "$DOWNLOAD_URL" -o "$TMP_DIR/$FILENAME" || { log_error "Download failed."; exit 1; }
  curl -sSfL "$CHECKSUMS_URL" -o "$TMP_DIR/checksums.txt" || { log_error "Checksum download failed."; exit 1; }

  log_info "Verifying checksum..."
  cd "$TMP_DIR"
  if command -v sha256sum >/dev/null 2>&1; then
    grep "$FILENAME" checksums.txt | sha256sum -c - >/dev/null 2>&1 || { log_error "Checksum verification failed!"; exit 1; }
  elif command -v shasum >/dev/null 2>&1; then
    grep "$FILENAME" checksums.txt | shasum -a 256 -c - >/dev/null 2>&1 || { log_error "Checksum verification failed!"; exit 1; }
  else
    log_warn "Neither sha256sum nor shasum found. Skipping verification."
  fi
  cd - >/dev/null

  log_info "Extracting..."
  tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

  BIN_SOURCE="$TMP_DIR/$BINARY_NAME"
  if [ ! -f "$BIN_SOURCE" ]; then
    log_error "Binary not found in archive."
    exit 1
  fi

  log_info "Installing to $INSTALL_DIR..."

  if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || {
      if command -v sudo >/dev/null 2>&1; then
        sudo mkdir -p "$INSTALL_DIR"
      else
        log_error "Cannot create directory $INSTALL_DIR. Permission denied. Retry with sudo."
        exit 1
      fi
    }
  fi

  # Check writability
  if [ -w "$INSTALL_DIR" ]; then
    mv "$BIN_SOURCE" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
  else
    if command -v sudo >/dev/null 2>&1; then
      if [ -t 0 ]; then
        log_warn "$INSTALL_DIR is not writable. Attempting sudo..."
        sudo mv "$BIN_SOURCE" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
      else
        log_error "Cannot install to $INSTALL_DIR (permission denied) and no TTY for sudo."
        log_info "Try: curl -fsSL https://get.varavel.com/vdl | INSTALL_DIR=\$HOME/.local/bin sh"
        exit 1
      fi
    else
      log_error "Cannot install to $INSTALL_DIR. Permission denied."
      exit 1
    fi
  fi

  log_info "Installation complete!"
  log_info "Run '$BINARY_NAME --version' to verify."
}

setup_colors
print_banner
check_dependencies
detect_platform
get_version
download_and_install
