# try Windows 安装脚本：安装 TUI（try）与 GUI（官方 fyne package 产物）
# 用法（PowerShell）:
#   irm https://raw.githubusercontent.com/loveloki/try/main/install.ps1 | iex
# 或:
#   .\install.ps1
#
# 环境变量:
#   TRY_INSTALL_GUI   设为 0 时跳过 try-gui（默认安装）
#   TRY_INSTALL_DIR   CLI 安装目录（默认 %LOCALAPPDATA%\Programs\try）

$ErrorActionPreference = "Stop"
$Repo = "loveloki/try"
$InstallGui = if ($env:TRY_INSTALL_GUI) { $env:TRY_INSTALL_GUI } else { "1" }
$InstallDir = if ($env:TRY_INSTALL_DIR) { $env:TRY_INSTALL_DIR } else {
    Join-Path $env:LOCALAPPDATA "Programs\try"
}
$StartMenuDir = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs"

function Get-LatestVersion {
    $rel = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $tag = $rel.tag_name
    if ($tag.StartsWith("v")) { return $tag.Substring(1) }
    return $tag
}

function Install-Cli([string]$Version) {
    $zipName = "try_windows_amd64.zip"
    $url = "https://github.com/$Repo/releases/download/v$Version/$zipName"
    $tmp = Join-Path $env:TEMP "try-install-$Version"
    New-Item -ItemType Directory -Force -Path $tmp | Out-Null
    $zipPath = Join-Path $tmp $zipName
    Write-Host "下载 try v$Version (windows/amd64)..."
    Invoke-WebRequest -Uri $url -OutFile $zipPath
    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item (Join-Path $tmp "try.exe") (Join-Path $InstallDir "try.exe") -Force
    Write-Host "✓ try 已安装到 $(Join-Path $InstallDir 'try.exe')"
}

function Install-Gui([string]$Version) {
    if ($InstallGui -eq "0") {
        Write-Host ""
        Write-Host "已跳过 try-gui（TRY_INSTALL_GUI=0）"
        return
    }
    $zipName = "try-gui_windows_amd64.zip"
    $url = "https://github.com/$Repo/releases/download/v$Version/$zipName"
    $tmp = Join-Path $env:TEMP "try-gui-install-$Version"
    New-Item -ItemType Directory -Force -Path $tmp | Out-Null
    $zipPath = Join-Path $tmp $zipName
    Write-Host "下载 GUI 官方包 $zipName..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $zipPath
    } catch {
        Write-Host "⚠  无 GUI 官方包，已跳过。"
        return
    }
    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
    $exeSrc = Join-Path $tmp "try-gui.exe"
    if (-not (Test-Path $exeSrc)) {
        Write-Host "GUI 包中缺少 try-gui.exe" -ForegroundColor Yellow
        return
    }
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $exeDst = Join-Path $InstallDir "try-gui.exe"
    Copy-Item $exeSrc $exeDst -Force

    New-Item -ItemType Directory -Force -Path $StartMenuDir | Out-Null
    $lnkPath = Join-Path $StartMenuDir "Try.lnk"
    $wsh = New-Object -ComObject WScript.Shell
    $shortcut = $wsh.CreateShortcut($lnkPath)
    $shortcut.TargetPath = $exeDst
    $shortcut.WorkingDirectory = $InstallDir
    $shortcut.Description = "Try - experiment directory manager"
    $shortcut.Save()

    Write-Host "✓ try-gui 已安装到 $exeDst"
    Write-Host "✓ 开始菜单快捷方式: $lnkPath"
}

function Ensure-UserPath([string]$Dir) {
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not $userPath) { $userPath = "" }
    $parts = $userPath -split ";" | Where-Object { $_ -ne "" }
    if ($parts -contains $Dir) { return }
    $newPath = if ($userPath) { "$userPath;$Dir" } else { $Dir }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$Dir;$env:Path"
    Write-Host "✓ 已将 $Dir 加入用户 PATH（新开终端生效）"
}

$version = Get-LatestVersion
Install-Cli $version
Install-Gui $version
Ensure-UserPath $InstallDir

$tryExe = Join-Path $InstallDir "try.exe"
if (Test-Path $tryExe) {
    Write-Host ""
    Write-Host "初始化配置（Shell 集成在 Windows 上可能不可用）..."
    try {
        & $tryExe install
    } catch {
        Write-Host "try install 提示: $_"
    }
}

Write-Host ""
Write-Host "完成。新开终端后可用:"
Write-Host "  try       # TUI"
Write-Host "  try-gui   # 或从开始菜单搜索 Try"
