package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
			// 打印子文件夹列表，包含大小与完整性标记
			if fileInfo.IsDir() {
				parentDir := gradleDists + fileInfo.Name()
				distsList, err := os.ReadDir(parentDir)
				if err != nil {
					fmt.Println("read dists dir error:", err)
					return
				}
				for _, distsInfo := range distsList {
					if !distsInfo.IsDir() {
						continue
					}
					target := filepath.Join(parentDir, distsInfo.Name())
					absPath, _ := filepath.Abs(target)

					// 计算目录大小
					size := dirSize(target)

					// 判断 .ok / .lck 完整性
					zipFileName := fileInfo.Name() + ".zip"
					okPath := filepath.Join(target, zipFileName+".ok")
					lckPath := filepath.Join(target, zipFileName+".lck")
					status := "[INCOMPLETE]"
					if _, err := os.Stat(okPath); err == nil {
						status = "[OK]"
					} else if _, err := os.Stat(lckPath); err == nil {
						status = "[LOCK]"
					}

					fmt.Printf("%s\t%s\t%s\n", strings.Replace(absPath, "\\", "/", -1), humanSize(size), status)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
}

// dirSize 递归计算目录占用字节数
func dirSize(path string) int64 {
	var total int64
	filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

// humanSize 格式化字节为可读字符串
func humanSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	if size >= GB {
		return fmt.Sprintf("%.2fG", float64(size)/float64(GB))
	}
	if size >= MB {
		return fmt.Sprintf("%.2fM", float64(size)/float64(MB))
	}
	if size >= KB {
		return fmt.Sprintf("%.2fK", float64(size)/float64(KB))
	}
	return fmt.Sprintf("%dB", size)
}
