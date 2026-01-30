# Gradlex 

## 简介

Gradlex 是一个用于 Gradle 下载的工具，可以通过镜像配置加速 Gradle 下载。

## 下载安装

下载地址：[![GitHub Release](https://img.shields.io/github/v/release/Jacknic/gradlex?include_prereleases&link=https%3A%2F%2Fgithub.com%2FJacknic%2Fgradlex%2Freleases)](https://github.com/Jacknic/gradlex/releases) 

累计下载: [![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/Jacknic/gradlex/total)](https://github.com/Jacknic/gradlex/releases) 


如果你想在任何路径下都能使用 gradlex 命令，你需要将 gradlex 添加到 PATH 中，否则你只能在 gradlex 文件所在位置打开命令行窗口。

### 使用安装脚本（推荐）

#### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/Jacknic/gradlex/master/scripts/install.sh | bash
```

#### Windows (PowerShell)

```powershell
Invoke-WebRequest -Uri https://raw.githubusercontent.com/Jacknic/gradlex/master/scripts/install-windows.ps1 -OutFile install.ps1; .\install.ps1
```

### 手动命令行下载

#### Linux

```bash
# 下载并解压最新 Linux amd64 release
url=$(curl -s https://api.github.com/repos/Jacknic/gradlex/releases/latest \
	| grep "browser_download_url" | grep "linux-amd64" | head -n1 | cut -d '"' -f4)
curl -L "$url" -o gradlex-linux-amd64.tar.gz
tar -xf gradlex-linux-amd64.tar.gz
rm gradlex-linux-amd64.tar.gz
sudo mv gradlex /usr/local/bin/
```

#### macOS

```bash
# 下载并解压最新 macOS (amd64) release
url=$(curl -s https://api.github.com/repos/Jacknic/gradlex/releases/latest \
	| grep "browser_download_url" | grep "darwin-amd64" | head -n1 | cut -d '"' -f4)
curl -L "$url" -o gradlex-darwin-amd64.tar.gz
tar -xf gradlex-darwin-amd64.tar.gz
rm gradlex-darwin-amd64.tar.gz
sudo mv gradlex /usr/local/bin/
```

##### M1/M2 芯片版本：

```bash
# 下载并解压最新 macOS (arm64) release
url=$(curl -s https://api.github.com/repos/Jacknic/gradlex/releases/latest \
	| grep "browser_download_url" | grep "darwin-arm64" | head -n1 | cut -d '"' -f4)
curl -L "$url" -o gradlex-darwin-arm64.tar.gz
tar -xf gradlex-darwin-arm64.tar.gz
rm gradlex-darwin-arm64.tar.gz
sudo mv gradlex /usr/local/bin/
```

#### Windows

PowerShell 中：

```powershell
# 使用 PowerShell 获取并下载最新 Windows amd64 release
$url = (Invoke-RestMethod -UseBasicParsing https://api.github.com/repos/Jacknic/gradlex/releases/latest).assets |
	Where-Object { $_.browser_download_url -match 'windows-amd64' } | Select-Object -First 1 -ExpandProperty browser_download_url
Invoke-WebRequest -Uri $url -OutFile gradlex-windows-amd64.zip
Expand-Archive -Path gradlex-windows-amd64.zip -DestinationPath .
Remove-Item gradlex-windows-amd64.zip
```

CMD 中：

```bat
:: 在 CMD 中使用 PowerShell 获取下载地址，然后用 curl 下载并解压
for /f "delims=" %u in ('powershell -Command "(Invoke-RestMethod https://api.github.com/repos/Jacknic/gradlex/releases/latest).assets | Where-Object {$_.browser_download_url -match 'windows-amd64'} | Select-Object -First 1 -ExpandProperty browser_download_url"') do (
	curl -L %u -o gradlex-windows-amd64.zip && tar -xf gradlex-windows-amd64.zip && del gradlex-windows-amd64.zip
)
```


## 配置

### 国内常用镜像地址列表

1. 腾讯云 [https://mirrors.cloud.tencent.com/gradle/](https://mirrors.cloud.tencent.com/gradle/)
2. 阿里云 [https://mirrors.aliyun.com/macports/distfiles/gradle/](https://mirrors.aliyun.com/macports/distfiles/gradle/)
3. 华为云 [https://mirrors.huaweicloud.com/gradle/](https://mirrors.huaweicloud.com/gradle/)


### 设置镜像地址-配置文件

目前内置了 3 个镜像地址，你可以通过 `gradlex proxy set 0` / `gradle p set 0 到2` 指定镜像地址。与下面的*环境变量*方式互斥，互斥优先级为：配置文件 > 环境变量 > 默认值。一般情况下设置这个就够了。使用 `gradlex proxy ls` 命令可以查看当前内置镜像地址列表。


### 设置镜像地址-环境变量

你需要设置镜像环境变量 `GRADLE_DIST_PROXY` ，否则 Gradle 包会从**官方或原地址**下载，而不是镜像地址也就没有加速效果。为了方便使用，建议在 **系统环境变量** 中设置环境变量，下方的命令行设置仅对当前窗口有效。

- Windows 环境

CMD 中

```bat
set GRADLE_DIST_PROXY="https://mirrors.cloud.tencent.com/gradle/"
```

PowerShell 中

```powershell
$env:GRADLE_DIST_PROXY="https://mirrors.cloud.tencent.com/gradle/"
```

- *nix 环境

```bash
export GRADLE_DIST_PROXY="https://mirrors.cloud.tencent.com/gradle/"
```

- 验证环境变量

```bash
gradlex env
```

## 使用

### 安装指定版本 Gradle

```bash
gradlex install -v 6.8.3
```

### 安装指定类型 Gradle 

默认下载 all ，你也可以通过 `-t` / `--type` 指定版本类型，支持 `bin` 和 `all` 两种类型。

```bash
gradlex i -v 8.2 -t bin
```

### 从指定地址安装 Gradle

只要你设置了环境变量 `GRADLE_DIST_PROXY` ，工具就会解析对应的版本信息，并从指定镜像地址下载对应 Gradle 安装包。项目依赖的 gradle 版本通常在 `gradle/wrapper/gradle-wrapper.properties` 文件中指定，如果对应版本无法下载，使用 `distributionUrl` 属性的链接就可以下载对应版本

```bash
gradlex i https://services.gradle.org/distributions/gradle-8.7-rc-3-bin.zip
```

### 查看 Gradle 版本列表

```bash
gradlex list
```

### 查看当前已安装的 Gradle 版本

```bash
gradlex local
```
