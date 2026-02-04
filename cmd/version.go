package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 这些变量由编译时通过 ldflags 注入
var (
	// Version 是软件版本号
	Version = "dev"
	// GitTag 是Git标签
	GitTag = ""
	// GitCommit 是Git提交hash
	GitCommit = ""
	// BuildTime 是构建时间
	BuildTime = ""
)

func init() {
	versionCmd.Aliases = []string{"v"}
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  `gradlex version https://github.com/Jacknic/gradlex `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: " + Version)
		fmt.Println("Git Tag: " + GitTag)
		fmt.Println("Git Commit: " + GitCommit)
		fmt.Println("Build Time: " + BuildTime)
		
		// 检查更新提示（非开发版本）
		if Version != "dev" {
			checkUpdateInBackground()
		}
	},
}

// checkUpdateInBackground 在后台检查更新
func checkUpdateInBackground() {
	latestRelease, err := getLatestRelease(false)
	if err != nil {
		// 静默失败，不影响版本输出
		return
	}

	// 比较版本，使用 CompareVersions 进行语义化版本比较
	if CompareVersions(Version, latestRelease.TagName) < 0 {
		fmt.Printf("\n发现新版本: %s (当前: %s)\n", latestRelease.TagName, Version)
		fmt.Printf("使用 'gradlex update' 进行升级\n")
	}
}
