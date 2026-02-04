package cmd

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var buildVersion string
var buildType string
var zipUrl string
var forceDownload bool

func init() {
	installCmd.Aliases = []string{"i"}
	installCmd.Flags().StringVarP(&buildVersion, "version", "v", "", "Gradle 版本")
	installCmd.Flags().StringVarP(&buildType, "type", "t", "all", "Gradle 类型")
	installCmd.Flags().StringVarP(&zipUrl, "url", "u", "", "下载版本")
	installCmd.Flags().BoolVarP(&forceDownload, "force", "f", false, "强制下载，即使已安装该版本")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install gradle",
	Long:  `Install Gradle to local cache. Skip download if version already exists (use -f to force).`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Printf("args:%v \n", args)
		if len(args) == 1 {
			zipUrl = args[0]
		}
		if len(zipUrl) > 0 {
			re, _ := regexp.Compile(`gradle-(.+)-(all|bin)\.zip$`)
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

		executeGradleInstall(buildVersion, buildType, linkRaw, forceDownload)
	},
}

// executeGradleInstall 执行 Gradle 安装的通用逻辑
func executeGradleInstall(version, distType, downloadUrl string, force bool) {
	zipFileName := fmt.Sprintf("gradle-%s-%s.zip", version, distType)

	// link := "https://mirrors.cloud.tencent.com/gradle/" + zipFileName
	link := downloadUrl
	if len(getGradleDistProxy()) > 0 {
		fmt.Println("use proxy: ", getGradleDistProxy())
		link = getGradleDistProxy() + zipFileName
	}

	linkHash := getLinkMd5(link)
	downloadUrlHash := getLinkMd5(downloadUrl)
	targetDir := getGradleUserHome() + "/wrapper/dists/gradle-" + version + "-" + distType + "/" + downloadUrlHash

	// 检查是否已安装该版本
	if !force && isGradleInstalled(targetDir, zipFileName) {
		fmt.Printf("Gradle %s-%s 已经安装，跳过下载\n", version, distType)
		fmt.Printf("如需重新下载，请使用 -f/--force 参数\n")
		fmt.Printf("安装目录: %s\n", targetDir)
		return
	}

	zipFilePath := getGradleUserHome() + "/" + linkHash + ".zip"
	log.Println(downloadUrl + " download from \n" + link + " => save to " + zipFilePath)

	err := downloadFile(link, zipFilePath)
	if err != nil {
		panic(err)
	}

	// 解压zip文件到指定目录
	log.Println("unzip to ", targetDir)
	unzip(zipFilePath, targetDir)
	log.Println("remove file:", zipFilePath)
	os.Remove(zipFilePath)
	os.Create(targetDir + "/" + zipFileName + ".lck")
	os.Create(targetDir + "/" + zipFileName + ".ok")
	log.Println("finish")
}

// isGradleInstalled 检查指定的 Gradle 版本是否已安装
func isGradleInstalled(targetDir, zipFileName string) bool {
	// 检查目标目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return false
	}

	// 检查 .ok 文件是否存在（表示安装完成）
	okFile := targetDir + "/" + zipFileName + ".ok"
	if _, err := os.Stat(okFile); os.IsNotExist(err) {
		return false
	}

	// 检查是否有实际的 Gradle 内容
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return false
	}

	// 目录应该包含 gradle-x.y.z 子目录
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "gradle-") {
			return true
		}
	}

	return false
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
	if resp == nil {
		return errors.New("空的 HTTP 响应")
	}
	defer resp.Body.Close()

	// 检查HTTP响应状态码
	if resp.StatusCode != http.StatusOK {
		return errors.New("HTTP请求失败: " + resp.Status)
	}

	// fmt.Printf("download fileSize: %d\n", resp.ContentLength)
	os.MkdirAll(getGradleUserHome(), os.ModePerm)
	// 打开文件用于写入
	outputFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// 将响应体数据复制到文件
	startTime := time.Now()
	counter := &WriteCounter{Total: resp.ContentLength}
	_, err = io.Copy(outputFile, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}
	// 计算网络下载速度
	countSeconds := time.Since(startTime).Seconds()
	log.Printf("download speed: %.1fs %.2f MB/s\n", countSeconds, float64(counter.Total/1024/1024)/countSeconds)
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

type WriteCounter struct {
	Total    int64
	Download int64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Download += int64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// 如果 Total 不可用（例如服务器未返回 Content-Length），避免除以 0 或负数
	if wc.Total <= 0 {
		fmt.Printf("\rDownloading... %d bytes", wc.Download)
		return
	}

	if done := wc.Download == wc.Total; done {
		fmt.Printf("\r")
		log.Printf("Downloaded %d%%   %d/%d\n", wc.Download*100/wc.Total, wc.Download, wc.Total)
	} else {
		fmt.Printf("\rDownloading... %d%%   %d/%d", wc.Download*100/wc.Total, wc.Download, wc.Total)
	}
}
