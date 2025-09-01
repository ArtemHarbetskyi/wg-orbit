#!/bin/bash

# Native deployment script for wg-orbit
# Helps deploy the correct binary for the target system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BIN_DIR="bin"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/wg-orbit"

echo -e "${BLUE}🚀 wg-orbit Native Deployment Script${NC}"
echo

# Detect system architecture
detect_arch() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        linux)
            case "$arch" in
                x86_64|amd64)
                    echo "linux-amd64"
                    ;;
                aarch64|arm64)
                    echo "linux-arm64"
                    ;;
                armv7l)
                    echo "linux-armv7"
                    ;;
                armv6l)
                    echo "linux-armv6"
                    ;;
                *)
                    echo -e "${RED}❌ Unsupported Linux architecture: $arch${NC}"
                    exit 1
                    ;;
            esac
            ;;
        darwin)
            case "$arch" in
                x86_64)
                    echo "darwin-amd64"
                    ;;
                arm64)
                    echo "darwin-arm64"
                    ;;
                *)
                    echo -e "${RED}❌ Unsupported macOS architecture: $arch${NC}"
                    exit 1
                    ;;
            esac
            ;;
        *)
            echo -e "${RED}❌ Unsupported operating system: $os${NC}"
            exit 1
            ;;
    esac
}

# Show usage
show_usage() {
    echo -e "${YELLOW}Usage: $0 [OPTIONS]${NC}"
    echo -e "${YELLOW}Options:${NC}"
    echo -e "  --install     Install binaries to system (requires sudo)"
    echo -e "  --service     Install systemd service (Linux only)"
    echo -e "  --config      Create default configuration"
    echo -e "  --arch ARCH   Override architecture detection"
    echo -e "  --help        Show this help"
    echo
    echo -e "${YELLOW}Supported architectures:${NC}"
    echo -e "  linux-amd64, linux-arm64, linux-armv7, linux-armv6"
    echo -e "  darwin-amd64, darwin-arm64"
    echo -e "  windows-amd64"
    echo
}

# Install binaries
install_binaries() {
    local platform=$1
    local server_bin="${BIN_DIR}/wg-orbit-server-${platform}"
    local client_bin="${BIN_DIR}/wg-orbit-client-${platform}"
    
    # Add .exe for Windows
    if [[ "$platform" == *"windows"* ]]; then
        server_bin="${server_bin}.exe"
        client_bin="${client_bin}.exe"
    fi
    
    if [[ ! -f "$server_bin" ]] || [[ ! -f "$client_bin" ]]; then
        echo -e "${RED}❌ Binaries not found for platform: $platform${NC}"
        echo -e "${YELLOW}💡 Run './scripts/build-native.sh' first to build binaries${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}📦 Installing binaries for $platform...${NC}"
    
    # Copy binaries
    sudo cp "$server_bin" "${INSTALL_DIR}/wg-orbit-server"
    sudo cp "$client_bin" "${INSTALL_DIR}/wg-orbit-client"
    
    # Set permissions
    sudo chmod +x "${INSTALL_DIR}/wg-orbit-server"
    sudo chmod +x "${INSTALL_DIR}/wg-orbit-client"
    
    echo -e "${GREEN}✅ Binaries installed to ${INSTALL_DIR}${NC}"
}

# Create systemd service
create_service() {
    if [[ "$(uname -s)" != "Linux" ]]; then
        echo -e "${YELLOW}⚠️  Systemd service only supported on Linux${NC}"
        return
    fi
    
    echo -e "${BLUE}🔧 Creating systemd service...${NC}"
    
    sudo tee "${SERVICE_DIR}/wg-orbit.service" > /dev/null <<EOF
[Unit]
Description=WireGuard Orbit Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
ExecStart=${INSTALL_DIR}/wg-orbit-server run
Restart=always
RestartSec=5
Environment=WG_ORBIT_CONFIG=${CONFIG_DIR}/server.yaml

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${CONFIG_DIR}

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    echo -e "${GREEN}✅ Systemd service created${NC}"
    echo -e "${YELLOW}💡 Enable with: sudo systemctl enable wg-orbit${NC}"
    echo -e "${YELLOW}💡 Start with: sudo systemctl start wg-orbit${NC}"
}

# Create default configuration
create_config() {
    echo -e "${BLUE}📝 Creating default configuration...${NC}"
    
    sudo mkdir -p "$CONFIG_DIR"
    
    # Server config
    sudo tee "${CONFIG_DIR}/server.yaml" > /dev/null <<EOF
# wg-orbit server configuration
server:
  interface: wg0
  listen_port: 51820
  api_port: 8080
  subnet: 10.0.0.0/24
  
database:
  type: sqlite
  path: ${CONFIG_DIR}/wg-orbit.db
  
auth:
  jwt_secret: "change-this-secret-key"
  token_ttl: 12h
EOF
    
    # Client config template
    sudo tee "${CONFIG_DIR}/client.yaml.example" > /dev/null <<EOF
# wg-orbit client configuration template
client:
  server_url: https://your-server:8080
  token: "your-enrollment-token"
  interface: wg0
  
logging:
  level: info
EOF
    
    sudo chown -R root:root "$CONFIG_DIR"
    sudo chmod 600 "${CONFIG_DIR}"/*.yaml*
    
    echo -e "${GREEN}✅ Configuration created in ${CONFIG_DIR}${NC}"
    echo -e "${YELLOW}⚠️  Remember to change the JWT secret in server.yaml${NC}"
}

# Main logic
PLATFORM=""
INSTALL=false
SERVICE=false
CONFIG=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --install)
            INSTALL=true
            shift
            ;;
        --service)
            SERVICE=true
            shift
            ;;
        --config)
            CONFIG=true
            shift
            ;;
        --arch)
            PLATFORM="$2"
            shift 2
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

# Auto-detect platform if not specified
if [[ -z "$PLATFORM" ]]; then
    PLATFORM=$(detect_arch)
fi

echo -e "${BLUE}🎯 Target platform: ${PLATFORM}${NC}"
echo

# Execute requested actions
if [[ "$INSTALL" == true ]]; then
    install_binaries "$PLATFORM"
fi

if [[ "$CONFIG" == true ]]; then
    create_config
fi

if [[ "$SERVICE" == true ]]; then
    create_service
fi

# Show next steps if no actions were taken
if [[ "$INSTALL" == false ]] && [[ "$CONFIG" == false ]] && [[ "$SERVICE" == false ]]; then
    echo -e "${YELLOW}💡 Available binaries for ${PLATFORM}:${NC}"
    ls -lh "${BIN_DIR}"/wg-orbit-*-"${PLATFORM}"* 2>/dev/null || echo -e "${RED}❌ No binaries found for ${PLATFORM}${NC}"
    echo
    echo -e "${YELLOW}💡 Quick deployment:${NC}"
    echo -e "  $0 --install --config --service"
    echo
    show_usage
fi

echo -e "${GREEN}🎉 Deployment completed!${NC}"