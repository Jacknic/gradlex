/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "list gradle local dists",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		gradleDists := getGradleUserHome() + "/wrapper/dists/"

		// 判断文件是否存在
		if _, err := os.Stat(gradleDists); os.IsNotExist(err) {
			fmt.Println("gradle home not exist")
			return
		}

		fileInfoList, err := os.ReadDir(gradleDists)
		if err != nil {
			fmt.Println("read dir error:", err)
			return
		}
		for _, fileInfo := range fileInfoList {
			// 打印子文件夹列表
			if fileInfo.IsDir() {
				distsList, err := os.ReadDir(gradleDists + fileInfo.Name())
				if err != nil {
					fmt.Println("read dists dir error:", err)
					return
				}
				for _, distsInfo := range distsList {
					path, _ := filepath.Abs(gradleDists + fileInfo.Name() + "/" + distsInfo.Name())
					fmt.Println(path)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
}
