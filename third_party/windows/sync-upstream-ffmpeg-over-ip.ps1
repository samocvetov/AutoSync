$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$releaseBase = "https://github.com/steelbrain/ffmpeg-over-ip/releases/download/v5.0.0"
$clientZip = Join-Path $root "windows-amd64-ffmpeg-over-ip-client.zip"
$serverZip = Join-Path $root "windows-amd64-ffmpeg-over-ip-server.zip"
$clientUrl = "$releaseBase/windows-amd64-ffmpeg-over-ip-client.zip"
$serverUrl = "$releaseBase/windows-amd64-ffmpeg-over-ip-server.zip"
$clientExtract = Join-Path $root "ffmpeg-over-ip\current"
$serverExtract = Join-Path $root "ffmpeg-over-ip\current"

Invoke-WebRequest -Uri $clientUrl -OutFile $clientZip
Invoke-WebRequest -Uri $serverUrl -OutFile $serverZip

New-Item -ItemType Directory -Force -Path $clientExtract | Out-Null
New-Item -ItemType Directory -Force -Path $serverExtract | Out-Null

Expand-Archive -LiteralPath $clientZip -DestinationPath $clientExtract -Force
Expand-Archive -LiteralPath $serverZip -DestinationPath $serverExtract -Force

Write-Host "Downloaded and unpacked upstream ffmpeg-over-ip release artifacts."
