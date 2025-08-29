# WireGuard Orbit Makefile

# Змінні
GO_VERSION := 1.21
APP_NAME := wg-orbit
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go змінні
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Директорії
BIN_DIR := bin
CMD_DIR := cmd
DIST_DIR := dist

# Бінарники
SERVER_BINARY := $(BIN_DIR)/wg-orbit-server
CLIENT_BINARY := $(BIN_DIR)/wg-orbit-client

# Флаги збірки
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker
DOCKER_IMAGE := wg-orbit
DOCKER_TAG := $(VERSION)

.PHONY: all build clean test deps docker help

# За замовчуванням
all: clean deps test build

# Допомога
help:
	@echo "WireGuard Orbit Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build server and client binaries"
	@echo "  server      - Build server binary only"
	@echo "  client      - Build client binary only"
	@echo "  test        - Run tests"
	@echo "  test-cover  - Run tests with coverage"
	@echo "  deps        - Download dependencies"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker      - Build Docker image"
	@echo "  docker-run  - Run with Docker Compose"
	@echo "  docker-stop - Stop Docker Compose"
	@echo "  lint        - Run linters"
	@echo "  fmt         - Format code"
	@echo "  install     - Install binaries to GOPATH/bin"
	@echo "  release     - Build release binaries for multiple platforms"
	@echo "  dev-setup   - Setup development environment"
	@echo "  help        - Show this help"

# Створення директорій
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Завантаження залежностей
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Збірка
build: $(BIN_DIR) server client

server: $(SERVER_BINARY)

$(SERVER_BINARY): $(BIN_DIR)
	@echo "Building server..."
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(SERVER_BINARY) ./$(CMD_DIR)/server

client: $(CLIENT_BINARY)

$(CLIENT_BINARY): $(BIN_DIR)
	@echo "Building client..."
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(CLIENT_BINARY) ./$(CMD_DIR)/client

# Тестування
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Лінтинг та форматування
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@which goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	goimports -w .

# Встановлення
install: build
	@echo "Installing binaries..."
	$(GOCMD) install $(LDFLAGS) ./$(CMD_DIR)/server
	$(GOCMD) install $(LDFLAGS) ./$(CMD_DIR)/client

# Очищення
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out coverage.html

# Docker
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Релізна збірка для різних платформ
release: $(DIST_DIR)
	@echo "Building release binaries..."
	# Linux AMD64
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-server-linux-amd64 ./$(CMD_DIR)/server
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-client-linux-amd64 ./$(CMD_DIR)/client
	# Linux ARM64
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-server-linux-arm64 ./$(CMD_DIR)/server
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-client-linux-arm64 ./$(CMD_DIR)/client
	# macOS AMD64
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-server-darwin-amd64 ./$(CMD_DIR)/server
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-client-darwin-amd64 ./$(CMD_DIR)/client
	# macOS ARM64 (Apple Silicon)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-server-darwin-arm64 ./$(CMD_DIR)/server
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/wg-orbit-client-darwin-arm64 ./$(CMD_DIR)/client
	@echo "Release binaries built in $(DIST_DIR)/"

# Налаштування середовища розробки
dev-setup:
	@echo "Setting up development environment..."
	# Встановлення інструментів розробки
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	$(GOCMD) install github.com/air-verse/air@latest
	# Створення директорій
	mkdir -p logs data
	@echo "Development environment setup complete!"
	@echo "Run 'make help' to see available commands"

# Швидкий запуск для розробки
dev-server: build
	@echo "Starting development server..."
	./$(SERVER_BINARY) run --config configs/server.yaml

dev-client: build
	@echo "Starting development client..."
	./$(CLIENT_BINARY) --help

# Автоматичне перезавантаження під час розробки (потребує air)
dev-watch:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

# Перевірка версії та інформації
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(shell go version)"

# Перевірка залежностей
check-deps:
	@echo "Checking for required tools..."
	@which go > /dev/null || (echo "Go is not installed" && exit 1)
	@which docker > /dev/null || (echo "Docker is not installed" && exit 1)
	@which docker-compose > /dev/null || (echo "Docker Compose is not installed" && exit 1)
	@echo "All required tools are available"

# Benchmark тести
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...