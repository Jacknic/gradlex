package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var workDir string
var wrapperForce bool

func init() {
	wrapperCmd.Flags().StringVarP(&workDir, "path", "p", ".", "Working directory")
	wrapperCmd.Flags().BoolVarP(&wrapperForce, "force", "f", false, "Force download even if version is already installed")
	wrapperCmd.Aliases = []string{"w"}
	rootCmd.AddCommand(wrapperCmd)
}

var wrapperCmd = &cobra.Command{
	Use:   "wrapper",
	Short: "parse gradle-wrapper.properties and download gradle",
	Long:  "Parse gradle-wrapper.properties to extract gradle version and download URL, then download and install gradle. Skip download if version already exists (use -f to force).",
	Run: func(cmd *cobra.Command, args []string) {
		// 查找 gradle-wrapper.properties 文件
		wrapperPropertiesPath := findWrapperProperties(workDir)
		if wrapperPropertiesPath == "" {
			fmt.Printf(T("wrapper.not_found", "Error: gradle-wrapper.properties file not found (working directory: %s)")+"\n", workDir)
			return
		}

		log.Printf("Found gradle-wrapper.properties: %s\n", wrapperPropertiesPath)

		// 解析文件内容
		distributionUrl, err := parseWrapperProperties(wrapperPropertiesPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if distributionUrl == "" {
			fmt.Println(T("wrapper.no_distribution", "Error: distributionUrl not found"))
			return
		}

		log.Printf("Found distributionUrl: %s\n", distributionUrl)

		// 提取版本和类型信息
		re, _ := regexp.Compile(`gradle-(.+)-(all|bin)\.zip`)
		matches := re.FindStringSubmatch(distributionUrl)
		if len(matches) < 3 {
			fmt.Printf(T("wrapper.parse_error", "Cannot extract version information from URL: %s")+"\n", distributionUrl)
			return
		}

		version := matches[1]
		distType := matches[2]

		log.Printf("Extracted version: %s, type: %s\n", version, distType)

		// 执行下载和安装
		executeGradleInstall(version, distType, distributionUrl, wrapperForce)
	},
}

// findWrapperProperties 查找 gradle-wrapper.properties 文件
// 从指定目录开始往上层目录查找
func findWrapperProperties(startDir string) string {
	absStartDir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}

	currentDir := absStartDir
	for {
		wrapperPath := filepath.Join(currentDir, "gradle", "wrapper", "gradle-wrapper.properties")
		if _, err := os.Stat(wrapperPath); err == nil {
			return wrapperPath
		}

		// 移到上一级目录
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// 已经到达文件系统根目录
			break
		}
		currentDir = parentDir
	}

	return ""
}

// parseWrapperProperties 解析 gradle-wrapper.properties 文件
// 提取 distributionUrl 属性
func parseWrapperProperties(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 忽略注释和空行
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 查找 distributionUrl 属性
		if strings.HasPrefix(line, "distributionUrl") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				url := strings.TrimSpace(parts[1])
				// 移除末尾的换行符和可能的反斜杠
				url = strings.TrimSuffix(url, "\\")
				// properties 文件中可能对 ":" 进行了转义（例如 https\://...），去除所有反斜杠
				url = strings.ReplaceAll(url, "\\", "")
				return url, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取文件出错: %v", err)
	}

	return "", nil
}
