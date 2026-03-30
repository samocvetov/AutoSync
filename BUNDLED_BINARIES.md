# Bundled Binaries

This project is moving toward self-contained Windows builds where the shipped `.exe` files contain the rendering tools they need.

## Single Source Of Truth

The authoritative version manifest lives in:

- `internal/bundles/manifest.json`

Both desktop programs read from that manifest so the UI and logs show the same bundled versions.

## Planned Windows Bundle Layout

- `third_party/windows/ffmpeg/current/ffmpeg.exe`
- `third_party/windows/ffmpeg/current/ffprobe.exe`
- `third_party/windows/ffmpeg-over-ip/current/ffmpeg-over-ip-client.exe`
- `third_party/windows/ffmpeg-over-ip/current/ffmpeg-over-ip-server.exe`
- `third_party/windows/ffmpeg-over-ip/current/ffmpeg.exe`
- `third_party/windows/ffmpeg-over-ip/current/ffprobe.exe`

The `current` folders are stable paths on purpose. Updating a dependency should not require searching the codebase for file paths.

## Update Flow

1. Download the new Windows release artifacts for `ffmpeg` and `ffmpeg-over-ip`.
2. Replace the binaries inside the matching `third_party/windows/.../current/` folder.
3. Update `internal/bundles/manifest.json` with the new versions and notes.
4. Rebuild:
   - `AutoSyncStudio.exe`
   - `AutoSyncRenderServer.exe`
5. Run smoke tests for:
   - local CPU render
   - local GPU render
   - remote render through `ffmpeg-over-ip`
