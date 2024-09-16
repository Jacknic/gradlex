# Gradlex 

## 简介

Gradlex 是一个用于 Gradle 下载的工具，可以通过镜像配置加速 Gradle 下载。

## 下载安装

下载地址：https://github.com/Jacknic/gradlex/releases

## 配置

### 将gradlex 添加到 PATH 中
如果你想在任何路径下都能使用 gradlex 命令，你需要将 gradlex 添加到 PATH 中，否则你只能在 gradlex 文件所在位置打开命令行窗口。

### 设置镜像地址环境变量

你需要设置镜像环境变量 `GRADLE_DIST_PROXY` ，否则 Gradle 包会从官方或原地址下载，而不是镜像地址也就没有加速效果。

国内常用镜像地址列表：
1. 腾讯云 https://mirrors.cloud.tencent.com/gradle/
2. 阿里云 https://mirrors.aliyun.com/macports/distfiles/gradle/



- Windows 环境

为了方便使用，建议在 **系统环境变量** 中设置环境变量，下方的命令行设置仅对当前窗口有效。

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

只要你设置了环境变量 `GRADLE_DIST_PROXY` ，你就可以从任意地址安装 Gradle 了。

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