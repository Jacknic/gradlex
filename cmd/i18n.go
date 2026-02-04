package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var currentLanguage string
var langInitOnce sync.Once

// 支持的语言列表
const (
	LangZH = "zh" // 中文
	LangEN = "en" // 英文
)

// initLanguage 初始化语言设置（只执行一次）
func initLanguage() {
	langInitOnce.Do(func() {
		// 优先从环境变量读取
		if lang := os.Getenv("GRADLEX_LANG"); lang != "" {
			setLanguage(lang)
			return
		}

		// 尝试从配置文件读取
		if config != nil && config.Language != "" {
			setLanguage(config.Language)
			return
		}

		// 尝试从系统环境变量获取
		if lang := os.Getenv("LANG"); lang != "" {
			if strings.HasPrefix(lang, "zh") {
				setLanguage(LangZH)
			} else {
				setLanguage(LangEN)
			}
			return
		}

		// 默认使用中文
		setLanguage(LangZH)
	})
}

// setLanguage 设置当前语言
func setLanguage(lang string) {
	// 如果不是已知语言，默认使用英文并输出警告
	// 注意：不能在这里调用 T()，因为 T() 会调用 initLanguage()，可能导致死锁
	var normalizedLang string
	switch strings.ToLower(lang) {
	case LangZH, "zh-cn", "zh_cn":
		currentLanguage = LangZH
		normalizedLang = LangZH
	case LangEN, "en-us", "en_us":
		currentLanguage = LangEN
		normalizedLang = LangEN
	default:
		if lang != "" {
			fmt.Printf("Unsupported language: %s, supported languages: zh, en\n", lang)
		}
		currentLanguage = LangEN
		normalizedLang = LangEN
	}

	// 持久化到配置文件（如果是从命令行参数设置的）
	if config != nil && config.Language != normalizedLang {
		config.Language = normalizedLang
		setGradleConfig(config)
	}
}

// GetLanguage 获取当前语言
func GetLanguage() string {
	initLanguage()
	return currentLanguage
}

// T 翻译函数，返回对应语言的文本
func T(key, defaultValue string) string {
	initLanguage()
	translations := getTranslations()
	if text, ok := translations[key]; ok {
		return text
	}
	return defaultValue
}

// getTranslations 获取当前语言的翻译映射
func getTranslations() map[string]string {
	if currentLanguage == LangZH {
		return zhTranslations
	}
	return enTranslations
}

// 中文翻译映射
var zhTranslations = map[string]string{
	// Install 命令
	"install.url_invalid":        "URL 无效: %s",
	"install.use_proxy":          "使用代理: %s",
	"install.already_installed":   "Gradle %s-%s 已经安装，跳过下载",
	"install.force_hint":         "如需重新下载，请使用 -f/--force 参数",
	"install.install_dir":         "安装目录: %s",
	"install.version_required":   "请指定 Gradle 版本，使用 -v 参数",

	// Wrapper 命令
	"wrapper.not_found":          "错误: 未找到 gradle-wrapper.properties 文件 (工作目录: %s)",
	"wrapper.no_distribution":    "错误: 未找到 distributionUrl 配置",
	"wrapper.parse_error":        "无法从 URL 提取版本信息: %s",

	// Update 命令
	"update.dev_version":        "当前是开发版本，无法检查更新",
	"update.current_version":     "当前版本: %s",
	"update.latest_version":      "最新版本: %s",
	"update.uptodate":          "已经是最新版本！",
	"update.checking":           "\n开始升级...",
	"update.done":              "\n升级完成！请重启 gradlex 以使用新版本。",
	"update.error":             "升级失败: %w",
	"update.fetch_failed":       "获取最新版本失败: %w",

	// 版本相关
	"version.new_version":       "\n发现新版本: %s (当前: %s)\n",
	"version.update_hint":        "使用 'gradlex update' 进行升级\n",

	// 通用
	"version_flag":             "Gradle 版本",
	"type_flag":               "Gradle 类型",
	"url_flag":                "下载版本",
	"force_flag":              "强制下载，即使已安装该版本",
	"check_flag":              "仅检查更新，不执行升级",
	"pre_release_flag":         "包含预发布版本",
	"script_flag":             "指定安装脚本路径（默认自动检测）",
	"path_flag":               "工作目录",
	"work_dir_default":         ".",

	// 语言相关
	"lang_flag":               "语言 (zh/en)",
	"unsupported_language":     "不支持的语言: %s，支持的语言: zh, en",

	// Root 命令
	"root.short":              "gradlex 是一个 gradle 下载工具",
	"root.usage":              "用法:",
	"root.available_commands":  "可用命令:",
	"root.flags":               "标志:",
	"root.use_info":            "使用 \"gradlex [command] --help\" 查看更多命令信息",

	// 子命令
	"completion.short":        "生成指定 shell 的自动补全脚本",
	"completion.long":         "为指定的 shell（bash, zsh, fish, powershell）生成自动补全脚本",
	"env.short":               "打印环境变量或默认配置",
	"env.long":                "显示 gradlex 相关的环境变量和配置信息",
	"help.short":              "查看任意命令的帮助信息",
	"help.long":               "显示指定命令的详细帮助信息",
	"install.short":           "安装 gradle",
	"install.long":            "安装 Gradle 到本地缓存。如果版本已存在则跳过下载（使用 -f 强制）。",
	"link.short":              "链接已存在的 gradle 版本",
	"link.long":               "将已存在的 Gradle 版本链接到本地缓存，避免重复下载",
	"list.short":              "显示 gradle 版本列表",
	"list.long":               "列出可用的 Gradle 版本",
	"local.short":             "列出本地已安装的 gradle 版本",
	"local.long":              "显示本地已下载和缓存的 Gradle 版本列表",
	"proxy.short":             "列出、设置代理",
	"proxy.long":              "配置或查看 Gradle 下载代理设置",
	"update.short":            "检查并升级到最新版本",
	"update.long":             "检查 GitHub Releases 上的最新版本，如果发现新版本则执行升级操作。",
	"version.short":           "打印版本信息",
	"version.long":            "显示当前 gradlex 的版本号",
	"wrapper.short":           "解析 gradle-wrapper.properties 并下载 gradle",
	"wrapper.long":            "解析 gradle-wrapper.properties 文件提取 gradle 版本和下载 URL，然后下载并安装 gradle。如果版本已存在则跳过下载（使用 -f 强制）。",
}

