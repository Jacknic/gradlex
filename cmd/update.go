package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	repoOwner = "jacknic"
	repoName  = "gradlex"
)

// GitHubRelease 表示 GitHub Release API 返回的数据结构
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
}

// UpdateConfig 更新命令配置
type UpdateConfig struct {
	CheckOnly     bool
	Force         bool
	PreRelease    bool
	InstallScript string
}

var updateConfig = UpdateConfig{}

func init() {
	updateCmd.Aliases = []string{"upgrade", "u"}
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVarP(&updateConfig.CheckOnly, "check", "c", false, "Check updates only, do not perform upgrade")
	updateCmd.Flags().BoolVarP(&updateConfig.Force, "force", "f", false, "Force upgrade even if already up to date")
	updateCmd.Flags().BoolVarP(&updateConfig.PreRelease, "pre-release", "p", false, "Include pre-release versions")
	updateCmd.Flags().StringVar(&updateConfig.InstallScript, "script", "", "Specify installation script path (default auto-detect)")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "check and upgrade to latest version",
	Long:  "Check latest version on GitHub Releases and upgrade if new version is available.",
	Run: func(cmd *cobra.Command, args []string) {
		runUpdate(updateConfig)
	},
}

// runUpdate 执行更新流程
func runUpdate(config UpdateConfig) error {
	// 获取当前版本
	currentVersion := Version
	if currentVersion == "dev" && !config.Force {
		fmt.Println(T("update.dev_version", "当前是开发版本，无法检查更新"))
		return nil
	}

	fmt.Printf(T("update.current_version", "Current version: %s")+"\n", currentVersion)

	// 获取最新版本信息
	latestRelease, err := getLatestRelease(config.PreRelease)
	if err != nil {
		return fmt.Errorf(T("update.fetch_failed", "Failed to fetch latest version: %w"), err)
	}

	fmt.Printf(T("update.latest_version", "Latest version: %s")+"\n", latestRelease.TagName)

	// 比较版本，使用 CompareVersions 进行语义化版本比较
	if !config.Force && CompareVersions(currentVersion, latestRelease.TagName) >= 0 {
		fmt.Println(T("update.uptodate", "Already up to date!"))
		return nil
	}

	if !config.CheckOnly {
		// 执行升级
		fmt.Println(T("update.checking", "\nStarting upgrade..."))
		if err := performUpdate(config); err != nil {
			return fmt.Errorf(T("update.error", "Upgrade failed: %w"), err)
		}
		fmt.Println(T("update.done", "\nUpgrade completed! Please restart gradlex to use the new version."))
	}

	return nil
}

// getLatestRelease 获取 GitHub 最新 Release
func getLatestRelease(includePreRelease bool) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	if includePreRelease {
		// 如果包含预发布版本，获取所有 releases 并筛选
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", repoOwner, repoName)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回错误: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if includePreRelease {
		var releases []GitHubRelease
		if err := json.Unmarshal(body, &releases); err != nil {
			return nil, err
		}
		if len(releases) == 0 {
			return nil, fmt.Errorf("未找到任何 release")
		}
		return &releases[0], nil
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// performUpdate 执行升级操作
func performUpdate(config UpdateConfig) error {
	// 确定操作系统
	var scriptURL, scriptFile string
	switch runtime.GOOS {
	case "windows":
		scriptURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/scripts/install-windows.ps1", repoOwner, repoName)
		scriptFile = "install.ps1"
	default:
		scriptURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/scripts/install.sh", repoOwner, repoName)
		scriptFile = "install.sh"
	}

	// 如果指定了安装脚本，使用指定的脚本
	if config.InstallScript != "" {
		scriptFile = config.InstallScript
	} else {
		// 下载安装脚本
		fmt.Printf("下载安装脚本: %s\n", scriptURL)
		if err := downloadInstallScript(scriptURL, scriptFile); err != nil {
			return fmt.Errorf("下载安装脚本失败: %w", err)
		}
		defer os.Remove(scriptFile) // 执行完成后删除脚本
	}

	// 执行安装脚本
	fmt.Println("执行安装脚本...")
	var cmd *exec.Cmd
	var args []string

	switch runtime.GOOS {
	case "windows":
		// Windows 使用 PowerShell
		args = []string{"powershell", "-ExecutionPolicy", "Bypass", "-File", scriptFile}
		if config.PreRelease {
			args = append(args, "-IncludePrerelease")
		}
		cmd = exec.Command(args[0], args[1:]...)
	default:
		// Unix-like 系统使用 bash
		args = []string{"bash", scriptFile}
		if config.PreRelease {
			// 设置环境变量
			cmd = exec.Command("bash", scriptFile)
			cmd.Env = append(os.Environ(), "INCLUDE_PRERELEASE=true")
		} else {
			cmd = exec.Command("bash", scriptFile)
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行安装脚本失败: %w", err)
	}

	return nil
}

// downloadInstallScript 下载安装脚本（与 install.go 中的 downloadFile 不同，这个专门用于下载脚本）
func downloadInstallScript(url, filepath string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// CompareVersions 比较两个版本号
// 返回: -1 表示 v1 < v2, 0 表示 v1 == v2, 1 表示 v1 > v2
// 支持预发布版本号，如: 0.1.2-alpha01, 0.1.2-alpha02
func CompareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// 分离版本号和预发布标签
	v1Base, v1Prerelease := splitVersion(v1)
	v2Base, v2Prerelease := splitVersion(v2)

	// 比较基础版本号
	baseCompare := compareNumericVersion(v1Base, v2Base)
	if baseCompare != 0 {
		return baseCompare
	}

	// 基础版本相同，比较预发布标签
	return comparePrerelease(v1Prerelease, v2Prerelease)
}

// splitVersion 分离版本号和预发布标签
func splitVersion(version string) (base, prerelease string) {
	if idx := strings.Index(version, "-"); idx != -1 {
		return version[:idx], version[idx+1:]
	}
	return version, ""
}

// compareNumericVersion 比较数字版本号部分
func compareNumericVersion(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// comparePrerelease 比较预发布标签
// 空字符串（正式版本）大于任何预发布版本
func comparePrerelease(p1, p2 string) int {
	// 如果两个都没有预发布标签，相等
	if p1 == "" && p2 == "" {
		return 0
	}
	// 如果 p1 没有预发布标签（正式版本），p1 更大
	if p1 == "" {
		return 1
	}
	// 如果 p2 没有预发布标签（正式版本），p2 更大
	if p2 == "" {
		return -1
	}

	// 都有预发布标签，字符串比较
	if p1 < p2 {
		return -1
	}
	if p1 > p2 {
		return 1
	}

	return 0
}
