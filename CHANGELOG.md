# Changelog

All notable changes to this project are recorded in this file.

## [1.0.86] - 2026-04-02

### Changed
- Widened the safe subtitle area and reduced vertical `Montserrat` size slightly so centered captions stay on two lines instead of wrapping into three or four lines.

## [1.0.85] - 2026-04-02

### Changed
- Switched vertical `Montserrat` subtitles from whole-utterance captions to sentence-by-sentence captions, so only one sentence is shown at a time while still holding it for its spoken interval.

## [1.0.84] - 2026-04-02

### Changed
- Switched vertical `Montserrat` captions to sentence-level subtitle chunks so a full spoken phrase stays on screen for the duration of the utterance instead of being broken into shorter word groups.

## [1.0.83] - 2026-04-02

### Changed
- Forced vertical `Montserrat` captions into a balanced two-line layout and added a light ASS blur pass for smoother text edges.

## [1.0.82] - 2026-04-02

### Changed
- Raised the vertical `Montserrat` subtitle block higher, widened its safe area, and made line wrapping more aggressive so captions break into two centered lines more consistently.

## [1.0.81] - 2026-04-02

### Changed
- Raised, widened, and enlarged the vertical `Montserrat` subtitle layout to better match the reference lower-third style.

## [1.0.80] - 2026-04-02

### Fixed
- Disabled ASS boxed background mode when subtitle background opacity is set to `0%`, so captions render with outline only and no visible subtitle box.

## [1.0.79] - 2026-04-02

### Changed
- Added a dedicated dark text outline to subtitle rendering so `Montserrat` captions have a clearer stroke on both shorts and full interview exports.

## [1.0.78] - 2026-04-02

### Added
- Embedded the `Montserrat` subtitle font into the app and wired ffmpeg/libass to load it from the app runtime instead of relying on system fonts.

### Changed
- Tuned subtitle layout for `Montserrat` to render as centered two-line captions with a calmer lower-third placement for both shorts and full interview exports.

## [1.0.77] - 2026-04-02

### Added
- Added `Montserrat` to subtitle font choices in Shorts / Reels.

### Changed
- Replaced subtitle background opacity steps with `0% / 10% / 25% / 50% / 75% / 100%`.
- Made `0%` subtitle background opacity mean no subtitle background for both shorts renders and full interview subtitle exports.

## [1.0.76] - 2026-04-02

### Fixed
- Normalized ffmpeg progress reporting in Shorts / Reels by parsing progress from stderr fallback lines and filtering noisy muxing summary lines.

## [1.0.75] - 2026-04-02

### Fixed
- Corrected ASS subtitle box styling so background transparency is actually applied to both shorts captions and full interview subtitles.

## [1.0.74] - 2026-04-02

### Fixed
- Reworked shorts caption rendering to use subtitle files instead of large inline drawtext graphs, avoiding `Cannot allocate memory` on long captioned clip renders.

### Added
- Added subtitle background opacity control and applied subtitle style settings to both shorts exports and full interview subtitle exports.

## [1.0.73] - 2026-04-02

### Added
- Added subtitle background opacity control and wired subtitle style controls into both shorts renders and full interview subtitle exports.

### Changed
- Grouped subtitle font, background color, and background opacity onto a single row in Shorts / Reels.

## [1.0.72] - 2026-04-02

### Added
- Added subtitle customization controls for full interview exports: selectable font presets and subtitle background color.

### Fixed
- Tightened full-interview subtitle wrapping and ASS styling so captions stay inside the video frame more reliably.

## [1.0.71] - 2026-04-02

### Changed
- Replaced the full-interview subtitle export path with a dedicated subtitle-file pipeline built around external `.ass`/`.srt` assets instead of giant inline `drawtext` graphs.

### Added
- Full-interview subtitle export now saves reusable `MP4 + SRT + TXT` files for YouTube uploads, translation, and repurposing transcript text.

## [1.0.70] - 2026-04-02

### Fixed
- Ensured the `.autosync-temp` staging folder is created before writing ffmpeg filter script files for long full-interview subtitle renders.

## [1.0.69] - 2026-04-02

