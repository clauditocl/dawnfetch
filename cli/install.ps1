# dawnfetch windows installer
# usage:
#   # modern powershell
#   powershell -c "irm https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1 | iex"
#   # no -ExecutionPolicy Bypass required by default
#   # windows 7 / powershell 2 compatible (simple webclient download + run)
#   $f="$env:TEMP\dawnfetch-install.ps1";(New-Object Net.WebClient).DownloadFile('https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1',$f);powershell -File $f
# optional:
#   -Version 1.2.3
#   -ZipPath "C:\Users\<you>\Downloads\dawnfetch_1.2.3_windows_386.zip"
#   -InstallDir "C:\Users\<you>\AppData\Local\dawnfetch"

param(
  [string]$Version = "",
  [string]$ZipPath = "",
  [string]$InstallDir = "",
  [switch]$Force
)

$ErrorActionPreference = "Stop"

function Is-Blank([string]$Value) {
  if ($null -eq $Value) {
    return $true
  }
  return ([string]::IsNullOrEmpty($Value.Trim()))
}

function Write-Info([string]$Message) {
  Write-Host "[dawnfetch] $Message" -ForegroundColor Cyan
}

function Write-Good([string]$Message) {
  Write-Host "[dawnfetch] $Message" -ForegroundColor Green
}

function Write-Warn([string]$Message) {
  Write-Host "[dawnfetch] $Message" -ForegroundColor Yellow
}

function Enable-Tls12 {
  try {
    $current = [int][Net.ServicePointManager]::SecurityProtocol
    $target = ($current -bor 3072)
    [Net.ServicePointManager]::SecurityProtocol = [Enum]::ToObject([Net.SecurityProtocolType], $target)
  } catch {
    # continue; older hosts may still work if already configured
  }
}

function Download-Text([string]$Url) {
  $wc = New-Object Net.WebClient
  $wc.Headers.Add("User-Agent", "dawnfetch-installer")
  try {
    return $wc.DownloadString($Url)
  } finally {
    $wc.Dispose()
  }
}

function Download-File([string]$Url, [string]$OutFile) {
  $wc = New-Object Net.WebClient
  $wc.Headers.Add("User-Agent", "dawnfetch-installer")
  try {
    $wc.DownloadFile($Url, $OutFile)
  } finally {
    $wc.Dispose()
  }
}

function Resolve-LatestVersion {
  $apiUrl = "https://api.github.com/repos/almightynan/dawnfetch/releases/latest"
  $json = Download-Text $apiUrl
  $m = [regex]::Match($json, '"tag_name"\s*:\s*"v?([^"]+)"')
  if (-not $m.Success) {
    throw "could not parse latest version from github releases"
  }
  return $m.Groups[1].Value
}

function Resolve-Arch {
  $archText = (($env:PROCESSOR_ARCHITECTURE + " " + $env:PROCESSOR_ARCHITEW6432).ToLower())
  if ($archText -match "arm64") {
    return "arm64"
  }
  if ([Environment]::Is64BitOperatingSystem) {
    return "amd64"
  }
  return "386"
}

function Try-ParseVersionFromZip([string]$Path) {
  $name = [IO.Path]::GetFileName($Path)
  if (Is-Blank $name) {
    return ""
  }
  $m = [regex]::Match($name, '^dawnfetch_([^_]+)_windows_(amd64|386|arm64)\.zip$')
  if ($m.Success) {
    return $m.Groups[1].Value
  }
  return ""
}

function Expand-ZipCompat([string]$ZipPath, [string]$Destination) {
  if (Get-Command Expand-Archive -ErrorAction SilentlyContinue) {
    Expand-Archive -Path $ZipPath -DestinationPath $Destination -Force
    return
  }

  try {
    Add-Type -AssemblyName System.IO.Compression.FileSystem -ErrorAction Stop
    [IO.Compression.ZipFile]::ExtractToDirectory($ZipPath, $Destination)
    return
  } catch {
    # fallback below
  }

  $shell = New-Object -ComObject Shell.Application
  $zipNs = $shell.NameSpace($ZipPath)
  $destNs = $shell.NameSpace($Destination)
  if ($null -eq $zipNs -or $null -eq $destNs) {
    throw "failed to extract zip archive"
  }
  # 16 = yes to all
  $destNs.CopyHere($zipNs.Items(), 16)
}

