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
	"runtime"
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
	installCmd.Flags().StringVarP(&buildVersion, "version", "v", "", "Gradle version")
	installCmd.Flags().StringVarP(&buildType, "type", "t", "all", "Gradle type")
	installCmd.Flags().StringVarP(&zipUrl, "url", "u", "", "Download URL")
	installCmd.Flags().BoolVarP(&forceDownload, "force", "f", false, "Force download even if version is already installed")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install gradle",
	Long:  "Install Gradle to local cache. Skip download if version already exists (use -f to force).",
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Printf("args:%v \n", args)
		if len(args) == 1 {
			zipUrl = args[0]
		}
		if len(zipUrl) > 0 {
			re, _ := regexp.Compile(`gradle-(.+)-(all|bin)\.zip$`)
			infos := re.FindStringSubmatch(zipUrl)
			if len(infos) != 3 {
				fmt.Printf(T("install.url_invalid", "URL invalid: %s")+"\n", zipUrl)
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
// Performs the generic Gradle installation logic.
// 下载指定版本和类型（all 或 bin）的 Gradle 到本地缓存。
// Downloads the specified Gradle version and type (all or bin) to the local cache.
// 如果 force 为 false，则跳过已安装的版本。
// If force is false, skips download if the version is already installed.
// 对于 bin 类型请求，尝试从现有 all 发行版复制。
// For bin type requests, attempts to copy from existing all distribution if available.
// 对于 all 类型安装，自动复制到对应的 bin 路径。
// For all type installations, automatically copies to the corresponding bin path.
func executeGradleInstall(version, distType, downloadUrl string, force bool) {
	zipFileName := fmt.Sprintf("gradle-%s-%s.zip", version, distType)

	// link := "https://mirrors.cloud.tencent.com/gradle/" + zipFileName
	link := downloadUrl
	if len(getGradleDistProxy()) > 0 {
		fmt.Printf(T("install.use_proxy", "use proxy: %s")+"\n", getGradleDistProxy())
		link = getGradleDistProxy() + zipFileName
	}

	linkHash := getLinkMd5(link)
	downloadUrlHash := getLinkMd5(downloadUrl)
	targetDir := getGradleUserHome() + "/wrapper/dists/gradle-" + version + "-" + distType + "/" + downloadUrlHash

	// 检查是否已安装该版本
	if !force && isGradleInstalled(targetDir, zipFileName) {
		fmt.Printf(T("install.already_installed", "Gradle %s-%s already installed, skip download")+"\n", version, distType)
		fmt.Printf(T("install.force_hint", "Use -f/--force flag to re-download") + "\n")
		fmt.Printf(T("install.install_dir", "Install directory: %s")+"\n", targetDir)
		return
	}

	// 如果请求 bin 包但本地已有 all 包，则直接复制 all 到 bin，避免重新下载
	if distType == "bin" && !isGradleInstalled(targetDir, zipFileName) {
		if copyBinFromAllDist(version, targetDir) {
			return
		}
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
	if err := createMarkerFiles(targetDir, zipFileName); err != nil {
		log.Fatalf("create marker files failed: %v", err)
	}
	if distType == "all" {
		copyAllToBinDist(version, targetDir, downloadUrl)
	}
	log.Println("finish")
}

// isGradleInstalled 检查指定的 Gradle 版本是否已安装
// Checks if the specified Gradle version is already installed.
// 验证目标目录存在、.ok 完成标记文件存在且 gradle 包目录存在。
// Verifies the target directory exists, the .ok completion marker file is present,
// and the gradle package directory is present.
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

// copyBinFromAllDist 从现有 all 发行版复制到 bin 目标路径
// Copies an existing installed all distribution to the bin target path.
// 如果找到该版本的 all 发行版，则复制包目录并在 bin 目标目录中创建锁定和完成标记文件。
// If the all distribution for the version is found locally, it copies the package directory
// and creates lock and ok marker files in the bin target directory.
// 复制成功返回 true，否则返回 false。
// Returns true if copy succeeds, false otherwise.
func copyBinFromAllDist(version, targetDir string) bool {
	sourceDir, ok := findInstalledAllDistDir(version)
	if !ok {
		return false
	}

	// 检查目标 bin 版本是否已安装
	binZipFileName := fmt.Sprintf("gradle-%s-bin.zip", version)
	if isGradleInstalled(targetDir, binZipFileName) {
		log.Printf("Gradle %s-bin already installed at %s", version, targetDir)
		return true
	}

	fmt.Printf(T("install.copy_from_all", "Gradle %s-bin not found, copying existing all distribution to bin path")+"\n", version)

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		log.Printf("create bin target dir failed: %v", err)
		return false
	}

	sourceEntries, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Printf("read all source dir failed: %v", err)
		return false
	}

	var sourcePackDir string
	for _, entry := range sourceEntries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "gradle-") {
			sourcePackDir = filepath.Join(sourceDir, entry.Name())
			break
		}
	}
	if sourcePackDir == "" {
		log.Printf("no gradle package found in all distribution: %s", sourceDir)
		return false
	}

	destPackDir := filepath.Join(targetDir, filepath.Base(sourcePackDir))
	if err := copyDirectory(sourcePackDir, destPackDir); err != nil {
		log.Printf("copy all distribution to bin failed: %v", err)
		return false
	}

	if err := createMarkerFiles(targetDir, binZipFileName); err != nil {
		log.Printf("create marker files failed: %v", err)
		return false
	}

	log.Printf("copied existing all distribution from %s to %s", sourceDir, targetDir)
	return true
}

