# Changelog

All notable changes to this project are recorded in this file.

## [1.0.23] - 2026-03-30

### Changed
- Replaced the placeholder Wails app icon with a project icon asset for Windows desktop builds.
- Switched the Windows resource build to use a generated project icon PNG instead of the old placeholder `.ico`.
- Added an upstream sync script for `ffmpeg-over-ip` Windows release zips and stopped treating those large zip artifacts as files that belong in Git.

## [1.0.22] - 2026-03-30

### Added
- Added Windows resource generation with the project icon so desktop builds embed the app icon into the executable.
- Added a repository `.gitignore` to keep local caches, runtime state, generated resources, and Windows build artifacts out of Git.

## [1.0.21] - 2026-03-30

### Fixed
- Fixed the multicam completion report using the streamed `totalTime` field so timeline duration no longer shows `undefined сек`.

## [1.0.20] - 2026-03-30

### Fixed
- Moved desktop API key persistence from iframe localStorage to backend settings storage, so AssemblyAI and AI keys now survive full app restarts even when the embedded UI runs on a different localhost port each session.

## [1.0.19] - 2026-03-30

### Fixed
- Restored multicam completion fallback after switching output selection to folder mode by resolving the final output file path in the UI the same way as the backend.
- Added a short post-stream flush delay so the final multicam completion event is less likely to be lost before the HTTP stream closes.

## [1.0.18] - 2026-03-30

### Changed
- Reworked the multicam UI to match the single-cam flow more closely: output picking now uses folder selection, render quality/speed use the same style of presets, and AI settings moved to the Render Backend tab.
- Removed the extra multicam explanatory texts and hid the internal shot-window/min-shot tuning from the UI, keeping stable defaults in code.

## [1.0.17] - 2026-03-30

### Fixed
- Added a final output existence check so the UI no longer reports a fatal `network error` when the streamed log drops after a successful multicam render.
- Empty derived `aligned` folders are now removed after multicam render completion.

## [1.0.16] - 2026-03-30

### Fixed
- Updated the AssemblyAI pre-recorded transcription request to use the required `speech_models` array with language detection, replacing the deprecated `speech_model` field.

## [1.0.15] - 2026-03-30

### Added
- Persisted AssemblyAI and LLM API keys locally in the desktop UI so they survive app restarts.

### Fixed
- Improved AssemblyAI upload and transcript-start diagnostics by surfacing the real API response text instead of a generic failure message.

## [1.0.14] - 2026-03-30

### Changed
- Reworked AI multicam editing to a steadier speaker-driven plan: the dominant speaker stays on the main camera, and secondary speakers switch only on longer utterances.

### Fixed
- Fixed AssemblyAI preparation on Windows by staging master audio through Windows-safe paths before ffmpeg converts it to WAV.
- Improved the `prepare wav` error message so ffmpeg stderr is shown instead of a raw exit status code.

## [1.0.13] - 2026-03-30

### Changed
- Made classic multicam switching less jittery by preferring the current camera unless another view has a clearly stronger score.
- Stopped creating the derived `aligned` directory during command export alone.

### Fixed
- Respected video rotation metadata when probing source streams so vertical phone footage keeps a vertical render canvas.
- Avoided false final `network error` reports in the UI when the render stream has already delivered a completed event.

## [1.0.12] - 2026-03-30

### Fixed
- Reworked the desktop shell to launch the embedded HTTP backend on a pre-bound free localhost port and inject that exact address into the window, removing the fixed-port startup hang.

## [1.0.11] - 2026-03-30

### Fixed
- Made the desktop shell wait for the embedded local backend before loading the UI, avoiding startup `network error` screens when the HTTP service is not ready yet.

## [1.0.10] - 2026-03-30

### Changed
- Removed the separate multicam aligned-output field from the UI and now derive the aligned clip folder automatically from the final output path.
- Added multicam preview length modes for render checks: full file, 2 minutes, or 5 minutes.

## [1.0.9] - 2026-03-30

### Changed
- Simplified the multicam screen by keeping the main workflow visible and moving render tuning plus AI/Shorts options into collapsible sections.

### Fixed
- Fixed multicam preprocessing on Windows so envelope extraction and video probing also use staged Windows-safe paths for files with Cyrillic names.

## [1.0.8] - 2026-03-30

### Changed
- Returned action buttons to a neutral default look, with role colors shown only on interaction.
- Added the app version to the window title, in-app title, and versioned desktop output filename.

## [1.0.7] - 2026-03-30

### Changed
- Updated primary single-cam action buttons to fixed role colors: green for Analyze, blue for Render, red for Cancel.
- Renamed the single-cam action labels to shorter button titles.

## [1.0.6] - 2026-03-30

### Changed
- Switched the single-cam folder picker to an explorer-style folder selection flow instead of the old tree dialog.
- Restored the previous single-cam output field copy and browse button styling.

## [1.0.5] - 2026-03-30

### Changed
- Simplified the single-cam output folder UI so it behaves like a plain folder picker without showing a file name in the control copy.

## [1.0.4] - 2026-03-30

### Changed
- Switched single-cam output selection back to folder picking with automatic output file naming.
- Single-cam default output names now follow the pattern `source_creationtime_sync.ext`.

### Fixed
- `.autosync-temp` is now cleaned up after staged render work completes.

## [1.0.3] - 2026-03-30

### Changed
- Avoided full render-time staging copies when Windows short paths are available for source files and output directories.

### Fixed
- Reduced disk-heavy pre-render copying that made renders look frozen before ffmpeg progress appeared.

## [1.0.2] - 2026-03-30

### Fixed
- Routed ffmpeg progress to the same output stream that the desktop UI reads, restoring live render updates.
- Continued Windows-safe render path handling for single-cam output and source staging fixes.

## [1.0.1] - 2026-03-30

### Changed
- Switched desktop builds to Windows GUI production mode so Wails apps no longer open a console window.
- Added a local build workflow in `build-local.ps1` with isolated cache and profile directories inside the workspace.
- Aligned multicam final output naming and save-path behavior with single-cam sync.
- Optimized Windows input handling by preferring short paths before falling back to staging copies.
- Fixed single-cam and multicam render paths so Windows-safe staging applies during render, not only during analysis.
- Restored ffmpeg progress streaming through the desktop UI.
- Replaced the single-cam output picker with a save-file dialog instead of a folder-only picker.

### Added
- Added explicit project version tracking via `VERSION`.
- Added Wails desktop config files for desktop targets.

### Fixed
- Fixed Wails desktop binaries being built without the required production tags.
- Fixed render failures on Windows when source or output paths contain non-ASCII characters.
