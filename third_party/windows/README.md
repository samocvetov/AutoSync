# Windows Third-Party Binaries

The desktop builds embed the unpacked Windows binaries from this directory.

Upstream source:
- `https://github.com/steelbrain/ffmpeg-over-ip/releases/tag/v5.0.0`
- `https://github.com/steelbrain/ffmpeg-over-ip/releases/download/v5.0.0/windows-amd64-ffmpeg-over-ip-client.zip`
- `https://github.com/steelbrain/ffmpeg-over-ip/releases/download/v5.0.0/windows-amd64-ffmpeg-over-ip-server.zip`

To refresh the archived upstream downloads, run:
- `powershell -ExecutionPolicy Bypass -File .\third_party\windows\sync-upstream-ffmpeg-over-ip.ps1`

The large release zip files are no longer meant to live in Git. We download them from the upstream GitHub release when needed, then keep only the unpacked `current\*.exe` files for embedding.

See [BUNDLED_BINARIES.md](C:\Admin\AutoSyncStudio\BUNDLED_BINARIES.md) and [manifest.json](C:\Admin\AutoSyncStudio\internal\bundles\manifest.json).