// findInstalledAllDistDir 查找本地缓存中指定版本的 all 发行版目录
// Searches for an installed all distribution directory
// 为指定的 Gradle 版本在本地缓存中查找 all 发行版目录。
// for the specified Gradle version in the local cache.
// 如果找到，返回目录路径和 true；否则返回空字符串和 false。
// Returns the path to the directory and true if found, empty string and false otherwise.
func findInstalledAllDistDir(version string) (string, bool) {
	baseDir := filepath.Join(getGradleUserHome(), "wrapper", "dists", fmt.Sprintf("gradle-%s-all", version))
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", false
	}

	zipFileName := fmt.Sprintf("gradle-%s-all.zip", version)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidateDir := filepath.Join(baseDir, entry.Name())
		if isGradleInstalled(candidateDir, zipFileName) {
			return candidateDir, true
		}
	}

	return "", false
}

// copyAllToBinDist 复制已安装的 all 发行版到对应的 bin 路径
// Copies an installed all distribution to the corresponding bin path.
// 将提供的 URL 中的 -all.zip 替换为 -bin.zip 以生成 bin URL。
// It replaces -all.zip with -bin.zip in the provided URL to generate the bin URL.
// 如果 bin 发行版已存在，返回 true 而不进行复制。
// If the bin distribution already exists, returns true without copying.
// 复制成功后创建锁定和完成标记文件。
// Creates lock and ok marker files after successful copy.
func copyAllToBinDist(version, allTargetDir, allDownloadUrl string) bool {
	// 将 all 类型的 URL 替换为 bin 类型
	binDownloadUrl := strings.ReplaceAll(allDownloadUrl, "-all.zip", "-bin.zip")
	binZipFileName := fmt.Sprintf("gradle-%s-bin.zip", version)
	binTargetDir := filepath.Join(getGradleUserHome(), "wrapper", "dists", fmt.Sprintf("gradle-%s-bin", version), getLinkMd5(binDownloadUrl))

	if isGradleInstalled(binTargetDir, binZipFileName) {
		return true
	}

	fmt.Printf(T("install.copy_all_to_bin", "Copying installed all distribution to %s")+"\n", binTargetDir)

	if err := os.MkdirAll(binTargetDir, os.ModePerm); err != nil {
		log.Printf("create bin target dir failed: %v", err)
		return false
	}

	sourceEntries, err := os.ReadDir(allTargetDir)
	if err != nil {
		log.Printf("read all target dir failed: %v", err)
		return false
	}

	var sourcePackDir string
	for _, entry := range sourceEntries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "gradle-") {
			sourcePackDir = filepath.Join(allTargetDir, entry.Name())
			break
		}
	}
	if sourcePackDir == "" {
		log.Printf("no gradle package found in all distribution dir: %s", allTargetDir)
		return false
	}

	destPackDir := filepath.Join(binTargetDir, filepath.Base(sourcePackDir))
	if err := copyDirectory(sourcePackDir, destPackDir); err != nil {
		log.Printf("copy all distribution to bin failed: %v", err)
		return false
	}

	if err := createMarkerFiles(binTargetDir, binZipFileName); err != nil {
		log.Printf("create bin marker files failed: %v", err)
		return false
	}

	log.Printf("copied all distribution from %s to %s", allTargetDir, binTargetDir)
	return true
}

// downloadFile 从指定的 URL 下载文件到目标文件路径
// Downloads a file from the specified URL to the target file path.
// 设置适当的 HTTP 头部并记录下载进度和速度。
// It sets appropriate HTTP headers and logs download progress and speed.
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

// unzip 将 zip 文件解压到目标目录
// Extracts a zip file to the destination directory.
// 根据需要创建目录并保留文件权限。
// It creates directories as needed and preserves file permissions.
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
			if err := os.MkdirAll(filePath, dirMode(file)); err != nil {
				return err
			}
			continue
		}

		// 如果是文件，确保父目录存在后再解压
		if err := os.MkdirAll(filepath.Dir(filePath), dirMode(file)); err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}

		// 创建目标文件
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode(file))
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

	// 按平台为 Gradle 命令入口及 gradlew 脚本赋予可执行权限
	setGradleExecutable(dest)

	return nil
}

// fileMode 返回解压文件的 Unix 权限位，类 Unix 平台优先采用 zip 头中记录的模式。
func fileMode(file *zip.File) os.FileMode {
	if runtime.GOOS == "windows" {
		return 0644
	}
	if mode := file.Mode(); mode != 0 {
		return mode.Perm()
	}
	return 0644
}

// dirMode 返回解压目录的 Unix 权限位。
func dirMode(file *zip.File) os.FileMode {
	if runtime.GOOS == "windows" {
		return 0755
	}
	if mode := file.Mode(); mode != 0 {
		return mode.Perm()
	}
	return 0755
}

// setGradleExecutable 在非 Windows 平台为 Gradle 命令及 gradlew 脚本赋予可执行权限。
// 解压后的顶层目录为 gradle-<version>/，故以 bin/gradle 结尾的文件（不论版本前缀）作为 Gradle 入口。
func setGradleExecutable(dest string) {
	if runtime.GOOS == "windows" {
		return
	}

	err := filepath.WalkDir(dest, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == "gradlew" || strings.HasSuffix(path, filepath.Join("bin", "gradle")) {
			if chmodErr := os.Chmod(path, 0755); chmodErr != nil {
				log.Printf("set executable failed: %s: %v", path, chmodErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("set executable walk failed: %v", err)
	}
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
