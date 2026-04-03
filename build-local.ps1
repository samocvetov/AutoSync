$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$cacheRoot = Join-Path $root ".cache"
$version = (Get-Content (Join-Path $root "VERSION") -Raw).Trim()

if (-not $version) {
  throw "VERSION file is empty"
}

$studioCliExe = "autosync-studio-$version.exe"
$renderCliExe = "autosync-render-server-$version.exe"
$studioDesktopExe = "AutoSyncStudioDesktop-$version.exe"
$renderDesktopExe = "AutoSyncRenderServerDesktop-$version.exe"

$dirs = @(
  $cacheRoot,
  (Join-Path $cacheRoot "go-build"),
  (Join-Path $cacheRoot "go-tmp"),
  (Join-Path $cacheRoot "tmp"),
  (Join-Path $cacheRoot "tmp2"),
  (Join-Path $cacheRoot "appdata"),
  (Join-Path $cacheRoot "localappdata"),
  (Join-Path $cacheRoot "home")
)

foreach ($dir in $dirs) {
  New-Item -ItemType Directory -Force -Path $dir | Out-Null
}

$env:GOCACHE = Join-Path $cacheRoot "go-build"
$env:GOTMPDIR = Join-Path $cacheRoot "go-tmp"
$env:TEMP = Join-Path $cacheRoot "tmp"
$env:TMP = Join-Path $cacheRoot "tmp2"
$env:GOTELEMETRY = "off"
$env:APPDATA = Join-Path $cacheRoot "appdata"
$env:LOCALAPPDATA = Join-Path $cacheRoot "localappdata"
$env:HOME = Join-Path $cacheRoot "home"
$env:USERPROFILE = $env:HOME
$env:GOPROXY = "off"

& (Join-Path $root "build-winres.ps1")

function Build-WailsDesktop {
  param(
    [string]$ProjectDir,
    [string]$ExpectedExeName
  )

  Write-Host "Building $ExpectedExeName as Windows GUI app from $ProjectDir..."
  go build -work -buildvcs=false -tags production -ldflags "-H windowsgui -w -s" -o $ExpectedExeName $ProjectDir
}

Write-Host "Building $studioCliExe from root package..."
go build -work -buildvcs=false -o $studioCliExe .

Write-Host "Building $renderCliExe from render-server package..."
go build -work -buildvcs=false -o $renderCliExe ./cmd/render-server

Build-WailsDesktop -ProjectDir "./cmd/studio-desktop" -ExpectedExeName $studioDesktopExe
Build-WailsDesktop -ProjectDir "./cmd/render-server-desktop" -ExpectedExeName $renderDesktopExe

Write-Host ""
Write-Host "Build artifacts:"
Get-Item $studioDesktopExe, $studioCliExe, $renderDesktopExe, $renderCliExe |
  Select-Object Name, Length, LastWriteTime
