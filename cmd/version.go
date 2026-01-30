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
	},
}
