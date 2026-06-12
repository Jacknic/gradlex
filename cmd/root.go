package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var language string

func init() {
	rootCmd.PersistentFlags().StringVarP(&language, "lang", "l", "", "Language (zh/en)")

	// 在 init() 中初始化语言和模板，这样 --help 也能生效
	if language != "" {
		setLanguage(language)
	}
	initLanguage()

	// 更新 root 命令的帮助文本
	rootCmd.Short = T("root.short", "gradlex is a gradle download tool")
	if flag := rootCmd.PersistentFlags().Lookup("lang"); flag != nil {
		flag.Usage = T("lang_flag", "Language (zh/en)")
	}

	// 更新所有子命令的翻译
	for _, cmd := range rootCmd.Commands() {
		switch cmd.Name() {
		case "completion":
			cmd.Short = T("completion.short", "Generate the autocompletion script for the specified shell")
			cmd.Long = T("completion.long", "Generate autocompletion script for specified shell (bash, zsh, fish, powershell)")
		case "env":
			cmd.Short = T("env.short", "Print env or default")
			cmd.Long = T("env.long", "Display gradlex environment variables and configuration")
		case "help":
			cmd.Short = T("help.short", "Help about any command")
			cmd.Long = T("help.long", "Display detailed help information for a specified command")
		case "install":
			cmd.Short = T("install.short", "install gradle")
			cmd.Long = T("install.long", "Install Gradle to local cache. Skip download if version already exists (use -f to force).")
		case "link":
			cmd.Short = T("link.short", "link exist gradle version")
			cmd.Long = T("link.long", "Link existing Gradle version to local cache to avoid re-downloading")
		case "list":
			cmd.Short = T("list.short", "show gradle version list")
			cmd.Long = T("list.long", "List available Gradle versions")
		case "local":
			cmd.Short = T("local.short", "list gradle local dists")
			cmd.Long = T("local.long", "Display list of downloaded and cached Gradle versions")
		case "proxy":
			cmd.Short = T("proxy.short", "list, ls, set")
			cmd.Long = T("proxy.long", "Configure or view Gradle download proxy settings")
		case "update":
			cmd.Short = T("update.short", "check and upgrade to latest version")
			cmd.Long = T("update.long", "Check latest version on GitHub Releases and upgrade if new version is available.")
		case "version":
			cmd.Short = T("version.short", "Print the version")
			cmd.Long = T("version.long", "Display current gradlex version number")
		case "wrapper":
			cmd.Short = T("wrapper.short", "parse gradle-wrapper.properties and download gradle")
			cmd.Long = T("wrapper.long", "Parse gradle-wrapper.properties to extract gradle version and download URL, then download and install gradle. Skip download if version already exists (use -f to force).")
		}
	}

	// 设置自定义帮助模板以支持国际化
	rootCmd.SetHelpTemplate(getHelpTemplate())
}

var rootCmd = &cobra.Command{
	Use:   "gradlex",
	Short: "gradlex is a gradle download tool",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 如果用户通过 --lang 指定了语言，设置该语言并持久化
		if language != "" {
			setLanguage(language)
		}
		// 初始化 i18n，确保在命令参数解析后设置语言
		initLanguage()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