// 英文翻译映射
var enTranslations = map[string]string{
	// Install command
	"install.url_invalid":        "URL invalid: %s",
	"install.use_proxy":          "use proxy: %s",
	"install.already_installed":   "Gradle %s-%s already installed, skip download",
	"install.force_hint":         "Use -f/--force flag to re-download",
	"install.install_dir":         "Install directory: %s",
	"install.version_required":   "Please specify Gradle version using -v flag",

	// Wrapper command
	"wrapper.not_found":          "Error: gradle-wrapper.properties file not found (working directory: %s)",
	"wrapper.no_distribution":    "Error: distributionUrl not found",
	"wrapper.parse_error":        "Cannot extract version information from URL: %s",

	// Update command
	"update.dev_version":        "Current version is dev, cannot check updates",
	"update.current_version":     "Current version: %s",
	"update.latest_version":      "Latest version: %s",
	"update.uptodate":          "Already up to date!",
	"update.checking":           "\nStarting upgrade...",
	"update.done":              "\nUpgrade completed! Please restart gradlex to use the new version.",
	"update.error":             "Upgrade failed: %w",
	"update.fetch_failed":       "Failed to fetch latest version: %w",

	// Version related
	"version.new_version":       "\nNew version available: %s (current: %s)\n",
	"version.update_hint":        "Use 'gradlex update' to upgrade\n",

	// Common flags
	"version_flag":             "Gradle version",
	"type_flag":               "Gradle type",
	"url_flag":                "Download URL",
	"force_flag":              "Force download even if version is already installed",
	"check_flag":              "Check updates only, do not perform upgrade",
	"pre_release_flag":         "Include pre-release versions",
	"script_flag":             "Specify installation script path (default auto-detect)",
	"path_flag":               "Working directory",
	"work_dir_default":         ".",

	// Language related
	"lang_flag":               "Language (zh/en)",
	"unsupported_language":     "Unsupported language: %s, supported languages: zh, en",

	// Root command
	"root.short":              "gradlex is a gradle download tool",
	"root.usage":              "Usage:",
	"root.available_commands":  "Available Commands:",
	"root.flags":               "Flags:",
	"root.use_info":            "Use \"gradlex [command] --help\" for more information about a command",

	// Subcommands
	"completion.short":        "Generate the autocompletion script for the specified shell",
	"completion.long":         "Generate autocompletion script for specified shell (bash, zsh, fish, powershell)",
	"env.short":               "Print env or default",
	"env.long":                "Display gradlex environment variables and configuration",
	"help.short":              "Help about any command",
	"help.long":               "Display detailed help information for a specified command",
	"install.short":           "install gradle",
	"install.long":            "Install Gradle to local cache. Skip download if version already exists (use -f to force).",
	"link.short":              "link exist gradle version",
	"link.long":               "Link existing Gradle version to local cache to avoid re-downloading",
	"list.short":              "show gradle version list",
	"list.long":               "List available Gradle versions",
	"local.short":             "list gradle local dists",
	"local.long":              "Display list of downloaded and cached Gradle versions",
	"proxy.short":             "list, ls, set",
	"proxy.long":              "Configure or view Gradle download proxy settings",
	"update.short":            "check and upgrade to latest version",
	"update.long":             "Check latest version on GitHub Releases and upgrade if new version is available.",
	"version.short":           "Print the version",
	"version.long":            "Display current gradlex version number",
	"wrapper.short":           "parse gradle-wrapper.properties and download gradle",
	"wrapper.long":            "Parse gradle-wrapper.properties to extract gradle version and download URL, then download and install gradle. Skip download if version already exists (use -f to force).",
}

// Printf 格式化输出并翻译
func TPrintf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key, ""), args...)
}

// getHelpTemplate 获取帮助模板（支持国际化）
func getHelpTemplate() string {
	usageText := T("root.usage", "Usage:")
	availableCommandsText := T("root.available_commands", "Available Commands:")
	flagsText := T("root.flags", "Flags:")
	useInfoText := T("root.use_info", "Use \"{{.CommandPath}} [command] --help\" for more information about a command.")

	return usageText + "\n" +
		"  {{.UseLine}}\n\n" +
		"{{if .HasAvailableSubCommands}}" + availableCommandsText + "\n" +
		"{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name \"help\"))}}\n  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}\n\n" +
		"{{end}}" +
		"{{if .HasAvailableLocalFlags}}" + flagsText + "\n" +
		"{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}\n\n" +
		"{{end}}" +
		"{{if .HasAvailableInheritedFlags}}" + flagsText + "\n" +
		"{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}\n\n" +
		"{{end}}" +
		"{{if .HasExample}}Examples:\n" +
		"{{.Example}}\n\n" +
		"{{end}}" +
		"{{if .HasAvailableSubCommands}}" + useInfoText + "\n{{end}}"
}
