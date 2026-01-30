.PHONY: build build-with-version clean help

# 获取版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell TZ=Asia/Shanghai date +'%Y%m%d %H:%M:%S')

# ldflags 用于编译时注入版本信息
LDFLAGS := -ldflags "\
	-X 'github.com/jacknic/gradlex/cmd.Version=$(VERSION)' \
	-X 'github.com/jacknic/gradlex/cmd.GitTag=$(GIT_TAG)' \
	-X 'github.com/jacknic/gradlex/cmd.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/jacknic/gradlex/cmd.BuildTime=$(BUILD_TIME)'"

help:
	@echo "Available targets:"
	@echo "  build              - Build with version info"
	@echo "  build-dev          - Build dev version"
	@echo "  clean              - Clean build artifacts"

build: build-with-version

build-with-version:
	@echo "Building with version info..."
	@echo "Version: $(VERSION)"
	@echo "Git Tag: $(GIT_TAG)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	go build $(LDFLAGS) -o gradlex

build-dev:
	@echo "Building dev version..."
	go build -o gradlex

clean:
	rm -f gradlex
