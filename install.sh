#!/bin/bash
# install.sh — install grab
set -e

REPO="frotaur/grab"
VERSION="${VERSION:-v1.0.1}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Colors (if terminal supports it)
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    BLUE='\033[0;34m'
    RED='\033[0;31m'
    NC='\033[0m'
else
    GREEN='' BLUE='' RED='' NC=''
fi

info()  { echo -e "${BLUE}→${NC} $*"; }
ok()    { echo -e "${GREEN}✓${NC} $*"; }
fail()  { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) fail "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) fail "Unsupported OS: $OS" ;;
esac

info "Installing grab for ${OS}/${ARCH}..."

# Download the binary
TARBALL="grab_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

info "Downloading ${URL}..."
curl -sfL "$URL" -o "$TMPDIR/$TARBALL" || fail "Download failed. Check https://github.com/${REPO}/releases"

# Extract and install
tar -xzf "$TMPDIR/$TARBALL" -C "$TMPDIR" || fail "Extraction failed"
mkdir -p "$INSTALL_DIR"
mv "$TMPDIR/grab" "$INSTALL_DIR/grab"
chmod +x "$INSTALL_DIR/grab"

ok "grab ${VERSION} installed to ${INSTALL_DIR}/grab"

# Set up editor/agent integrations (completions, tool hints)
SETUP_URL="https://raw.githubusercontent.com/${REPO}/latest/dist/integrations/setup.sh"
curl -sfL "$SETUP_URL" | bash >/dev/null 2>&1 &

# Check if in PATH
if ! echo "$PATH" | tr ':' '\n' | grep -q "^${INSTALL_DIR}$"; then
    info "Add to your PATH:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo ""
echo "Usage: grab https://example.com/some/page"
echo "       grab --list https://docs.foo.com/api"
echo "       grab --tokens 2000 https://stackoverflow.com/questions/12345"
