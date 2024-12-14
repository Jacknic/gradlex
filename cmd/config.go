package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	proxyCmd.Aliases = []string{"p"}
	rootCmd.AddCommand(proxyCmd)
}

// 默认代理列表
var proxyList = []string{
	"https://mirrors.cloud.tencent.com/gradle/",
	"https://mirrors.aliyun.com/macports/distfiles/gradle/",
	"https://mirrors.huaweicloud.com/gradle/",
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "list, ls, set",
	Long:  "list: list all proxy \nset: set proxy, index or url. set 0, set https://mirrors.cloud.tencent.com/gradle/",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}
		switch args[0] {
		case "ls":
			printProxyList()
		case "list":
			printProxyList()
		case "set":
			if i, err := strconv.Atoi(args[1]); err == nil {
				if i < len(proxyList) {
					config.GradleDistProxy = proxyList[i]
				}
			} else {
				config.GradleDistProxy = args[1]
			}
			setGradleConfig(config)
			fmt.Println("set proxy to", config.GradleDistProxy)
		case "unset":
			config.GradleDistProxy = ""
			setGradleConfig(config)
			fmt.Println("unset proxy")
		default:
			cmd.Help()
		}
	},
}

func printProxyList() {
	for i, proxy := range proxyList {
		fmt.Println(i, proxy)
	}
}
