.PHONY: gen docs build build-linux-arm64 build-linux-amd64 upx lint lint-fix test test-coverage clean help

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := pf
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -s -w -X 'github.com/PokeForum/PokeForum/internal/consts.Version=$(VERSION)' -X 'github.com/PokeForum/PokeForum/internal/consts.GitCommit=$(GIT_COMMIT)'

# ============================================================================
# 开发命令
# ============================================================================

## gen: 生成 Ent 代码
gen:
	go generate ./ent

## docs: 生成 Swagger 文档
docs:
	swag init -g cmd/server.go -o docs

## run: 运行开发服务器
run:
	go run main.go server --debug

# ============================================================================
# 代码质量
# ============================================================================

## lint: 运行代码检查
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

## lint-fix: 运行代码检查并自动修复
lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@golangci-lint run --fix ./...

## fmt: 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w -local github.com/PokeForum/PokeForum .

## vet: 运行 go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# ============================================================================
# 测试
# ============================================================================

## test: 运行所有测试
test:
	@echo "Running tests..."
	@go test -v -race ./...

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-short: 运行短测试（跳过长时间运行的测试）
test-short:
	@echo "Running short tests..."
	@go test -v -short ./...

# ============================================================================
# 构建
# ============================================================================

## build: 构建 Linux AMD64 版本并压缩
build: build-linux-amd64
	@echo "Compressing binary with upx..."
	@upx $(APP_NAME)-linux-amd64 || true

## build-linux-arm64: 构建 Linux ARM64 版本
build-linux-arm64:
	@echo "Building for Linux ARM64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(APP_NAME)-linux-arm64 -ldflags '$(LDFLAGS) -extldflags "-static"'

## build-linux-amd64: 构建 Linux AMD64 版本
build-linux-amd64:
	@echo "Building for Linux AMD64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux-amd64 -ldflags '$(LDFLAGS) -extldflags "-static"'

## build-darwin: 构建 macOS 版本
build-darwin:
	@echo "Building for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-darwin-amd64 -ldflags '$(LDFLAGS)'

## build-darwin-arm64: 构建 macOS ARM64 版本
build-darwin-arm64:
	@echo "Building for macOS ARM64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(APP_NAME)-darwin-arm64 -ldflags '$(LDFLAGS)'

## build-all: 构建所有平台版本
build-all: build-linux-amd64 build-linux-arm64 build-darwin build-darwin-arm64
	@echo "All builds completed!"

## upx: 压缩所有二进制文件
upx:
	@echo "Compressing binaries with upx..."
	@upx $(APP_NAME)-* || true

# ============================================================================
# 依赖管理
# ============================================================================

## deps: 下载依赖
deps:
	@echo "Downloading dependencies..."
	@go mod download

## deps-tidy: 整理依赖
deps-tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

## deps-verify: 验证依赖
deps-verify:
	@echo "Verifying dependencies..."
	@go mod verify

## deps-update: 更新所有依赖
deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# ============================================================================
# 工具安装
# ============================================================================

## install-tools: 安装开发工具
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed successfully!"

# ============================================================================
# 清理
# ============================================================================

## clean: 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(APP_NAME)-*
	@rm -f coverage.out coverage.html
	@echo "Clean completed!"

# ============================================================================
# CI/CD
# ============================================================================

## ci: 运行完整的 CI 流程
ci: deps-verify lint test build
	@echo "CI pipeline completed!"

## pre-commit: 提交前检查
pre-commit: fmt lint test-short
	@echo "Pre-commit checks passed!"

# ============================================================================
# 帮助
# ============================================================================

## help: 显示帮助信息
help:
	@echo "PokeForum Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
