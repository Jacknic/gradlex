/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(linkCmd)
}

// 通过复制已有版本链接到指定版本
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "link exist gradle version gradlex link /root/.gradle/wrapper/dists/gradle-8.8-all/6gdy1pgp427xkqcjbxw3ylt6h https://services.gradle.org/distributions/gradle-8.3-bin.zip",
	Long:  `link exist gradle version`,
	Run: func(cmd *cobra.Command, args []string) {
		fromPath := args[0]
		toUrl := args[1]
		// fromPath := "/root/.gradle/wrapper/dists/gradle-8.8-all/6gdy1pgp427xkqcjbxw3ylt6h"
		// toUrl := "https://services.gradle.org/distributions/gradle-8.3-bin.zip"
		fromVersionType := path.Dir(fromPath)
		fromVersion := path.Base(fromVersionType)

		fileName := path.Base(toUrl)
		fileNameDir := strings.TrimSuffix(fileName, ".zip")
		linkMd5 := getLinkMd5(toUrl)

		fromPathPack := path.Join(fromPath, fromVersion[:len(fromVersion)-4])
		fromPathLck := path.Join(fromPath, fromVersion+".zip.lck")
		fromPathOk := path.Join(fromPath, fromVersion+".zip.ok")

		targetDir := getGradleUserHome() + "/wrapper/dists/" + fileNameDir + "/" + linkMd5
		os.MkdirAll(targetDir, os.ModePerm)

		toPathLck := targetDir + "/" + fileNameDir + ".zip.lck"
		toPathOk := targetDir + "/" + fileNameDir + ".zip.ok"
		toPathPack := targetDir + "/" + fileNameDir[:len(fileNameDir)-4]
		// 复制文件夹内容
		copyDirectory(fromPathPack, toPathPack)
		copyFile(fromPathLck, toPathLck)
		copyFile(fromPathOk, toPathOk)
		log.Println("copy ", "\n", fromPathPack, "=>", toPathPack, "\n", fromPathLck, "=>", toPathLck, "\n", fromPathOk, "=>", toPathOk)
	},
}

// 复制文件夹内容
func copyDirectory(src, dst string) error {
	// 检查目标目录是否存在，如果不存在则创建
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}
	}

	// 遍历源目录中的所有文件和子目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 如果是目录，则递归复制
		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 如果是文件，则复制文件内容
			copyFile(srcPath, dstPath)
		}
	}

	return nil
}

// 复制文件内容
func copyFile(src, dst string) error {
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

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}
