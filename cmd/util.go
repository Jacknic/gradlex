package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/martinlindhe/base36"
)

const GRADLE_USER_DEFAULT_DIR = ".gradle"
const GRADLE_HOME = "GRADLE_HOME"
const GRADLE_USER_HOME = "GRADLE_USER_HOME"
const GRADLE_DIST_PROXY = "GRADLE_DIST_PROXY"

type GradleConfig struct {
	GradleDistProxy string `json:"gradle_dist_proxy"`
	Language        string `json:"language"`
}

var config *GradleConfig
var configFilePath = getGradleUserHome() + "/gradlex_config.json"

func init() {
	config = getGradleConfig()
}

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
	if config.GradleDistProxy != "" {
		return config.GradleDistProxy
	}
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

// 获取 Gradle 配置
func getGradleConfig() *GradleConfig {
	gradleConfig := GradleConfig{}
	data, err := os.ReadFile(configFilePath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
	} else {
		json.Unmarshal(data, &gradleConfig)
	}
	return &gradleConfig
}

// 设置 Gradle 配置
func setGradleConfig(config *GradleConfig) {
	data, _ := json.Marshal(config)
	_ = os.WriteFile(configFilePath, data, 0644)
}

// copyDirectory recursively copies a directory from src to dst.
// It creates the destination directory if it does not already exist.
func copyDirectory(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
// It ensures the destination directory exists before writing the file.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyFiles copies multiple files given a slice of source/destination pairs.
func copyFiles(filePairs [][2]string) error {
	for _, pair := range filePairs {
		src := pair[0]
		dst := pair[1]
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// createMarkerFiles creates .lck and .ok marker files for a Gradle zip package.
func createMarkerFiles(targetDir, zipFileName string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	lckPath := filepath.Join(targetDir, zipFileName+".lck")
	if _, err := os.Create(lckPath); err != nil {
		return err
	}

	okPath := filepath.Join(targetDir, zipFileName+".ok")
	if _, err := os.Create(okPath); err != nil {
		return err
	}

	return nil
}
