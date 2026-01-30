Param(
  [string]$InstallDir = "C:\\Program Files\\gradlex"
)

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

Write-Host "Installing to $InstallDir"
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item -Path $bin.FullName -Destination (Join-Path $InstallDir 'gradlex.exe') -Force

Write-Host "Installed gradlex to $InstallDir"
