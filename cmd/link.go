package cmd

import (
	"log"
	"os"
	"path"
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
		fromPath := strings.Replace(args[0], "\\", "/", -1)
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
		if err := copyDirectory(fromPathPack, toPathPack); err != nil {
			log.Fatalf("copy pack failed: %v", err)
		}
		if err := copyFiles([][2]string{{fromPathLck, toPathLck}, {fromPathOk, toPathOk}}); err != nil {
			log.Fatalf("copy marker files failed: %v", err)
		}
		log.Println("copy ", "\n", fromPathPack, "=>", toPathPack, "\n", fromPathLck, "=>", toPathLck, "\n", fromPathOk, "=>", toPathOk)
	},
}
