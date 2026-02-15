#!/usr/bin/env bash
set -euo pipefail

MIN_GO_VERSION="1.21"

echo "=== Goodreads CLI Installer ==="
echo

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed."
    echo "Install Go >= $MIN_GO_VERSION from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
echo "Found Go $GO_VERSION"

# Check Go version >= 1.21
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)
REQ_MAJOR=$(echo "$MIN_GO_VERSION" | cut -d. -f1)
REQ_MINOR=$(echo "$MIN_GO_VERSION" | cut -d. -f2)

if [ "$GO_MAJOR" -lt "$REQ_MAJOR" ] || { [ "$GO_MAJOR" -eq "$REQ_MAJOR" ] && [ "$GO_MINOR" -lt "$REQ_MINOR" ]; }; then
    echo "Error: Go >= $MIN_GO_VERSION required, found $GO_VERSION"
    exit 1
fi

# On Linux, check/suggest Chromium system dependencies
if [ "$(uname -s)" = "Linux" ]; then
    echo "Detected Linux — checking Chromium dependencies..."
    MISSING_LIBS=false
    for lib in libnss3.so libatk-1.0.so.0 libatk-bridge-2.0.so.0 libcups.so.2 libXdamage.so.1 libXrandr.so.2 libgbm.so.1 libpango-1.0.so.0 libcairo.so.2 libasound.so.2 libXcomposite.so.1 libXfixes.so.3 libxkbcommon.so.0 libdrm.so.2 libatspi.so.0; do
        if ! ldconfig -p 2>/dev/null | grep -q "$lib"; then
            MISSING_LIBS=true
            break
        fi
    done
    if [ "$MISSING_LIBS" = true ]; then
        echo "Some Chromium dependencies are missing. Install them with:"
        echo "  sudo apt install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libxdamage1 \\"
        echo "    libxrandr2 libgbm1 libpango-1.0-0 libcairo2 libasound2 libxcomposite1 \\"
        echo "    libxfixes3 libxkbcommon0 libdrm2 libatspi2.0-0"
        echo
    else
        echo "Chromium dependencies OK"
    fi
fi

# Check for Chrome/Chromium
if command -v google-chrome &> /dev/null || \
   command -v chromium &> /dev/null || \
   command -v chromium-browser &> /dev/null || \
   [ -d "/Applications/Google Chrome.app" ] 2>/dev/null; then
    echo "Found Chrome/Chromium"
else
    echo "Note: Chrome/Chromium not found on PATH."
    echo "Rod will automatically download Chromium on first run."
fi

echo
echo "Installing dependencies..."
go mod tidy

echo
echo "Building goodreads CLI..."
go build -o goodreads .

echo "Building goodreads-recorder..."
go build -o goodreads-recorder ./cmd/recorder

# Determine install directory
INSTALL_DIR="${GOPATH:-$HOME/go}/bin"
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

echo
echo "Installing to $INSTALL_DIR..."
mv goodreads "$INSTALL_DIR/goodreads"
mv goodreads-recorder "$INSTALL_DIR/goodreads-recorder"

echo
echo "Installation complete!"
echo "  goodreads          - main CLI"
echo "  goodreads-recorder - request recorder for reverse engineering"
echo
echo "Make sure $INSTALL_DIR is in your PATH."
echo
echo "Next steps:"
echo "  1. Create ~/.goodreads-cli.yaml with your credentials:"
echo "     email: you@example.com"
echo "     password: yourpassword"
echo "  2. Run: goodreads login"