function Path-Contains([string]$PathList, [string]$Needle) {
  $needleNorm = $Needle.Trim().TrimEnd('\').ToLower()
  if (Is-Blank $needleNorm) {
    return $false
  }
  foreach ($item in ($PathList -split ";")) {
    $partNorm = $item.Trim().TrimEnd('\').ToLower()
    if ($partNorm -eq $needleNorm) {
      return $true
    }
  }
  return $false
}

function Add-UserPath([string]$Dir) {
  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if ($null -eq $userPath) {
    $userPath = ""
  }
  $changed = $false
  if (-not (Path-Contains $userPath $Dir)) {
    if ($userPath.Length -gt 0 -and -not $userPath.EndsWith(";")) {
      $userPath += ";"
    }
    $userPath += $Dir
    [Environment]::SetEnvironmentVariable("Path", $userPath, "User")
    $changed = $true
  }

  if (-not (Path-Contains $env:Path $Dir)) {
    $env:Path = "$Dir;$env:Path"
  }
  return $changed
}

function Refresh-PathMessage {
  Write-Warn "if 'dawnfetch' is not recognized in this shell, run:"
  Write-Host '$env:Path = [Environment]::GetEnvironmentVariable("Path","User") + ";" + [Environment]::GetEnvironmentVariable("Path","Machine")' -ForegroundColor DarkGray
  Write-Warn "or restart powershell/cmd."
}

Enable-Tls12

if (-not (Is-Blank $ZipPath)) {
  $ZipPath = [IO.Path]::GetFullPath($ZipPath)
  if (-not (Test-Path -LiteralPath $ZipPath)) {
    throw "zip path does not exist: $ZipPath"
  }
  if (Is-Blank $Version) {
    $parsed = Try-ParseVersionFromZip $ZipPath
    if (-not (Is-Blank $parsed)) {
      $Version = $parsed
    } else {
      $Version = "local"
    }
  }
} else {
  if (Is-Blank $Version) {
    Write-Info "resolving latest release version..."
    $Version = Resolve-LatestVersion
  }
}
$Version = ($Version.Trim() -replace '^[vV]', '')
if (Is-Blank $Version) { $Version = "local" }

if (Is-Blank $InstallDir) {
  if (-not (Is-Blank $env:LOCALAPPDATA)) {
    $InstallDir = Join-Path $env:LOCALAPPDATA "dawnfetch"
  } elseif (-not (Is-Blank $env:USERPROFILE)) {
    $InstallDir = Join-Path $env:USERPROFILE "AppData\\Local\\dawnfetch"
  } else {
    throw "could not resolve install directory"
  }
}

$arch = Resolve-Arch
Write-Info "install version: $Version"
Write-Info "detected arch: $arch"
Write-Info "install dir: $InstallDir"

$tempRoot = Join-Path $env:TEMP ("dawnfetch-install-" + [Guid]::NewGuid().ToString("N"))
$downloadedZipPath = ""
$extractDir = Join-Path $tempRoot "extract"

New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

if (-not (Is-Blank $ZipPath)) {
  Write-Info "using local zip: $ZipPath"
  Expand-ZipCompat -ZipPath $ZipPath -Destination $extractDir
} else {
  $assetName = "dawnfetch_${Version}_windows_${arch}.zip"
  $downloadUrl = "https://github.com/almightynan/dawnfetch/releases/download/v${Version}/${assetName}"
  $downloadedZipPath = Join-Path $tempRoot $assetName
  Write-Info "download: $assetName"
  try {
    Download-File -Url $downloadUrl -OutFile $downloadedZipPath
  } catch {
    Write-Warn "download failed: $($_.Exception.Message)"
    Write-Warn "if you're on older windows, powershell may not support github tls requirements."
    Write-Warn "fallback: download the release zip in browser, then run:"
    Write-Warn "  powershell -File .\\install.ps1 -ZipPath C:\\path\\to\\dawnfetch_windows.zip"
    throw
  }
  Expand-ZipCompat -ZipPath $downloadedZipPath -Destination $extractDir
}

$srcDir = $extractDir
$children = @(Get-ChildItem -LiteralPath $extractDir -Force)
if ($children.Count -eq 1 -and $children[0].PSIsContainer) {
  $srcDir = $children[0].FullName
}

if ($Force) {
  Get-ChildItem -LiteralPath $InstallDir -Force -ErrorAction SilentlyContinue |
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
}

Copy-Item -Path (Join-Path $srcDir "*") -Destination $InstallDir -Recurse -Force

$exePath = Join-Path $InstallDir "dawnfetch.exe"
if (-not (Test-Path -LiteralPath $exePath)) {
  $exeCandidate = Get-ChildItem -Path $InstallDir -Filter "dawnfetch*.exe" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($null -ne $exeCandidate) {
    Copy-Item -Path $exeCandidate.FullName -Destination $exePath -Force
  }
}
if (-not (Test-Path -LiteralPath $exePath)) {
  throw "installation failed: dawnfetch.exe was not found after extraction"
}

$pathUpdated = Add-UserPath -Dir $InstallDir

Write-Good "dawnfetch installed successfully."
Write-Good "run now:"
Write-Host "  dawnfetch" -ForegroundColor DarkGray
if ($pathUpdated) {
  Write-Info "path updated for current user."
}

if (-not (Get-Command dawnfetch -ErrorAction SilentlyContinue)) {
  Refresh-PathMessage
} else {
  Write-Info "if you opened a new shell and command is still missing, restart powershell/cmd."
}

try {
  Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
} catch {
  # no-op
}
