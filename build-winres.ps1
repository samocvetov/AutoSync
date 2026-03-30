$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$version = (Get-Content (Join-Path $root "VERSION") -Raw).Trim()
$goWinres = "C:\Users\sysop\go\bin\go-winres.exe"
$icon = Join-Path $root "build\windows\icon.png"

if (-not (Test-Path $goWinres)) {
  throw "go-winres not found at $goWinres"
}

if (-not (Test-Path $icon)) {
  throw "Icon not found at $icon"
}

function Write-WinResources {
  param(
    [string]$TargetDir,
    [string]$Description
  )

  Push-Location $TargetDir
  try {
    & $goWinres simply `
      --icon $icon `
      --manifest gui `
      --product-version $version `
      --file-version $version `
      --product-name "AutoSync Studio" `
      --file-description $Description | Out-Host
  } finally {
    Pop-Location
  }
}

Write-WinResources -TargetDir (Join-Path $root "cmd\studio-desktop") -Description "AutoSync Studio Desktop"
Write-WinResources -TargetDir (Join-Path $root "cmd\render-server-desktop") -Description "AutoSync Studio Render Server"
