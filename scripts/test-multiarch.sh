#!/bin/bash

# Local multi-architecture Docker build test script for wg-orbit
# Tests building for different architectures without pushing to registry

set -e

# Configuration
IMAGE_NAME="wg-orbit"
VERSION="test-$(date +%Y%m%d-%H%M%S)"
PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Testing multi-architecture Docker build for wg-orbit${NC}"
echo -e "${YELLOW}Platforms: ${PLATFORMS}${NC}"
echo -e "${YELLOW}Test version: ${VERSION}${NC}"
echo

# Check if buildx is available
if ! docker buildx version > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker buildx is not available. Please install Docker Desktop or enable buildx.${NC}"
    exit 1
fi

# Create and use buildx builder if it doesn't exist
BUILDER_NAME="wg-orbit-test-builder"
if ! docker buildx ls | grep -q "$BUILDER_NAME"; then
    echo -e "${BLUE}📦 Creating test buildx builder: $BUILDER_NAME${NC}"
    docker buildx create --name "$BUILDER_NAME" --driver docker-container --bootstrap
fi

echo -e "${BLUE}🔧 Using buildx builder: $BUILDER_NAME${NC}"
docker buildx use "$BUILDER_NAME"

# Test build for each platform individually
echo -e "${GREEN}🏗️  Testing individual platform builds...${NC}"
echo

# Test AMD64
echo -e "${BLUE}Testing linux/amd64...${NC}"
docker buildx build \
    --platform linux/amd64 \
    --tag "${IMAGE_NAME}:${VERSION}-amd64" \
    --load \
    --progress=plain \
    .
echo -e "${GREEN}✅ AMD64 build successful${NC}"
echo

# Test ARM64
echo -e "${BLUE}Testing linux/arm64...${NC}"
docker buildx build \
    --platform linux/arm64 \
    --tag "${IMAGE_NAME}:${VERSION}-arm64" \
    --progress=plain \
    .
echo -e "${GREEN}✅ ARM64 build successful${NC}"
echo

# Test ARMv7
echo -e "${BLUE}Testing linux/arm/v7...${NC}"
docker buildx build \
    --platform linux/arm/v7 \
    --tag "${IMAGE_NAME}:${VERSION}-armv7" \
    --progress=plain \
    .
echo -e "${GREEN}✅ ARMv7 build successful${NC}"
echo

# Test multi-platform build (without push)
echo -e "${BLUE}Testing multi-platform build...${NC}"
docker buildx build \
    --platform "$PLATFORMS" \
    --tag "${IMAGE_NAME}:${VERSION}-multiarch" \
    --progress=plain \
    .
echo -e "${GREEN}✅ Multi-platform build successful${NC}"
echo

# Show loaded images
echo -e "${BLUE}📋 Locally available test images:${NC}"
docker images | grep "$IMAGE_NAME" | grep "$VERSION"
echo

# Cleanup test images
echo -e "${YELLOW}🧹 Cleaning up test images...${NC}"
docker rmi "${IMAGE_NAME}:${VERSION}-amd64" 2>/dev/null || true
echo

echo -e "${GREEN}✅ All multi-architecture builds completed successfully!${NC}"
echo -e "${BLUE}🎉 Your project is ready for deployment on:${NC}"
echo -e "   • Intel/AMD 64-bit systems"
echo -e "   • ARM 64-bit devices (Raspberry Pi 4, Orange Pi 5, etc.)"
echo -e "   • ARM 32-bit devices (Raspberry Pi 3, Orange Pi PC, etc.)"
echo
echo -e "${YELLOW}💡 To build and push to registry, use: ./scripts/build-multiarch.sh${NC}"