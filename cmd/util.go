package cmd

import (
	"crypto/md5"
	"os"
	"strings"

	"github.com/martinlindhe/base36"
)

const GRADLE_USER_DEFAULT_DIR = ".gradle"
const GRADLE_HOME = "GRADLE_HOME"
const GRADLE_USER_HOME = "GRADLE_USER_HOME"
const GRADLE_DIST_PROXY = "GRADLE_DIST_PROXY"

// 获取 Gradle 用户目录
func getGradleUserHome() string {
	home, exists := os.LookupEnv(GRADLE_USER_HOME)
	if !exists || home == "" {
		userHome, _ := os.UserHomeDir()
		home = userHome + "/" + GRADLE_USER_DEFAULT_DIR
	}

	return home
}

// 获取 Gradle 安装目录
func getGradleHome() string {
	return os.Getenv(GRADLE_HOME)
}

// 获取 Gradle 代理地址
func getGradleDistProxy() string {
	return os.Getenv(GRADLE_DIST_PROXY)
}

// 获取链接的 md5 值
func getLinkMd5(link string) string {
	hasher := md5.New()
	hasher.Write([]byte(link))
	hash := hasher.Sum(nil)
	md5Hash := base36.EncodeBytes(hash)
	return strings.ToLower(md5Hash)
}
