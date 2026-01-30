#!/bin/bash
# 编译脚本，支持编译时注入版本信息

set -e

# 获取版本信息
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(TZ=Asia/Shanghai date +'%Y%m%d %H:%M:%S')

echo "Building gradlex..."
echo "Version: $VERSION"
echo "Git Tag: $GIT_TAG"
echo "Git Commit: $GIT_COMMIT"
echo "Build Time: $BUILD_TIME"
echo ""

# 编译
go build -ldflags "\
	-X 'github.com/jacknic/gradlex/cmd.Version=$VERSION' \
	-X 'github.com/jacknic/gradlex/cmd.GitTag=$GIT_TAG' \
	-X 'github.com/jacknic/gradlex/cmd.GitCommit=$GIT_COMMIT' \
	-X 'github.com/jacknic/gradlex/cmd.BuildTime=$BUILD_TIME'" \
	-o gradlex

echo "Build complete! Binary: ./gradlex"
