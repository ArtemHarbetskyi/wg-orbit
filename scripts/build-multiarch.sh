#!/bin/bash

# Multi-architecture Docker build script for wg-orbit
# Supports AMD64, ARM64, and ARMv7 architectures

set -e

# Configuration
IMAGE_NAME="wg-orbit"
REGISTRY="ghcr.io/artem"
VERSION=${VERSION:-"latest"}
PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Building multi-architecture Docker images for wg-orbit${NC}"
echo -e "${YELLOW}Platforms: ${PLATFORMS}${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo

# Check if buildx is available
if ! docker buildx version > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker buildx is not available. Please install Docker Desktop or enable buildx.${NC}"
    exit 1
fi

# Create and use buildx builder if it doesn't exist
BUILDER_NAME="wg-orbit-builder"
if ! docker buildx ls | grep -q "$BUILDER_NAME"; then
    echo -e "${BLUE}📦 Creating buildx builder: $BUILDER_NAME${NC}"
    docker buildx create --name "$BUILDER_NAME" --driver docker-container --bootstrap
fi

echo -e "${BLUE}🔧 Using buildx builder: $BUILDER_NAME${NC}"
docker buildx use "$BUILDER_NAME"

# Inspect builder to ensure all platforms are supported
echo -e "${BLUE}🔍 Inspecting builder capabilities${NC}"
docker buildx inspect --bootstrap

# Build and push multi-arch images
echo -e "${GREEN}🏗️  Building multi-architecture images...${NC}"

# Build for all platforms
docker buildx build \
    --platform "$PLATFORMS" \
    --tag "${REGISTRY}/${IMAGE_NAME}:${VERSION}" \
    --tag "${REGISTRY}/${IMAGE_NAME}:latest" \
    --push \
    --progress=plain \
    .

echo -e "${GREEN}✅ Multi-architecture build completed successfully!${NC}"
echo -e "${YELLOW}📋 Images pushed to:${NC}"
echo -e "   ${REGISTRY}/${IMAGE_NAME}:${VERSION}"
echo -e "   ${REGISTRY}/${IMAGE_NAME}:latest"
echo
echo -e "${BLUE}🔍 To inspect the manifest:${NC}"
echo -e "   docker buildx imagetools inspect ${REGISTRY}/${IMAGE_NAME}:${VERSION}"
echo
echo -e "${BLUE}📱 Supported architectures:${NC}"
echo -e "   • linux/amd64 (Intel/AMD 64-bit)"
echo -e "   • linux/arm64 (ARM 64-bit - Raspberry Pi 4, Orange Pi 5)"
echo -e "   • linux/arm/v7 (ARM 32-bit - Raspberry Pi 3, Orange Pi PC)"