### Fixed
- Moved oversized full-interview subtitle filter graphs out of the Windows command line into temporary ffmpeg filter script files.
- This avoids `The filename or extension is too long` when burning subtitles into long interviews.

## [1.0.68] - 2026-04-02

### Fixed
- Reworked AssemblyAI upload to stream the prepared audio file instead of reading the entire WAV into memory first.
- Added retry handling for transient AssemblyAI upload failures such as `EOF`, timeouts, and broken connections on long interviews.

## [1.0.67] - 2026-04-02

### Fixed
- Hardened full-interview subtitle export duration detection by falling back to sync metrics and video stream metadata when `ffprobe format=duration` is missing.

## [1.0.66] - 2026-04-02

### Changed
- Reworked the `Титры` action in Shorts / Reels into a standalone full-interview subtitle export that no longer depends on `Build plan`.

### Fixed
- Added a dedicated full-captions backend flow that uploads only audio to AssemblyAI, builds a word-level transcript, and burns subtitles into the entire source interview in its original aspect ratio.

## [1.0.65] - 2026-04-02

### Fixed
- Explicitly re-enabled Shorts / Reels render actions after `Build plan` so the `Титры` button no longer stays stuck in an inert state.
- Added immediate log feedback on `Титры` click, making it clear whether the full subtitled export is starting or missing prerequisites.

## [1.0.64] - 2026-04-02

### Fixed
- Made the new `Full interview with captions` action clickable before Build plan completes, so it now shows a clear validation message instead of silently looking unresponsive.
- Added full-interview subtitle export through the existing Shorts / Reels pipeline, preserving the source aspect ratio and saving a separate full export plan file.

## [1.0.62] - 2026-04-02

### Changed
- Replaced the Shorts / Reels progress bar with a proper cancel button and live status messages in the output log.
- Added cancellable AssemblyAI planning requests and log messages with explicit estimated processing percentages where the API does not provide a real progress field.

## [1.0.63] - 2026-04-02

### Fixed
- Switched Shorts / Reels burned-in captions to AssemblyAI word-level timings to stop subtitle drift by the end of a clip.
- Removed multi-line drawtext escaping from shorts captions so stray literal `n` characters no longer appear between words.

## [1.0.61] - 2026-04-02

### Fixed
- Reworked burned-in Shorts / Reels captions into shorter timed subtitle chunks so vertical exports no longer show one oversized static line.
- Reduced the clip review card metadata to a compact duration-only line.

## [1.0.60] - 2026-04-02

### Added
- Added a live progress bar for `Build plan` in Shorts / Reels, driven by a new streaming shorts-plan endpoint.

## [1.0.59] - 2026-04-01

### Fixed
- Simplified the Shorts / Reels clip list so it no longer shows bulky preview command details.
- Added scrollable clip review area in Shorts / Reels.
- Forced burned-in captions to use an explicit Windows font with Cyrillic support to avoid square glyphs.

## [1.0.58] - 2026-04-01

### Fixed
- Cleaned the Shorts / Reels tab by removing redundant helper text and the long formats summary line.
- Added a non-stream fallback for Shorts / Reels rendering so `Render selected` still works when streaming fetch is unreliable in the desktop runtime.

## [1.0.57] - 2026-04-01

### Changed
- Split Shorts / Reels into a dedicated top-level tab and kept API keys in Render Backend so Single-Cam stays compact.

## [1.0.56] - 2026-04-01

### Fixed
- Forced the Shorts / Reels block to stay visible in Single-Cam by overriding the legacy hidden state from the old sync advanced-toggle CSS.

## [1.0.55] - 2026-04-01

### Fixed
- Restored the Shorts / Reels AI block visibility in Single-Cam after the move from Render Backend.

## [1.0.54] - 2026-04-01

### Changed
- Moved the Shorts / Reels workflow into Single-Cam for finished interview videos, with plan review, selected-clip rendering, social presets, captions mode, and saved `plan.json`.

## [1.0.53] - 2026-03-31

