# builds windows executables into ./dist for amd64, 386, and arm64.

[CmdletBinding()]
param(
  [string]$GoBin = "go",
  [string]$Package = ".",
  [string]$Ldflags = ""
)

$ErrorActionPreference = "Stop"

$rootDir = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$distDir = Join-Path $rootDir "dist"

if (-not (Test-Path $distDir)) {
  New-Item -ItemType Directory -Path $distDir | Out-Null
}

$targets = @("amd64", "386", "arm64")
foreach ($arch in $targets) {
  $outPath = Join-Path $distDir ("dawnfetch-windows-{0}.exe" -f $arch)
  Write-Host ("building {0}" -f $outPath)

  $env:CGO_ENABLED = "0"
  $env:GOOS = "windows"
  $env:GOARCH = $arch

  $args = @("build", "-trimpath")
  if ($Ldflags -ne "") {
    $args += @("-ldflags", $Ldflags)
  }
  $args += @("-o", $outPath, $Package)

  & $GoBin @args
  if ($LASTEXITCODE -ne 0) {
    throw ("go build failed for GOARCH={0}" -f $arch)
  }
}

Write-Host ("done. windows binaries are in {0}" -f $distDir)
