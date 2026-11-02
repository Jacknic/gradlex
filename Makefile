.PHONY: build build-with-version build-all clean help

# 获取版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell TZ=Asia/Shanghai date +'%Y%m%d %H:%M:%S')

# ldflags 用于编译时注入版本信息
LDFLAGS := -ldflags "\
	-s -w \
	-X 'github.com/jacknic/gradlex/cmd.Version=$(VERSION)' \
	-X 'github.com/jacknic/gradlex/cmd.GitTag=$(GIT_TAG)' \
	-X 'github.com/jacknic/gradlex/cmd.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/jacknic/gradlex/cmd.BuildTime=$(BUILD_TIME)'"

# 需要构建的目标平台，格式为 "os/arch"
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
# 构建产物输出目录
DIST_DIR ?= dist

help:
	@echo "Available targets:"
	@echo "  build              - Build with version info"
	@echo "  build-dev          - Build dev version"
	@echo "  build-all          - Cross compile for all platforms ($(PLATFORMS))"
	@echo "  clean              - Clean build artifacts"

build: build-with-version

build-with-version:
	@echo "Building with version info..."
	@echo "Version: $(VERSION)"
	@echo "Git Tag: $(GIT_TAG)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	go build $(LDFLAGS)

build-dev:
	@echo "Building dev version..."
	go build

# 交叉编译各个平台的可执行文件
build-all:
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		output="$(DIST_DIR)/gradlex-$$os-$$arch$$ext"; \
		echo "Building $$output ..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output; \
	done
	@echo "All platforms built into $(DIST_DIR)/"

clean:
	rm -f gradlex
	rm -f gradlex.exe
	rm -rf $(DIST_DIR)
