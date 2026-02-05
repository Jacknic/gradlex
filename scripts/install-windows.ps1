Param(
  [string]$InstallDir = $null,
  [switch]$IncludePrerelease
)

# 如果未指定安装目录，尝试自动检测已安装的 gradlex 位置
if (-not $InstallDir) {
    # 检查 PATH 中是否有 gradlex
    $inPath = Get-Command gradlex -ErrorAction SilentlyContinue
    if ($inPath) {
        $InstallDir = Split-Path $inPath.Source -Parent
        Write-Host "检测到已安装的 gradlex 在: $InstallDir"
    } else {
        # 默认安装目录
        $InstallDir = "C:\Program Files\gradlex"
    }
}

$Repo = 'Jacknic/gradlex'

# 根据是否包含预发布版本选择不同的 API 端点
if ($IncludePrerelease) {
    Write-Host "包含预发布版本..."
    $Api = "https://api.github.com/repos/$Repo/releases"
    Write-Host "Querying all releases..."
    $releases = Invoke-RestMethod -UseBasicParsing -Uri $Api
    $asset = $releases | ForEach-Object { $_.assets } | Where-Object { $_.browser_download_url -match 'windows-amd64' } | Select-Object -First 1
} else {
    $Api = "https://api.github.com/repos/$Repo/releases/latest"
    Write-Host "Querying latest release..."
    $release = Invoke-RestMethod -UseBasicParsing -Uri $Api
    $asset = $release.assets | Where-Object { $_.browser_download_url -match 'windows-amd64' } | Select-Object -First 1
}

if (-not $asset) {
    Write-Error "No windows-amd64 asset found in release"
    exit 1
}

$url = $asset.browser_download_url
$out = Join-Path $env:TEMP (Split-Path $url -Leaf)

Write-Host "Downloading $url to $out..."
Invoke-WebRequest -Uri $url -OutFile $out

$extractDir = Join-Path $env:TEMP 'gradlex_install'
if (Test-Path $extractDir) { Remove-Item $extractDir -Recurse -Force }
New-Item -ItemType Directory -Path $extractDir | Out-Null

Write-Host "Extracting..."
Expand-Archive -Path $out -DestinationPath $extractDir -Force

$bin = Get-ChildItem -Path $extractDir -Recurse -Filter 'gradlex.exe' -File | Select-Object -First 1
if (-not $bin) {
    Write-Error "gradlex.exe not found in the archive"
    exit 1
}

# 检查安装目录是否已存在
$existingVersion = $null
$existingPath = Join-Path $InstallDir 'gradlex.exe'
if (Test-Path $InstallDir) {
    if (Test-Path $existingPath) {
        try {
            $existingVersion = & $existingPath version 2>&1 | Select-String "Version:" | ForEach-Object { $_.ToString().Split(':')[1].Trim() }
            Write-Host "检测到已安装版本: $existingVersion"
            Write-Host "将覆盖旧版本..."
        } catch {
            Write-Host "无法读取已安装版本信息，继续安装..."
        }

        # 检查进程是否正在运行
        $process = Get-Process gradlex -ErrorAction SilentlyContinue
        if ($process) {
            Write-Host "检测到 gradlex 进程正在运行，尝试终止..."
            try {
                Stop-Process -Name gradlex -Force -ErrorAction Stop
                Start-Sleep -Seconds 2
                Write-Host "进程已终止"
            } catch {
                Write-Host "无法终止进程，请手动关闭 gradlex 后重试"
                exit 1
            }
        }
    }
}

Write-Host "Installing to $InstallDir"
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

# 尝试复制文件，如果被占用则重试
$maxRetries = 3
$retryDelay = 2
$success = $false

for ($i = 0; $i -lt $maxRetries; $i++) {
    try {
        Copy-Item -Path $bin.FullName -Destination $existingPath -Force -ErrorAction Stop
        $success = $true
        break
    } catch {
        $errorMessage = $_.Exception.Message
        if ($errorMessage -match "被.*进程使用|used by another process") {
            Write-Host "文件被占用，第 $($i + 1) 次尝试失败..."
            if ($i -lt $maxRetries - 1) {
                Start-Sleep -Seconds $retryDelay
            }
        } else {
            Write-Host "复制文件出错: $errorMessage"
            throw
        }
    }
}

if (-not $success) {
    Write-Error "无法复制文件，请确保 gradlex 未运行后重试"
    exit 1
}

Write-Host "Installed gradlex to $InstallDir"