### Fixed
- Restored the stable multicam render flow by reusing measured offsets from the latest Analyze step during Render, preventing Analyze/Render drift on the same material.
- Added explicit preview-render confirmation and clearer preview status so 2-minute and 5-minute test renders are no longer easy to confuse with full final renders.
- Reduced remote backend secret exposure by switching runtime remote ffmpeg-over-ip config files to temporary per-run files for actual render execution.

## [1.0.52] - 2026-03-31

### Fixed
- Fixed the remaining Russian multicam label regression where `label_primary_short` still rendered as `Primary camera` instead of `Главная камера`.

## [1.0.51] - 2026-03-31

### Fixed
- Reissued the UI build with the multicam camera selector label locked to `Главная камера` in the rendered desktop bundle.

## [1.0.50] - 2026-03-31

### Changed
- Replaced the remaining English `Primary camera` field label with `Главная камера` so the label stays identical in both language modes.

## [1.0.49] - 2026-03-31

### Changed
- Renamed the remaining English `Main camera` label to `Primary camera` for consistency with the rest of the UI.

## [1.0.48] - 2026-03-31

### Changed
- Renamed `Куда сохранить` to `Сохранить` across the UI labels.
- Moved the multicam final output field onto its own row.
- Replaced the `Main Cam` numeric field with a separate `Главная камера` dropdown row.

## [1.0.47] - 2026-03-31

### Fixed
- Reverted the overly aggressive 1.0.46 spacing rules that distorted the `Single-Cam Sync` and `Multicam` forms.
- Shortened the multicam row label to `Режим` to keep the field height aligned with the rest of the form.

## [1.0.46] - 2026-03-31

### Changed
- Tightened vertical spacing around `Куда сохранить`, `Проверка`, `Качество` and the following rows in both `Single-Cam Sync` and `Multicam` by removing extra bottom margin inside option grids and using explicit visible-row spacing rules.
- Translated the multicam `Edit Mode` label to Russian: `Режим монтажа`.

## [1.0.45] - 2026-03-31

### Changed
- Unified row spacing in both `Single-Cam Sync` and `Multicam` with one direct per-row rhythm rule, so standalone rows, two-column option grids, and action buttons now keep the same vertical distance.

## [1.0.44] - 2026-03-31

### Changed
- Reworked the form layout rhythm in both `Single-Cam Sync` and `Multicam`: row spacing is now driven by one stronger 12px vertical system instead of a mix of small per-control margins.

## [1.0.43] - 2026-03-31

### Changed
- Unified the vertical spacing between form rows in both `Single-Cam Sync` and `Multicam`, using one consistent spacing step instead of mixed gaps.

## [1.0.42] - 2026-03-31

### Changed
- Renamed the multicam primary action buttons to match the simpler single-cam wording: `Анализ`, `Рендер`, `Отмена`.

## [1.0.41] - 2026-03-31

### Changed
- Simplified the multicam action row: `Анализ` now shows both camera offsets and export commands in one pass, so the separate `Экспорт команд` button was removed from the main workflow.

## [1.0.40] - 2026-03-31

### Changed
- Removed the stale hardcoded version line from the repository README so the project description stays evergreen.
- Moved the main tab navigation into the same top bar row as the app title and `RU/EN` language switch, matching the intended compact desktop layout.

## [1.0.39] - 2026-03-31

### Fixed
- Fixed the multicam render coordinate bug introduced in 1.0.38: after trimming leading empty timeline time, shot segments now stay in original master-audio coordinates instead of being shifted a second time, so the first real camera no longer disappears and later shots no longer freeze from negative source trims.

## [1.0.38] - 2026-03-31

### Fixed
- Normalized final multicam shot segments against real camera availability before rendering: delayed cameras are no longer allowed to start a shot before they actually have footage, and tiny timing gaps are now reassigned or extended instead of being rendered as black `tpad` inserts between cuts.
- Trimmed the leading empty portion of the final multicam output when the first usable camera starts later than master audio, so the exported video no longer begins with artificial empty pre-roll.

## [1.0.37] - 2026-03-31

### Fixed
- Updated exported multicam aligned-camera commands to match the actual aligned-render implementation: delayed cameras now use black `tpad` lead-in (`start_mode=add:color=black`) instead of cloning the first frame.

