Param(
  [string]$InstallDir = $null
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
$Api = "https://api.github.com/repos/$Repo/releases/latest"

Write-Host "Querying latest release..."
$release = Invoke-RestMethod -UseBasicParsing -Uri $Api
$asset = $release.assets | Where-Object { $_.browser_download_url -match 'windows-amd64' } | Select-Object -First 1
if (-not $asset) {
    Write-Error "No windows-amd64 asset found in latest release"
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
if (Test-Path $InstallDir) {
    $existingPath = Join-Path $InstallDir 'gradlex.exe'
    if (Test-Path $existingPath) {
        try {
            $existingVersion = & $existingPath version 2>&1 | Select-String "Version:" | ForEach-Object { $_.ToString().Split(':')[1].Trim() }
            Write-Host "检测到已安装版本: $existingVersion"
            Write-Host "将覆盖旧版本..."
        } catch {
            Write-Host "无法读取已安装版本信息，继续安装..."
        }
    }
}

Write-Host "Installing to $InstallDir"
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item -Path $bin.FullName -Destination (Join-Path $InstallDir 'gradlex.exe') -Force

Write-Host "Installed gradlex to $InstallDir"
