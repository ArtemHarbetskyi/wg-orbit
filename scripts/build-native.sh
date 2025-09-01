#!/bin/bash

# Native multi-architecture build script for wg-orbit
# Builds native binaries for different architectures without Docker

set -e

# Configuration
VERSION=${VERSION:-"$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')"}
OUTPUT_DIR="bin"
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Supported architectures (platform:goos:goarch:goarm)
ARCHITECTURES=(
    "linux-amd64:linux:amd64:"
    "linux-arm64:linux:arm64:"
    "linux-armv7:linux:arm:7"
    "linux-armv6:linux:arm:6"
    "darwin-amd64:darwin:amd64:"
    "darwin-arm64:darwin:arm64:"
    "windows-amd64:windows:amd64:"
)

echo -e "${BLUE}🚀 Building native binaries for wg-orbit${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Output directory: ${OUTPUT_DIR}${NC}"
echo

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Clean previous builds
echo -e "${BLUE}🧹 Cleaning previous builds...${NC}"
rm -f "${OUTPUT_DIR}"/wg-orbit-*

# Build function
build_binary() {
    local platform=$1
    local goos=$2
    local goarch=$3
    local goarm=$4
    
    local server_binary="${OUTPUT_DIR}/wg-orbit-server-${platform}"
    local client_binary="${OUTPUT_DIR}/wg-orbit-client-${platform}"
    
    # Add .exe extension for Windows
    if [[ "$goos" == "windows" ]]; then
        server_binary="${server_binary}.exe"
        client_binary="${client_binary}.exe"
    fi
    
    echo -e "${BLUE}🏗️  Building for ${platform}...${NC}"
    
    # Set environment variables
    export GOOS="$goos"
    export GOARCH="$goarch"
    if [[ -n "$goarm" ]]; then
        export GOARM="$goarm"
    else
        unset GOARM
    fi
    
    # Build server
    echo -e "   📦 Building server binary..."
    go build -ldflags="${LDFLAGS}" -o "$server_binary" ./cmd/server
    
    # Build client
    echo -e "   📦 Building client binary..."
    go build -ldflags="${LDFLAGS}" -o "$client_binary" ./cmd/client
    
    # Show file sizes
    if [[ "$goos" != "windows" ]]; then
        echo -e "   📊 Binary sizes:"
        ls -lh "$server_binary" "$client_binary" | awk '{print "      " $9 ": " $5}'
    fi
    
    echo -e "${GREEN}   ✅ ${platform} build completed${NC}"
    echo
}

# Build for all architectures
for arch_config in "${ARCHITECTURES[@]}"; do
    IFS=':' read -r platform goos goarch goarm <<< "$arch_config"
    build_binary "$platform" "$goos" "$goarch" "$goarm"
done

# Show summary
echo -e "${GREEN}🎉 All builds completed successfully!${NC}"
echo -e "${BLUE}📋 Built binaries:${NC}"
ls -lh "${OUTPUT_DIR}"/wg-orbit-* | awk '{print "   " $9 ": " $5}'
echo

# Create checksums
echo -e "${BLUE}🔐 Generating checksums...${NC}"
cd "${OUTPUT_DIR}"
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum wg-orbit-* > checksums.txt
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 wg-orbit-* > checksums.txt
else
    echo -e "${YELLOW}⚠️  No checksum utility found, skipping checksums${NC}"
fi
if [[ -f checksums.txt ]]; then
    echo -e "${GREEN}✅ Checksums saved to ${OUTPUT_DIR}/checksums.txt${NC}"
fi
cd ..

echo -e "${YELLOW}💡 Usage examples:${NC}"
echo -e "   # Linux AMD64: ./bin/wg-orbit-server-linux-amd64"
echo -e "   # Raspberry Pi 4: ./bin/wg-orbit-server-linux-arm64"
echo -e "   # Raspberry Pi 3: ./bin/wg-orbit-server-linux-armv7"
echo -e "   # macOS ARM: ./bin/wg-orbit-server-darwin-arm64"
echo