## [1.0.36] - 2026-03-31

### Fixed
- Reverted the incorrect 1.0.35 multicam baseline shift. Master audio now remains anchored at timeline zero again, matching the Premiere layout where all cameras start later than the external audio.

## [1.0.35] - 2026-03-31

### Fixed
- Changed final multicam rendering to trim the master audio by the earliest positive camera offset and shift all camera delays relative to that baseline, matching the working single-cam sync model instead of keeping full audio pre-roll and padding video from absolute master time zero.

## [1.0.34] - 2026-03-31

### Fixed
- Restored the Russian UTF-8 UI strings in the desktop interface after the 1.0.33 encoding regression in `main.js`.
- Reapplied the clean multicam fallback completion report without the misleading log-break warning, this time without damaging file encoding.
## [1.0.33] - 2026-03-31

### Fixed
- Stopped the fast multicam planner from blindly falling back to the primary camera for windows that no camera fully covers, which had still been producing bad cuts and apparent desync on large-offset timelines.
- Restored per-shot black lead-in inside the fast multicam render path for ranges that begin before a delayed camera actually starts, preserving timeline timing without reviving the slow aligned pre-render stage.
- Removed the leftover multicam fallback message claiming the log stream had broken after completion; successful fallback completion now shows a clean finish report instead of an error-like warning.

## [1.0.32] - 2026-03-31

### Fixed
- Removed the accidental per-camera aligned pre-render step from final multicam rendering, restoring the fast direct-edit pipeline instead of creating mandatory temporary mezzanine files before every render.
- Tightened camera eligibility in multicam shot selection so the fast path only uses cameras that fully cover the requested shot window, avoiding the earlier fake lead-in behavior without bringing back slow aligned staging.

## [1.0.31] - 2026-03-30

### Fixed
- Fixed multicam aligned-mezzanine generation after the 1.0.29 render rewrite by staging intermediate aligned outputs through the same Windows-safe output path handling used elsewhere, avoiding `Error opening output files` on Cyrillic paths.

## [1.0.30] - 2026-03-30

### Fixed
- Restored the broken Russian UI strings in the desktop multicam interface after the previous encoding regression in `main.js`.

## [1.0.29] - 2026-03-30

### Fixed
- Switched final multicam rendering to pre-render per-camera aligned mezzanine files and cut the final timeline from those aligned sources instead of trimming raw delayed camera inputs directly.
- Changed internal multicam aligned lead-in from frozen first-frame padding to black padding, matching the practical “camera is not on the timeline yet” behavior more closely for delayed sources.
- Removed the alarming multicam fallback text about the log stream breaking after completion so successful renders now end with a clean completion report.

## [1.0.28] - 2026-03-30

### Fixed
- Prevented `Smart AI` multicam from choosing cameras for timeline ranges they do not actually cover yet, which had been causing frozen lead-in frames and severe apparent desync on sources with large positive offsets.
- Added coverage-aware camera selection for utterance segments, silent gaps, and long cutaway insertion so delayed cameras only enter the edit when they have enough real footage for that portion of the timeline.

## [1.0.27] - 2026-03-30

### Fixed
- Reworked `Smart AI` multicam shot planning so it no longer behaves like a hardcoded two-camera interview mode when three or more cameras are present.
- Added camera-aware alternate shot selection for non-primary speakers and stable cutaway insertion inside long primary-camera segments, allowing extra cameras to appear without chaotic switching.

## [1.0.26] - 2026-03-30

### Changed
- Cleaned the project root by removing the legacy `OLD` tree from version control and tightening `.gitignore` for cache folders, runtime state, Windows thumbnails, and obsolete local build artifacts.

## [1.0.25] - 2026-03-30

### Added
- Added a Russian `README.md` with an overview of the desktop apps, key features, build flow, project structure, and bundled Windows dependencies.

## [1.0.24] - 2026-03-30

### Changed
- Installed the new user-provided app icon into the Windows build asset.
- Reworked the Windows resource generation so Wails desktop builds place the icon under the resource ID expected by the native window title bar, with a dedicated 16x16 variant for the small system icon.

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
