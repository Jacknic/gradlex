package cmd

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

var buildVersion string
var buildType string
var zipUrl string

func init() {
	installCmd.Aliases = []string{"i"}
	installCmd.Flags().StringVarP(&buildVersion, "version", "v", "", "Gradle 版本")
	installCmd.Flags().StringVarP(&buildType, "type", "t", "all", "Gradle 类型")
	installCmd.Flags().StringVarP(&zipUrl, "url", "u", "", "下载版本")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install gradle",
	Long:  `gradlex version $version https://github.com/Jacknic/gradlex `,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Printf("args:%v \n", args)
		if len(args) == 1 {
			zipUrl = args[0]
		}
		if len(zipUrl) > 0 {
			re, _ := regexp.Compile(`gradle-(.+)-(.+)\.zip$`)
			infos := re.FindStringSubmatch(zipUrl)
			if len(infos) != 3 {
				fmt.Println("URL invalid :", zipUrl)
				return
			}
			buildVersion = infos[1]
			buildType = infos[2]
		}
		zipFileName := fmt.Sprintf("gradle-%s-%s.zip", buildVersion, buildType)
		linkRaw := "https://services.gradle.org/distributions/" + zipFileName
		if len(zipUrl) > 0 {
			linkRaw = zipUrl
		}
		// link := "https://mirrors.cloud.tencent.com/gradle/" + zipFileName
		link := linkRaw
		if len(getGradleDistProxy()) > 0 {
			fmt.Println("use proxy: ", getGradleDistProxy())
			link = getGradleDistProxy() + zipFileName
		}
		timeStart := time.Now()

		linkHash := getLinkMd5(link)
		linkRawHash := getLinkMd5(linkRaw)
		zipFilePath := getGradleUserHome() + "/" + linkHash + ".zip"
		fmt.Println(linkRaw + " download from \n" + link + " to => " + zipFilePath)

		err := downloadFile(link, zipFilePath)
		if err != nil {
			panic(err)
		}

		// 解压zip文件到指定目录
		targetDir := getGradleUserHome() + "/wrapper/dists/gradle-" + buildVersion + "-" + buildType + "/" + linkRawHash
		fmt.Println("unzip to ", targetDir)
		unzip(zipFilePath, targetDir)
		fmt.Println("remove file:", zipFilePath)
		os.Remove(zipFilePath)
		os.Create(targetDir + "/" + zipFileName + ".lck")
		os.Create(targetDir + "/" + zipFileName + ".ok")
		fmt.Println("done: " + time.Since(timeStart).String())
	},
}

// 下载文件
func downloadFile(url string, filePath string) error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Proto = "HTTP/1.1"
	req.Header = map[string][]string{
		// "Accept-Encoding": {"gzip"},
		"User-Agent": {"curl/7.68.0"},
		"Accept":     {"*/*"},
	}
	// 发起HTTP GET请求
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查HTTP响应状态码
	if resp.StatusCode != http.StatusOK {
		return errors.New("HTTP请求失败: " + resp.Status)
	}

	fmt.Printf("download fileSize: %d\n", resp.ContentLength)
	os.MkdirAll(getGradleHome(), os.ModePerm)
	// 打开文件用于写入
	outputFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	
	// 将响应体数据复制到文件
	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		return err
	}
	// 写入成功
	return nil
}

// 解压zip文件
func unzip(src, dest string) error {
	// 打开zip文件
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// 创建目标目录
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	// 遍历zip文件中的每个文件和文件夹
	for _, file := range r.File {
		// 获取目标文件或文件夹的路径
		filePath := filepath.Join(dest, file.Name)

		// 如果是文件夹，则创建文件夹
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// 如果是文件，则解压文件
		rc, err := file.Open()
		if err != nil {
			return err
		}

		// 创建目标文件
		outFile, err := os.Create(filePath)
		if err != nil {
			rc.Close()
			return err
		}

		// 将文件内容复制到目标文件中
		_, err = io.Copy(outFile, rc)
		if err != nil {
			outFile.Close()
			rc.Close()
			return err
		}

		// 关闭文件和目录句柄
		outFile.Close()
		rc.Close()
	}

	return nil
}
