const STORAGE_LANGUAGE_KEY = "autosyncstudio.language";
const STORAGE_BACKEND_PREFS_KEY = "autosyncstudio.backendPrefs";
const STORAGE_SHORTS_PREFS_KEY = "autosyncstudio.shortsPrefs";
const DEFAULT_LANGUAGE = "ru";

const I18N = {
  ru: {
    document_title: "AutoSync Studio 1.0.34",
    app_title: "AutoSync Studio 1.0.34",
    hero_subtitle: "Новая версия проекта с фокусом на точный sync, а не на хрупкую магию.",
    tab_sync: "Single-Cam Sync",
    tab_multicam: "Multicam",
    tab_backend: "Render Backend",
    studio_single_subtitle: "Для сценария, где есть видеозапись с камеры и отдельный мастер-аудиофайл.",
    multicam_subtitle: "Сначала надежно измеряем смещения всех камер относительно мастер-аудио, а уже потом строим точный рендер без скрытой магии.",
    backend_panel_subtitle: "Единый backend исполнения для single-cam и multicam: локальный CPU, локальный GPU или удаленный ffmpeg-over-ip.",
    label_video_path: "Путь к видео",
    label_video_path_short: "Видео",
    label_audio_path: "Путь к мастер-аудио",
    label_audio_path_short: "Аудио",
    label_analyze_seconds: "Сколько секунд анализировать",
    label_analyze_short: "Сек. анализа",
    label_max_lag: "Максимальное окно поиска смещения",
    label_max_lag_short: "Макс. сдвиг",
    label_output_path: "Куда сохранить рендер",
    label_output_short: "Сохранить",
    label_crf: "CRF",
    label_crf_short: "Качество",
    label_preset: "Пресет кодирования",
    label_preset_short: "Скорость",
    backend_title: "Исполнитель рендера",
    label_execution_mode: "Где выполнять рендер",
    label_execution_short: "Mode",
    label_remote_client_path: "Путь к клиенту ffmpeg-over-ip",
    label_client_short: "Client",
    label_remote_address: "Адрес сервера",
    label_remote_address_short: "Address",
    label_remote_secret: "Общий секрет",
    label_remote_secret_short: "Secret",
    btn_analyze_sync: "Анализ",
    btn_render_sync: "Рендер",
    btn_cancel: "Отмена",
    sync_output_idle: "Здесь появятся результаты анализа смещения и точного рендера.",
    sync_note: "Если смещение положительное, раньше стартует мастер-аудио. Если отрицательное, раньше стартует видеозапись камеры.",
    sync_terms_note: "`Сек. анализа` — сколько первых секунд сравнивать. `Макс. сдвиг` — предел поиска рассинхрона. `Качество (CRF)` — баланс веса и качества: меньше число, выше качество. `Скорость` — насколько быстро кодировать видео.",
    label_master_audio: "Путь к мастер-аудио",
    label_master_audio_short: "Мастер",
    label_camera_paths: "Камеры, по одной на строку",
    label_camera_paths_short: "Камеры",
    label_aligned_dir: "Куда сохранить aligned-клипы",
    label_aligned_dir_short: "Сохранить",
    label_aligned_crf: "CRF для выровненных клипов",
    label_aligned_crf_short: "CRF",
    label_multicam_render_output: "Куда сохранить финальный multicam-рендер",
    label_final_short: "Сохранить",
    label_primary_camera: "Основная камера",
    label_primary_short: "Главная камера",
    label_shot_window: "Окно анализа плана, сек",
    label_shot_window_short: "Window",
    label_min_shot: "Минимальная длина плана, сек",
    label_min_shot_short: "Min Shot",
    multicam_backend_note: "Экспорт и финальный multicam-рендер используют тот же backend, что и single-cam режим.",
    btn_analyze_multicam: "Анализ",
    btn_export_commands: "Экспорт команд",
    btn_render_final: "Рендер",
    status_render_cancelled: "Рендер остановлен пользователем.",
    status_progress: "Прогресс",
    multicam_output_idle: "Здесь появятся результаты измерения смещений, экспортные команды и предпросмотр плана склеек.",
    multicam_note: "Это новая архитектура проекта: сначала надежная диагностика, затем точный рендер. Автомонтаж наращивается уже поверх корректной временной модели.",
    backend_output_idle: "Здесь отображаются статус встроенных компонентов и выбранная конфигурация backend.",
    browse_btn: "Обзор",
    placeholder_video_path: "C:\\Video\\camera.mp4",
    placeholder_audio_path: "C:\\Audio\\master.wav",
    placeholder_output_path: "C:\\Video\\out или C:\\Video\\result.mp4",
    placeholder_remote_client_path: "ffmpeg-over-ip-client",
    placeholder_remote_address: "127.0.0.1:5050",
    placeholder_remote_secret: "shared-secret",
    placeholder_camera_paths: "/path/to/cam1.mp4\n/path/to/cam2.mp4\n/path/to/cam3.mp4",
    placeholder_aligned_dir: "C:\\Video\\aligned или оставить пустым",
    placeholder_multicam_render_output: "C:\\Video\\out или C:\\Video\\master_multicam.mp4",
    system_http: "HTTP",
    system_ffmpeg_missing: "не найден",
    system_ffprobe_missing: "не найден",
    system_unavailable: "Системная информация недоступна",
    unit_seconds_short: "сек",
    unit_milliseconds_short: "мс",
    status_sync_analyzing: "Анализирую смещение...",
    status_sync_rendering: "Запускаю точный рендер...",
    status_multicam_analyzing: "Считаю смещения по камерам...",
    status_multicam_exporting: "Готовлю ffmpeg-команды по камерам...",
    status_multicam_rendering: "Запускаю точный multicam-рендер...",
    label_delay: "Смещение",
    label_confidence: "Уверенность",
    label_video_duration: "Длительность видео",
    label_audio_duration: "Длительность аудио",
    label_render_complete: "Рендер завершен.",
    label_offset_used: "Использованное смещение",
    label_saved_to: "Сохранено в",
    label_elapsed: "Время выполнения",
    label_command: "Команда",
    label_camera: "Камера",
    label_duration: "Длительность",
    label_note: "Примечание",
    label_output: "Выходной файл",
    label_transcript_source: "Источник титров",
    label_strategy: "Стратегия",
    label_multicam_render_complete: "Multicam-рендер завершен.",
    label_timeline_duration: "Длительность таймлайна",
    label_shots: "Количество планов",
    label_shot_plan_preview: "Предпросмотр плана склеек",
    label_more_segments: "... и еще {count} сегментов",
    label_system_manifest: "Встроенные компоненты",
    label_unknown_request_error: "Неизвестная ошибка запроса",
    mode_cpu: "Локальный CPU",
    mode_gpu: "Локальный GPU",
    mode_remote: "Удаленный ffmpeg-over-ip",
  },
  en: {
    document_title: "AutoSync Studio 1.0.21",
    app_title: "AutoSync Studio 1.0.21",
    hero_subtitle: "A rebuilt version focused on exact sync instead of fragile automation.",
    tab_sync: "Single-Cam Sync",
    tab_multicam: "Multicam",
    tab_backend: "Render Backend",
    studio_single_subtitle: "For the case where you have a camera video and a separate master audio file.",
    multicam_subtitle: "First we measure every camera offset against the master audio honestly, and only then build an exact render without hidden magic.",
    backend_panel_subtitle: "A single execution backend for both single-cam and multicam: local CPU, local GPU, or remote ffmpeg-over-ip.",
    label_video_path: "Video path",
    label_video_path_short: "Video",
    label_audio_path: "Master audio path",
    label_audio_path_short: "Audio",
    label_analyze_seconds: "Analysis length in seconds",
    label_analyze_short: "Analysis sec",
    label_max_lag: "Maximum offset search window",
    label_max_lag_short: "Max offset",
    label_output_path: "Output render path",
    label_output_short: "Save",
    label_crf: "CRF",
    label_crf_short: "Quality",
    label_preset: "Encoding preset",
    label_preset_short: "Speed",
    backend_title: "Execution backend",
    label_execution_mode: "Where to render",
    label_execution_short: "Mode",
    label_remote_client_path: "ffmpeg-over-ip client path",
    label_client_short: "Client",
    label_remote_address: "Server address",
    label_remote_address_short: "Address",
    label_remote_secret: "Shared secret",
    label_remote_secret_short: "Secret",
    btn_analyze_sync: "Analyze",
    btn_render_sync: "Render",
    btn_cancel: "Cancel",
    sync_output_idle: "Offset analysis and exact render results will appear here.",
    sync_note: "If the offset is positive, the master audio starts earlier. If it is negative, the camera video starts earlier.",
    sync_terms_note: "`Analysis sec` is how many initial seconds to compare. `Max offset` is the widest desync search window. `Quality (CRF)` controls quality versus file size: lower means better quality. `Speed` controls how fast encoding runs.",
    label_master_audio: "Master audio path",
    label_master_audio_short: "Master",
    label_camera_paths: "Camera files, one per line",
    label_camera_paths_short: "Cameras",
    label_aligned_dir: "Where to save aligned clips",
    label_aligned_dir_short: "Save",
    label_aligned_crf: "CRF for aligned clips",
    label_aligned_crf_short: "CRF",
    label_multicam_render_output: "Where to save the final multicam render",
    label_final_short: "Save",
    label_primary_camera: "Главная камера",
    label_primary_short: "Главная камера",
    label_shot_window: "Shot analysis window, sec",
    label_shot_window_short: "Window",
    label_min_shot: "Minimum shot length, sec",
    label_min_shot_short: "Min Shot",
    multicam_backend_note: "Export and final multicam rendering use the same backend as the single-cam mode.",
    btn_analyze_multicam: "Analyze",
    btn_export_commands: "Export commands",
    btn_render_final: "Render",
    status_render_cancelled: "Render cancelled by user.",
    status_progress: "Progress",
    multicam_output_idle: "Measured offsets, export commands, and shot plan preview will appear here.",
    multicam_note: "This is the new project architecture: reliable diagnosis first, exact rendering second. Auto-editing is layered on top of a correct time model.",
    backend_output_idle: "Bundled component status and the active backend configuration are shown here.",
    browse_btn: "Browse",
    placeholder_video_path: "C:\\Video\\camera.mp4",
    placeholder_audio_path: "C:\\Audio\\master.wav",
    placeholder_output_path: "C:\\Video\\out or C:\\Video\\result.mp4",
    placeholder_remote_client_path: "ffmpeg-over-ip-client",
    placeholder_remote_address: "127.0.0.1:5050",
    placeholder_remote_secret: "shared-secret",
    placeholder_camera_paths: "/path/to/cam1.mp4\n/path/to/cam2.mp4\n/path/to/cam3.mp4",
    placeholder_aligned_dir: "C:\\Video\\aligned or leave empty",
    placeholder_multicam_render_output: "C:\\Video\\out or C:\\Video\\master_multicam.mp4",
    system_http: "HTTP",
    system_ffmpeg_missing: "not found",
    system_ffprobe_missing: "not found",
    system_unavailable: "System info unavailable",
    unit_seconds_short: "sec",
    unit_milliseconds_short: "ms",
    status_sync_analyzing: "Analyzing offset...",
    status_sync_rendering: "Starting exact render...",
    status_multicam_analyzing: "Measuring camera offsets...",
    status_multicam_exporting: "Preparing ffmpeg camera commands...",
    status_multicam_rendering: "Starting exact multicam render...",
    label_delay: "Delay",
    label_confidence: "Confidence",
    label_video_duration: "Video duration",
    label_audio_duration: "Audio duration",
    label_render_complete: "Render complete.",
    label_offset_used: "Offset used",
    label_saved_to: "Saved to",
    label_elapsed: "Elapsed",
    label_command: "Command",
    label_camera: "Camera",
    label_duration: "Duration",
    label_note: "Note",
    label_output: "Output",
    label_transcript_source: "Captions source",
    label_strategy: "Strategy",
    label_multicam_render_complete: "Multicam render complete.",
    label_timeline_duration: "Timeline duration",
    label_shots: "Shots",
    label_shot_plan_preview: "Shot plan preview",
    label_more_segments: "... and {count} more segments",
    label_system_manifest: "Bundled components",
    label_unknown_request_error: "Unknown request error",
    mode_cpu: "Local CPU",
    mode_gpu: "Local GPU",
    mode_remote: "Remote ffmpeg-over-ip",
  },
};

Object.assign(I18N.ru, {
  document_title: "AutoSync Studio",
  app_title: "AutoSync Studio",
  label_primary_camera: "Основная камера",
  label_primary_short: "Основная камера",
  label_preview_short: "Превью",
  preview_full: "Весь файл",
  preview_2min: "2 минуты",
  preview_5min: "5 минут",
  label_edit_mode_short: "Режим",
  edit_mode_classic: "Классический",
  edit_mode_ai: "Smart AI",
  label_analysis_settings: "Параметры анализа",
  label_ai_settings: "AI настройки",
  toggle_show: "Показать",
  toggle_hide: "Скрыть",
  label_assembly_ai_short: "AssemblyAI",
  label_ai_provider_short: "AI",
  ai_provider_off: "Без AI",
  ai_provider_gemini: "Gemini",
  ai_provider_openai: "OpenAI",
  label_ai_key_short: "AI Key",
  label_shorts_short: "Shorts",
  label_ai_prompt_short: "AI Prompt",
  label_shorts_plan_short: "Shorts Plan",
  btn_plan_shorts: "Построить",
  placeholder_assembly_ai_key: "AssemblyAI API key",
  placeholder_gemini_ai_key: "Gemini API key",
  placeholder_openai_ai_key: "ChatGPT / OpenAI API key",
  placeholder_ai_prompt: "Найди эмоционально сильные хайлайты и полезные шортсы",
  status_plan_shorts: "Собираю shorts plan...",
});

Object.assign(I18N.en, {
  document_title: "AutoSync Studio",
  app_title: "AutoSync Studio",
  label_primary_camera: "Primary camera",
  label_primary_short: "Primary camera",
  label_preview_short: "Preview",
  preview_full: "Whole file",
  preview_2min: "2 minutes",
  preview_5min: "5 minutes",
  label_edit_mode_short: "Mode",
  edit_mode_classic: "Classic",
  edit_mode_ai: "Smart AI",
  label_analysis_settings: "Analysis settings",
  label_ai_settings: "AI settings",
  toggle_show: "Show",
  toggle_hide: "Hide",
  label_assembly_ai_short: "AssemblyAI",
  label_ai_provider_short: "AI",
  ai_provider_off: "Off",
  ai_provider_gemini: "Gemini",
  ai_provider_openai: "OpenAI",
  label_ai_key_short: "AI Key",
  label_shorts_short: "Shorts",
  label_ai_prompt_short: "AI Prompt",
  label_shorts_plan_short: "Shorts Plan",
  btn_plan_shorts: "Build",
  placeholder_assembly_ai_key: "AssemblyAI API key",
  placeholder_gemini_ai_key: "Gemini API key",
  placeholder_openai_ai_key: "ChatGPT / OpenAI API key",
  placeholder_ai_prompt: "Find emotionally strong highlights and useful short clips",
  status_plan_shorts: "Building shorts plan...",
});

const syncOutput = document.getElementById("syncOutput");
let shortsOutput = document.getElementById("shortsOutput");
const multicamOutput = document.getElementById("multicamOutput");
const backendOutput = document.getElementById("backendOutput");
const langRuBtn = document.getElementById("langRuBtn");
const langEnBtn = document.getElementById("langEnBtn");
const tabs = ["Sync", "Shorts", "Multicam", "Backend"];

let currentLanguage = loadLanguage();
let lastDelaySeconds = null;
let lastMulticamResult = null;
let currentTab = "Sync";
let activeRenderOutput = null;
let currentSystemDisplayName = "";
let lastBackendStatus = null;
let lastRemoteToolsStatus = null;
let lastShortsPlan = null;
let shortsActionActive = false;

const SHORTS_FORMAT_PRESETS = [
  { id: "youtube-shorts", labelKey: "shorts_format_youtube" },
  { id: "tiktok", labelKey: "shorts_format_tiktok" },
  { id: "instagram-reels", labelKey: "shorts_format_instagram" },
];

Object.assign(I18N.ru, {
  confirm_sync_preview_render: "Budet sdelan tolko preview-render single-cam na {seconds}. Prodolzhit?",
  confirm_multicam_preview_render: "Budet sdelan tolko preview-render multicam na {seconds}. Prodolzhit?",
  label_preview_render: "Preview render",
  label_offsets_source_cached: "Offsets reused from the last Analyze",
  label_offsets_source_fresh: "Offsets recalculated during Render",
  btn_check_backend: "Проверить соединение",
  backend_status_mode: "Режим",
  backend_status_client: "Клиент",
  backend_status_server: "Сервер",
  backend_status_idle: "Здесь можно проверить доступность render backend до запуска рендера.",
  backend_status_checking: "Проверяю backend...",
  backend_status_ready: "Remote backend готов к работе.",
  backend_status_needs_attention: "Remote backend требует внимания.",
  backend_status_local_ready: "Локальный backend готов к работе.",
  backend_status_mode_cpu: "Локальный CPU",
  backend_status_mode_gpu: "Локальный GPU",
  backend_status_mode_remote: "Удаленный ffmpeg-over-ip",
  backend_status_not_checked: "Не проверено",
});

Object.assign(I18N.en, {
  confirm_sync_preview_render: "This will render only a single-cam preview for {seconds}. Continue?",
  confirm_multicam_preview_render: "This will render only a multicam preview for {seconds}. Continue?",
  label_preview_render: "Preview render",
  label_offsets_source_cached: "Offsets reused from the last Analyze",
  label_offsets_source_fresh: "Offsets recalculated during Render",
  btn_check_backend: "Check connection",
  backend_status_mode: "Mode",
  backend_status_client: "Client",
  backend_status_server: "Server",
  backend_status_idle: "Check render backend availability before starting a render.",
  backend_status_checking: "Checking backend...",
  backend_status_ready: "Remote backend is ready.",
  backend_status_needs_attention: "Remote backend needs attention.",
  backend_status_local_ready: "Local backend is ready.",
  backend_status_mode_cpu: "Local CPU",
  backend_status_mode_gpu: "Local GPU",
  backend_status_mode_remote: "Remote ffmpeg-over-ip",
  backend_status_not_checked: "Not checked",
});

Object.assign(I18N.ru, {
  btn_update_remote_tools: "Обновить ffmpeg-over-ip",
  remote_tools_idle: "ffmpeg-over-ip будет автоматически установлен и обновлён внутри папки программы.",
  remote_tools_checking: "Проверяю ffmpeg-over-ip...",
  remote_tools_updating: "Скачиваю и обновляю ffmpeg-over-ip...",
  remote_tools_ready: "Установлена версия {installed}.",
  remote_tools_update_available: "Установлена {installed}. Доступна новая версия {available}.",
  remote_tools_updated: "ffmpeg-over-ip обновлён до {installed}.",
  remote_tools_error: "Не удалось проверить обновления ffmpeg-over-ip.",
});

Object.assign(I18N.en, {
  btn_update_remote_tools: "Update ffmpeg-over-ip",
  remote_tools_idle: "ffmpeg-over-ip will be installed and updated automatically inside the app folder.",
  remote_tools_checking: "Checking ffmpeg-over-ip...",
  remote_tools_updating: "Downloading and updating ffmpeg-over-ip...",
  remote_tools_ready: "Installed version: {installed}.",
  remote_tools_update_available: "Installed {installed}. New version {available} is available.",
  remote_tools_updated: "ffmpeg-over-ip updated to {installed}.",
  remote_tools_error: "Unable to check ffmpeg-over-ip updates.",
});

Object.assign(I18N.ru, {
  tab_shorts: "Shorts / Reels",
  label_shorts_reels: "Shorts / Reels",
  shorts_hint: "Загрузите готовое интервью, постройте план клипов и экспортируйте выбранные форматы для соцсетей.",
  shorts_source_hint: "Эта вкладка работает отдельно: сюда можно загрузить готовое интервью и сразу сделать shorts / reels.",
  label_speech_to_text_short: "AssemblyAI",
  label_idea_model_short: "AI",
  label_gemini_ai_key_short: "Gemini API",
  label_openai_ai_key_short: "ChatGPT API",
  label_clips_count_short: "Кол-во",
  label_what_to_look_for_short: "Промпт",
  label_source_video_short: "Видео",
  label_source_audio_short: "Мастер-аудио",
  label_captions_short: "Титры",
  label_subtitle_font_short: "Шрифт",
  label_subtitle_bg_short: "Фон",
  label_subtitle_bg_opacity_short: "Прозрачность",
  captions_off: "ВЫКЛ",
  captions_burned_in: "ВКЛ",
  subtitle_font_segoe: "Segoe UI",
  subtitle_font_montserrat: "Montserrat",
  subtitle_font_arial: "Arial",
  subtitle_font_verdana: "Verdana",
  subtitle_font_tahoma: "Tahoma",
  subtitle_font_trebuchet: "Trebuchet MS",
  subtitle_font_georgia: "Georgia",
  label_formats_short: "Formats",
  label_shorts_output_short: "Output folder",
  btn_build_plan: "Построить план",
  btn_render_selected_shorts: "Рендер выбранных",
  btn_render_full_subtitled: "Титры",
  shorts_plan_idle: "План шортсов появится здесь после Build plan.",
  shorts_output_idle: "Здесь появятся план клипов и результаты batch-экспорта shorts / reels.",
  shorts_review_title: "Найденные клипы",
  shorts_review_empty: "Сначала постройте план клипов.",
  shorts_plan_progress_idle: "Готов к построению плана",
  shorts_select_required: "Выберите хотя бы один клип.",
  shorts_rendering: "Рендерю shorts / reels...",
  shorts_render_complete: "Экспорт shorts / reels завершен.",
  shorts_full_requires_video: "Укажите Видео для полной версии с титрами.",
  shorts_full_requires_ai_key: "Укажите ключ AssemblyAI для титров.",
  shorts_full_requires_captions: "Для полной версии с титрами включите Титры = ВКЛ.",
  shorts_full_rendering: "Рендерю полную версию интервью с титрами...",
  shorts_full_render_complete: "Экспорт полной версии интервью с титрами завершен.",
  label_srt_file: "SRT",
  label_transcript_text: "TXT",
  label_ass_file: "ASS",
  shorts_output_default: "Рядом с исходным интервью будет создана папка *_shorts.",
  shorts_output_required: "Укажите папку для экспорта shorts / reels.",
  placeholder_shorts_output: "C:\\Video\\interview_shorts",
  status_plan_shorts: "Строю план shorts / reels...",
  shorts_plan_source_video: "Таймлайн: аудио из самого видео.",
  shorts_plan_source_audio: "Таймлайн: отдельный master audio, sync измерен автоматически.",
  shorts_format_youtube: "YouTube Shorts 9:16",
  shorts_format_tiktok: "TikTok 9:16",
  shorts_format_instagram: "Instagram Reels 9:16",
  shorts_format_square: "Square 1:1",
  shorts_format_feed: "Feed 4:5",
  shorts_format_story: "Story 9:16",
  shorts_format_horizontal: "Horizontal 16:9",
});

Object.assign(I18N.en, {
  tab_shorts: "Shorts / Reels",
  label_shorts_reels: "Shorts / Reels",
  shorts_hint: "Load a finished interview, build a clip plan, then export the selected social formats.",
  shorts_source_hint: "This tab works on its own: load a finished interview here and export shorts / reels directly.",
  label_speech_to_text_short: "AssemblyAI",
  label_idea_model_short: "Idea model",
  label_gemini_ai_key_short: "Gemini API",
  label_openai_ai_key_short: "ChatGPT API",
  label_clips_count_short: "How many clips",
  label_what_to_look_for_short: "What to look for",
  label_source_video_short: "Interview video",
  label_source_audio_short: "Master audio",
  label_captions_short: "Captions",
  label_subtitle_font_short: "Font",
  label_subtitle_bg_short: "Background",
  label_subtitle_bg_opacity_short: "Opacity",
  captions_off: "Off",
  captions_burned_in: "Burned-in",
  subtitle_font_segoe: "Segoe UI",
  subtitle_font_montserrat: "Montserrat",
  subtitle_font_arial: "Arial",
  subtitle_font_verdana: "Verdana",
  subtitle_font_tahoma: "Tahoma",
  subtitle_font_trebuchet: "Trebuchet MS",
  subtitle_font_georgia: "Georgia",
  label_formats_short: "Formats",
  label_shorts_output_short: "Output folder",
  btn_build_plan: "Build plan",
  btn_render_selected_shorts: "Render selected",
  btn_render_full_subtitled: "Captions",
  shorts_plan_idle: "Your shorts plan will appear here after Build plan.",
  shorts_output_idle: "Your clip plan and batch export results for shorts / reels will appear here.",
  shorts_review_title: "Planned clips",
  shorts_review_empty: "Build a clip plan first.",
  shorts_plan_progress_idle: "Ready to build the plan",
  shorts_select_required: "Select at least one clip.",
  shorts_rendering: "Rendering shorts / reels...",
  shorts_render_complete: "Shorts / reels export complete.",
  shorts_full_requires_video: "Choose a video file for the full subtitled export.",
  shorts_full_requires_ai_key: "Enter your AssemblyAI key for subtitles.",
  shorts_full_requires_captions: "Turn captions ON to export the full interview with subtitles.",
  shorts_full_rendering: "Rendering the full interview with captions...",
  shorts_full_render_complete: "Full interview export with captions complete.",
  label_srt_file: "SRT",
  label_transcript_text: "TXT",
  label_ass_file: "ASS",
  shorts_output_default: "A *_shorts folder will be created next to the source interview.",
  shorts_output_required: "Choose an output folder for shorts / reels export.",
  placeholder_shorts_output: "C:\\Video\\interview_shorts",
  status_plan_shorts: "Building shorts / reels plan...",
  shorts_plan_source_video: "Timeline source: audio from the video file.",
  shorts_plan_source_audio: "Timeline source: external master audio, sync measured automatically.",
  shorts_format_youtube: "YouTube Shorts 9:16",
  shorts_format_tiktok: "TikTok 9:16",
  shorts_format_instagram: "Instagram Reels 9:16",
  shorts_format_square: "Square 1:1",
  shorts_format_feed: "Feed 4:5",
  shorts_format_story: "Story 9:16",
  shorts_format_horizontal: "Horizontal 16:9",
});

function normalizeComparablePath(value) {
  return String(value || "").trim().replace(/\//g, "\\").toLowerCase();
}

function previewLabel(seconds) {
  if (seconds === 120) {
    return t("preview_2min");
  }
  if (seconds === 300) {
    return t("preview_5min");
  }
  return `${seconds} ${t("unit_seconds_short")}`;
}

function confirmPreviewRender(seconds, mode) {
  if (!(seconds > 0)) {
    return true;
  }
  const key = mode === "multicam" ? "confirm_multicam_preview_render" : "confirm_sync_preview_render";
  return window.confirm(t(key, { seconds: previewLabel(seconds) }));
}

function loadLanguage() {
  const saved = localStorage.getItem(STORAGE_LANGUAGE_KEY);
  return saved && I18N[saved] ? saved : DEFAULT_LANGUAGE;
}

function loadBackendPrefs() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_BACKEND_PREFS_KEY) || "{}");
  } catch (_) {
    return {};
  }
}

function loadShortsPrefs() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_SHORTS_PREFS_KEY) || "{}");
    return {
      aiPrompt: typeof saved.aiPrompt === "string" ? saved.aiPrompt : "",
      aiProvider: typeof saved.aiProvider === "string" ? saved.aiProvider : "",
      shortsCount: Number(saved.shortsCount || 3),
      captionsMode: saved.captionsMode || "off",
      subtitleFont: typeof saved.subtitleFont === "string" ? saved.subtitleFont : "segoe-ui",
      subtitleBgColor: typeof saved.subtitleBgColor === "string" ? saved.subtitleBgColor : "#000000",
      subtitleBgOpacity: Number(saved.subtitleBgOpacity ?? 50),
      formats: Array.isArray(saved.formats)
        ? saved.formats
        : ["youtube-shorts", "instagram-reels", "tiktok"],
      outputDir: typeof saved.outputDir === "string" ? saved.outputDir : "",
      videoPath: typeof saved.videoPath === "string" ? saved.videoPath : "",
      audioPath: typeof saved.audioPath === "string" ? saved.audioPath : "",
    };
  } catch (_) {
    return {
      aiPrompt: "",
      aiProvider: "",
      shortsCount: 3,
      captionsMode: "off",
      subtitleFont: "segoe-ui",
      subtitleBgColor: "#000000",
      subtitleBgOpacity: 50,
      formats: ["youtube-shorts", "instagram-reels", "tiktok"],
      outputDir: "",
      videoPath: "",
      audioPath: "",
    };
  }
}

function saveShortsPrefs() {
  const payload = {
    aiPrompt: document.getElementById("shortsPrompt")?.value || "",
    aiProvider: document.getElementById("aiProvider")?.value || "",
    shortsCount: Number(document.getElementById("shortsCount")?.value || 3),
    captionsMode: document.getElementById("shortsCaptionsMode")?.value || "off",
    subtitleFont: document.getElementById("shortsSubtitleFont")?.value || "segoe-ui",
    subtitleBgColor: document.getElementById("shortsSubtitleBgColor")?.value || "#000000",
    subtitleBgOpacity: Number(document.getElementById("shortsSubtitleBgOpacity")?.value ?? 50),
    formats: Array.from(document.querySelectorAll(".shorts-format-checkbox:checked")).map((node) => node.value),
    outputDir: document.getElementById("shortsOutputDir")?.value.trim() || "",
    videoPath: document.getElementById("shortsVideoPath")?.value.trim() || "",
    audioPath: document.getElementById("shortsAudioPath")?.value.trim() || "",
  };
  localStorage.setItem(STORAGE_SHORTS_PREFS_KEY, JSON.stringify(payload));
}

function applyBackendPrefs() {
  const prefs = loadBackendPrefs();
  const executionModeNode = document.getElementById("executionMode");
  const remoteAddressNode = document.getElementById("remoteAddress");

  if (executionModeNode && typeof prefs.executionMode === "string" && prefs.executionMode) {
    executionModeNode.value = prefs.executionMode;
  }
  if (remoteAddressNode) {
    if (typeof prefs.remoteAddress === "string" && prefs.remoteAddress) {
      remoteAddressNode.value = prefs.remoteAddress;
    } else if (!remoteAddressNode.value.trim()) {
      remoteAddressNode.value = "127.0.0.1:5050";
    }
  }
}

function saveBackendPrefs() {
  const payload = {
    executionMode: document.getElementById("executionMode")?.value || "cpu",
    remoteAddress: document.getElementById("remoteAddress")?.value.trim() || "",
  };
  localStorage.setItem(STORAGE_BACKEND_PREFS_KEY, JSON.stringify(payload));
}

async function loadStoredSecrets() {
  const assemblyAiKeyNode = document.getElementById("assemblyAiKey");
  const geminiAiKeyNode = document.getElementById("geminiAiKey");
  const openAiKeyNode = document.getElementById("openAiKey");
  try {
    const settings = await request("/api/settings");
    if (assemblyAiKeyNode) {
      assemblyAiKeyNode.value = settings.assemblyAiKey || "";
    }
    if (geminiAiKeyNode) {
      geminiAiKeyNode.value = settings.geminiAiKey || settings.aiKey || "";
    }
    if (openAiKeyNode) {
      openAiKeyNode.value = settings.openAiKey || settings.aiKey || "";
    }
  } catch (_) {
  }
  const persist = async () => {
    try {
      await request("/api/settings", {
        assemblyAiKey: assemblyAiKeyNode ? assemblyAiKeyNode.value : "",
        geminiAiKey: geminiAiKeyNode ? geminiAiKeyNode.value : "",
        openAiKey: openAiKeyNode ? openAiKeyNode.value : "",
      });
    } catch (_) {
    }
  };
  if (assemblyAiKeyNode) {
    assemblyAiKeyNode.addEventListener("change", persist);
    assemblyAiKeyNode.addEventListener("blur", persist);
  }
  if (geminiAiKeyNode) {
    geminiAiKeyNode.addEventListener("change", persist);
    geminiAiKeyNode.addEventListener("blur", persist);
  }
  if (openAiKeyNode) {
    openAiKeyNode.addEventListener("change", persist);
    openAiKeyNode.addEventListener("blur", persist);
  }
}

function t(key, replacements = {}) {
  const dict = I18N[currentLanguage] || I18N[DEFAULT_LANGUAGE];
  let template = dict[key] ?? I18N[DEFAULT_LANGUAGE][key] ?? key;
  Object.entries(replacements).forEach(([name, value]) => {
    template = template.replaceAll(`{${name}}`, String(value));
  });
  return template;
}

function setNodeText(selector, text) {
  const node = typeof selector === "string" ? document.querySelector(selector) : selector;
  if (node) {
    node.textContent = text;
  }
}

function updateTitleFromState() {
  const displayName = currentSystemDisplayName || t("app_title");
  document.title = displayName;
  setNodeText("#appTitle", displayName);
}

function setSelectOptionText(selectId, labelsByValue) {
  const select = document.getElementById(selectId);
  if (!select) {
    return;
  }
  Array.from(select.options).forEach((option) => {
    if (Object.prototype.hasOwnProperty.call(labelsByValue, option.value)) {
      option.textContent = labelsByValue[option.value];
    }
  });
}

function configureDetailsSummary(details, labelText) {
  if (!details) {
    return;
  }
  const summary = details.querySelector("summary");
  if (!summary) {
    return;
  }

  let labelNode = summary.querySelector(".advanced-toggle-label");
  if (!labelNode) {
    labelNode = Array.from(summary.children).find((node) => (
      node.tagName === "SPAN" && !node.classList.contains("advanced-toggle-state")
    ));
    if (labelNode) {
      labelNode.classList.add("advanced-toggle-label");
    } else {
      const existingText = summary.textContent.trim();
      summary.textContent = "";
      labelNode = document.createElement("span");
      labelNode.className = "advanced-toggle-label";
      labelNode.textContent = existingText;
      summary.appendChild(labelNode);
    }
  }
  labelNode.textContent = labelText;

  let stateNode = summary.querySelector(".advanced-toggle-state");
  if (!stateNode) {
    stateNode = document.createElement("span");
    stateNode.className = "advanced-toggle-state";
    summary.appendChild(stateNode);
  }
  stateNode.textContent = details.open ? t("toggle_hide") : t("toggle_show");

  if (!details.dataset.toggleBound) {
    details.addEventListener("toggle", () => {
      stateNode.textContent = details.open ? t("toggle_hide") : t("toggle_show");
    });
    details.dataset.toggleBound = "true";
  }
}

function normalizeSyncLayout() {
  const qualityGroup = document.querySelector("#crf")?.closest(".input-group");
  const speedGroup = document.querySelector("#preset")?.closest(".input-group");
  const previewGroup = document.querySelector("#previewMode")?.closest(".input-group");
  const optionsGrid = qualityGroup?.parentElement;

  if (
    !qualityGroup ||
    !speedGroup ||
    !previewGroup ||
    !optionsGrid ||
    !optionsGrid.classList.contains("options-grid")
  ) {
    return;
  }

  if (!optionsGrid.dataset.normalized) {
    optionsGrid.replaceWith(qualityGroup, speedGroup, previewGroup);
    optionsGrid.dataset.normalized = "true";
  }
}

function applyStaticLabels() {
  setNodeText("#analysisSettingsLabel", t("label_analysis_settings"));
  setNodeText("#analyzeSecondsLabel", t("label_analyze_short"));
  setNodeText("#maxLagSecondsLabel", t("label_max_lag_short"));

  setNodeText(document.querySelector("#previewMode")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_preview_short"));
  setNodeText(document.querySelector("#multicamPreviewMode")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_preview_short"));
  setNodeText(document.querySelector("#editMode")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_edit_mode_short"));
  setNodeText(document.querySelector("#assemblyAiKey")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_speech_to_text_short"));
  setNodeText(document.querySelector("#aiProvider")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_idea_model_short"));
  setNodeText(document.querySelector("#geminiAiKey")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_gemini_ai_key_short"));
  setNodeText(document.querySelector("#openAiKey")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_openai_ai_key_short"));
  setNodeText(document.querySelector("#shortsCount")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_clips_count_short"));
  setNodeText(document.querySelector("#shortsPrompt")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_what_to_look_for_short"));
  setNodeText(document.querySelector("#shortsVideoPath")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_source_video_short"));
  setNodeText(document.querySelector("#shortsAudioPath")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_source_audio_short"));
  setNodeText(document.querySelector("#shortsOutputDir")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_shorts_output_short"));
  setNodeText(document.querySelector("#shortsCaptionsMode")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_captions_short"));
  setNodeText(document.querySelector("#shortsSubtitleFont")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_subtitle_font_short"));
  setNodeText(document.querySelector("#shortsSubtitleBgColor")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_subtitle_bg_short"));
  setNodeText(document.querySelector("#shortsSubtitleBgOpacity")?.closest(".input-group")?.querySelector(".input-prefix"), t("label_subtitle_bg_opacity_short"));
  setNodeText(document.querySelector("#shortsFormatsGroup")?.querySelector(".input-prefix"), t("label_formats_short"));

  const assemblyAiKeyNode = document.getElementById("assemblyAiKey");
  if (assemblyAiKeyNode) {
    assemblyAiKeyNode.setAttribute("placeholder", t("placeholder_assembly_ai_key"));
  }
  const geminiAiKeyNode = document.getElementById("geminiAiKey");
  if (geminiAiKeyNode) {
    geminiAiKeyNode.setAttribute("placeholder", t("placeholder_gemini_ai_key"));
  }
  const openAiKeyNode = document.getElementById("openAiKey");
  if (openAiKeyNode) {
    openAiKeyNode.setAttribute("placeholder", t("placeholder_openai_ai_key"));
  }
  const aiPromptNode = document.getElementById("shortsPrompt");
  if (aiPromptNode) {
    aiPromptNode.setAttribute("placeholder", t("placeholder_ai_prompt"));
  }
  const shortsVideoPathNode = document.getElementById("shortsVideoPath");
  if (shortsVideoPathNode) {
    shortsVideoPathNode.setAttribute("placeholder", t("placeholder_video_path"));
  }
  const shortsAudioPathNode = document.getElementById("shortsAudioPath");
  if (shortsAudioPathNode) {
    shortsAudioPathNode.setAttribute("placeholder", t("placeholder_audio_path"));
  }
  const shortsOutputDirNode = document.getElementById("shortsOutputDir");
  if (shortsOutputDirNode) {
    shortsOutputDirNode.setAttribute("placeholder", t("placeholder_shorts_output"));
  }

  setNodeText("#planShortsBtn", t("btn_build_plan"));
  setNodeText("#renderShortsBtn", t("btn_render_selected_shorts"));
  setNodeText("#renderFullShortsBtn", t("btn_render_full_subtitled"));
  setNodeText("#cancelShortsBtn", t("btn_cancel"));
  setNodeText("#shortsReviewTitle", t("shorts_review_title"));
  setNodeText("#shortsReviewEmpty", t("shorts_review_empty"));
  configureDetailsSummary(document.querySelector("#viewSync .advanced-toggle"), t("label_analysis_settings"));
  configureDetailsSummary(document.querySelector("#viewBackend .advanced-toggle"), t("label_ai_settings"));

  setSelectOptionText("previewMode", {
    0: t("preview_full"),
    120: t("preview_2min"),
    300: t("preview_5min"),
  });
  setSelectOptionText("multicamPreviewMode", {
    0: t("preview_full"),
    120: t("preview_2min"),
    300: t("preview_5min"),
  });
  setSelectOptionText("editMode", {
    classic: t("edit_mode_classic"),
    ai: t("edit_mode_ai"),
  });
  setSelectOptionText("aiProvider", {
    "": t("ai_provider_off"),
    gemini: t("ai_provider_gemini"),
    openai: t("ai_provider_openai"),
  });
  setSelectOptionText("shortsCaptionsMode", {
    off: t("captions_off"),
    "burned-in": t("captions_burned_in"),
  });
  setSelectOptionText("shortsSubtitleFont", {
    "segoe-ui": t("subtitle_font_segoe"),
    montserrat: t("subtitle_font_montserrat"),
    arial: t("subtitle_font_arial"),
    verdana: t("subtitle_font_verdana"),
    tahoma: t("subtitle_font_tahoma"),
    "trebuchet-ms": t("subtitle_font_trebuchet"),
    georgia: t("subtitle_font_georgia"),
  });

  document.querySelectorAll(".shorts-format-option span").forEach((node, index) => {
    const preset = SHORTS_FORMAT_PRESETS[index];
    if (preset) {
      node.textContent = t(preset.labelKey);
    }
  });
}

function setLanguage(language) {
  if (!I18N[language]) {
    return;
  }
  currentLanguage = language;
  localStorage.setItem(STORAGE_LANGUAGE_KEY, language);
  document.documentElement.lang = language;

  document.querySelectorAll("[data-i18n]").forEach((node) => {
    node.textContent = t(node.dataset.i18n);
  });

  document.querySelectorAll("[data-i18n-placeholder]").forEach((node) => {
    node.setAttribute("placeholder", t(node.dataset.i18nPlaceholder));
  });

  document.querySelectorAll("[data-i18n-option]").forEach((node) => {
    node.textContent = t(node.dataset.i18nOption);
  });

  ensureShortsTabUI();
  applyStaticLabels();
  updateTitleFromState();
  langRuBtn.classList.toggle("active", language === "ru");
  langEnBtn.classList.toggle("active", language === "en");
  if (lastBackendStatus) {
    applyBackendStatus(lastBackendStatus);
  } else {
    resetBackendStatusCards();
  }
  ensureRemoteToolsUI();
  if (lastRemoteToolsStatus) {
    applyRemoteToolsStatus(lastRemoteToolsStatus);
  }
  if (lastShortsPlan) {
    renderShortsPlanReview(lastShortsPlan);
  }
}

function switchTab(tabName) {
  currentTab = tabName;
  tabs.forEach((tab) => {
    document.getElementById(`tab${tab}Btn`).classList.toggle("active", tab === tabName);
    document.getElementById(`view${tab}`).classList.toggle("active", tab === tabName);
  });
  syncOutput.classList.toggle("active", tabName === "Sync");
  if (shortsOutput) {
    shortsOutput.classList.toggle("active", tabName === "Shorts");
  }
  multicamOutput.classList.toggle("active", tabName === "Multicam");
  backendOutput.classList.toggle("active", tabName === "Backend");
  if (tabName === "Backend" && !lastRemoteToolsStatus) {
    loadRemoteToolsStatus(false);
  }
}

async function request(url, payload) {
  const response = await fetch(url, {
    method: payload ? "POST" : "GET",
    headers: payload ? { "Content-Type": "application/json" } : undefined,
    body: payload ? JSON.stringify(payload) : undefined,
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || t("label_unknown_request_error"));
  }
  return data;
}

async function streamRequest(url, payload, onEvent) {
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || t("label_unknown_request_error"));
  }
  if (!response.body) {
    throw new Error("Streaming is unavailable in this runtime.");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split("\n");
    buffer = lines.pop() || "";
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed) {
        continue;
      }
      const event = JSON.parse(trimmed);
      if (event.error) {
        throw new Error(event.error);
      }
      onEvent(event);
    }
  }

  if (buffer.trim()) {
    const event = JSON.parse(buffer.trim());
    if (event.error) {
      throw new Error(event.error);
    }
    onEvent(event);
  }
}

function buildProgressText(baseMessage, event, fallbackDoneText) {
  const lines = [baseMessage];
  if (typeof event.percent === "number" && Number.isFinite(event.percent)) {
    lines.push(`${t("status_progress")}: ${event.percent.toFixed(1)}%`);
  }
  if (event.message) {
    lines.push(event.message);
  }
  if (event.done) {
    lines.push(fallbackDoneText);
  }
  if (event.outputPath) {
    lines.push(`${t("label_saved_to")}: ${event.outputPath}`);
  }
  if (event.duration) {
    lines.push(`${t("label_elapsed")}: ${event.duration}`);
  }
  if (event.command) {
    lines.push("");
    lines.push(`${t("label_command")}:`);
    lines.push(event.command);
  }
  return lines.join("\n");
}

async function runStreamedRender(url, payload, outputNode, baseMessage, fallbackDoneText, onDone) {
  activeRenderOutput = outputNode;
  setOutput(outputNode, baseMessage, false);
  let lastEvent = null;
  try {
    await streamRequest(url, payload, (event) => {
      lastEvent = event;
      setOutput(outputNode, buildProgressText(baseMessage, event, fallbackDoneText), false);
    });
    if (!lastEvent || !lastEvent.done) {
      throw new Error("Render stream finished unexpectedly.");
    }
    if (onDone) {
      onDone(lastEvent);
    }
  } catch (error) {
    if (lastEvent && lastEvent.done) {
      if (onDone) {
        onDone(lastEvent);
      }
      return;
    }
    throw error;
  } finally {
    activeRenderOutput = null;
  }
}

async function pickFile(kind) {
  const result = await request("/api/pick-file", { kind });
  return (result.path || "").trim();
}

async function pickDirectory() {
  const result = await request("/api/pick-directory", {});
  return (result.path || "").trim();
}

async function pickSavePath(kind, path) {
  const result = await request("/api/pick-save", { kind, path });
  return (result.path || "").trim();
}

async function pathExists(path) {
  const result = await request("/api/path-exists", { path });
  return !!result.exists;
}

function setOutput(node, text, isError = false) {
  node.textContent = text;
  node.classList.toggle("error", isError);
}

function setBackendStatusCard(id, text, tone = "") {
  const node = document.getElementById(id);
  if (!node) {
    return;
  }
  node.textContent = text;
  node.className = `backend-status-value${tone ? ` ${tone}` : ""}`;
}

function resetBackendStatusCards() {
  setBackendStatusCard("backendModeStatus", t("backend_status_not_checked"));
  setBackendStatusCard("backendClientStatus", t("backend_status_not_checked"));
  setBackendStatusCard("backendServerStatus", t("backend_status_not_checked"));
  const note = document.getElementById("backendStatusNote");
  if (note) {
    note.textContent = t("backend_status_idle");
  }
  lastBackendStatus = null;
}

function applyBackendStatus(status) {
  lastBackendStatus = status;
  const modeText =
    status.mode === "gpu"
      ? t("backend_status_mode_gpu")
      : status.mode === "remote"
        ? t("backend_status_mode_remote")
        : t("backend_status_mode_cpu");
  const overallTone = status.overallStatus === "ok" ? "ok" : status.overallStatus === "warn" ? "warn" : "error";

  setBackendStatusCard("backendModeStatus", modeText, overallTone);
  setBackendStatusCard("backendClientStatus", status.clientStatus || t("backend_status_not_checked"), status.clientFound ? "ok" : "error");
  setBackendStatusCard(
    "backendServerStatus",
    status.serverStatus || t("backend_status_not_checked"),
    status.serverReachable ? "ok" : status.mode === "remote" ? "warn" : "ok",
  );

  const note = document.getElementById("backendStatusNote");
  if (note) {
    note.textContent = status.message || t("backend_status_idle");
  }
}

function ensureRemoteToolsUI() {
  const clientPathGroup = document.getElementById("remoteClientPathGroup");
  if (clientPathGroup) {
    clientPathGroup.style.display = "none";
  }

  const backendIntro = document.querySelector('#viewBackend [data-i18n="backend_panel_subtitle"]');
  if (backendIntro) {
    backendIntro.style.display = "none";
  }

  const backendSummaryNote = document.querySelector('[data-i18n="multicam_backend_note"]');
  if (backendSummaryNote) {
    backendSummaryNote.style.display = "none";
  }

  const actions = document.querySelector("#viewBackend .backend-status-actions");
  if (actions && !document.getElementById("updateRemoteToolsBtn")) {
    const button = document.createElement("button");
    button.id = "updateRemoteToolsBtn";
    button.type = "button";
    button.className = "run-btn secondary compact";
    button.dataset.i18n = "btn_update_remote_tools";
    button.textContent = t("btn_update_remote_tools");
    actions.appendChild(button);
  }

  if (actions && !document.getElementById("remoteToolsNote")) {
    const note = document.createElement("div");
    note.id = "remoteToolsNote";
    note.className = "mini-note";
    note.dataset.i18n = "remote_tools_idle";
    note.textContent = t("remote_tools_idle");
    actions.insertAdjacentElement("afterend", note);
  }

  const remoteToolsNote = document.getElementById("remoteToolsNote");
  if (remoteToolsNote) {
    remoteToolsNote.style.display = "none";
  }

  const backendStatusNote = document.getElementById("backendStatusNote");
  if (backendStatusNote) {
    backendStatusNote.style.display = "none";
  }
}

function describeRemoteToolsStatus(status) {
  if (!status) {
    return t("remote_tools_idle");
  }
  if (status.updateAvailable && status.availableVersion) {
    return t("remote_tools_update_available", {
      installed: status.installedVersion || "unknown",
      available: status.availableVersion,
    });
  }
  if (status.installedVersion) {
    return t("remote_tools_ready", { installed: status.installedVersion });
  }
  return t("remote_tools_idle");
}

function applyRemoteToolsStatus(status, { showOutput = false, updated = false, isError = false } = {}) {
  lastRemoteToolsStatus = status || null;
  const note = document.getElementById("remoteToolsNote");
  if (note) {
    note.textContent = isError
      ? (status?.lastError || t("remote_tools_error"))
      : (updated && status?.installedVersion
        ? t("remote_tools_updated", { installed: status.installedVersion })
        : describeRemoteToolsStatus(status));
  }
  if (!showOutput) {
    return;
  }
  if (isError) {
    setOutput(backendOutput, status?.lastError || t("remote_tools_error"), true);
    return;
  }
  const lines = [
    updated && status?.installedVersion
      ? t("remote_tools_updated", { installed: status.installedVersion })
      : describeRemoteToolsStatus(status),
  ];
  if (status?.managedRoot) {
    lines.push(`Path: ${status.managedRoot}`);
  }
  if (status?.lastError) {
    lines.push(`Note: ${status.lastError}`);
  }
  setOutput(backendOutput, lines.filter(Boolean).join("\n"), false);
}

async function loadRemoteToolsStatus(showOutput = false) {
  ensureRemoteToolsUI();
  const note = document.getElementById("remoteToolsNote");
  if (note) {
    note.textContent = t("remote_tools_checking");
  }
  try {
    const status = await request("/api/ffmpeg-over-ip-tools");
    applyRemoteToolsStatus(status, { showOutput });
  } catch (error) {
    applyRemoteToolsStatus({ lastError: error.message }, { showOutput, isError: true });
  }
}

async function updateRemoteTools(showOutput = true) {
  ensureRemoteToolsUI();
  const note = document.getElementById("remoteToolsNote");
  if (note) {
    note.textContent = t("remote_tools_updating");
  }
  try {
    const status = await request("/api/update-ffmpeg-over-ip-tools", {});
    applyRemoteToolsStatus(status, { showOutput, updated: true });
    await loadSystem(false);
  } catch (error) {
    applyRemoteToolsStatus({ lastError: error.message }, { showOutput, isError: true });
  }
}

function fmtSeconds(seconds) {
  const ms = Math.round(seconds * 1000);
  return `${seconds.toFixed(3)} ${t("unit_seconds_short")} (${ms} ${t("unit_milliseconds_short")})`;
}

function currentSyncPayload() {
  return {
    videoPath: document.getElementById("videoPath").value.trim(),
    audioPath: document.getElementById("audioPath").value.trim(),
    analyzeSeconds: Number(document.getElementById("analyzeSeconds").value || 180),
    maxLagSeconds: Number(document.getElementById("maxLagSeconds").value || 12),
  };
}

function currentBackendPayload() {
  return {
    executionMode: document.getElementById("executionMode").value,
    remoteAddress: document.getElementById("remoteAddress").value.trim(),
    remoteSecret: document.getElementById("remoteSecret").value.trim(),
  };
}

async function checkBackendConnection(showOutput = true) {
  const note = document.getElementById("backendStatusNote");
  if (note) {
    note.textContent = t("backend_status_checking");
  }
  try {
    const status = await request("/api/backend-status", currentBackendPayload());
    applyBackendStatus(status);
    if (showOutput) {
      const lines = [
        status.message || "",
        `Mode: ${status.modeLabel || status.mode || "-"}`,
        `Client: ${status.clientStatus || "-"}`,
        `Server: ${status.serverStatus || "-"}`,
      ];
      if (status.resolvedAddress) {
        lines.push(`Address: ${status.resolvedAddress}`);
      }
      setOutput(backendOutput, lines.filter(Boolean).join("\n"), status.overallStatus === "error");
    }
  } catch (error) {
    resetBackendStatusCards();
    const noteNode = document.getElementById("backendStatusNote");
    if (noteNode) {
      noteNode.textContent = error.message;
    }
    if (showOutput) {
      setOutput(backendOutput, error.message, true);
    }
  }
}

function currentPreviewSeconds() {
  return Number(document.getElementById("previewMode").value || 0);
}

function resolveSyncRenderOutputPath() {
  const rawOutput = document.getElementById("outputPath").value.trim();
  const videoPath = document.getElementById("videoPath").value.trim().replace(/\//g, "\\");
  const videoBase = videoPath
    ? videoPath.replace(/.*\\/, "").replace(/\.[^.]+$/, "")
    : "camera";
  const defaultName = `${videoBase}_sync.mp4`;

  if (!rawOutput) {
    if (videoPath.includes("\\")) {
      return `${videoPath.replace(/\\[^\\]+$/, "")}\\${defaultName}`;
    }
    return defaultName;
  }

  const normalized = rawOutput.replace(/\//g, "\\");
  const looksLikeFile = /\.[^\\/.]+$/.test(normalized);
  if (looksLikeFile) {
    return normalized;
  }
  return `${normalized.replace(/\\+$/, "")}\\${defaultName}`;
}

function currentMulticamPreviewSeconds() {
  return Number(document.getElementById("multicamPreviewMode").value || 0);
}

function deriveMulticamAlignedDir() {
  const renderOutput = document.getElementById("multicamRenderOutput").value.trim();
  if (renderOutput) {
    const normalized = renderOutput.replace(/\//g, "\\");
    const looksLikeFile = /\.[^\\/.]+$/.test(normalized);
    const baseDir = looksLikeFile
      ? normalized.replace(/\\[^\\]+$/, "")
      : normalized.replace(/\\+$/, "");
    if (baseDir) {
      return `${baseDir}\\aligned`;
    }
  }
  const masterAudioPath = document.getElementById("masterAudioPath").value.trim().replace(/\//g, "\\");
  if (masterAudioPath.includes("\\")) {
    return masterAudioPath.replace(/\\[^\\]+$/, "\\aligned");
  }
  return "";
}

function resolveMulticamRenderOutputPath() {
  const rawOutput = document.getElementById("multicamRenderOutput").value.trim();
  const masterAudioPath = document.getElementById("masterAudioPath").value.trim().replace(/\//g, "\\");
  const masterBase = masterAudioPath
    ? masterAudioPath.replace(/.*\\/, "").replace(/\.[^.]+$/, "")
    : "master";
  const defaultName = `${masterBase}_multicam.mp4`;

  if (!rawOutput) {
    if (masterAudioPath.includes("\\")) {
      return `${masterAudioPath.replace(/\\[^\\]+$/, "")}\\${defaultName}`;
    }
    return defaultName;
  }

  const normalized = rawOutput.replace(/\//g, "\\");
  const looksLikeFile = /\.[^\\/.]+$/.test(normalized);
  if (looksLikeFile) {
    return normalized;
  }
  return `${normalized.replace(/\\+$/, "")}\\${defaultName}`;
}

function currentMulticamPayload() {
  return {
    masterAudioPath: document.getElementById("masterAudioPath").value.trim(),
    cameraPaths: [
      document.getElementById("camera1Path").value.trim(),
      document.getElementById("camera2Path").value.trim(),
      document.getElementById("camera3Path").value.trim(),
      document.getElementById("camera4Path").value.trim(),
    ].filter(Boolean),
    analyzeSeconds: Number(document.getElementById("analyzeSeconds").value || 180),
    maxLagSeconds: Number(document.getElementById("maxLagSeconds").value || 12),
  };
}

function multicamAnalysisSignature(payload) {
  return JSON.stringify({
    masterAudioPath: normalizeComparablePath(payload.masterAudioPath),
    cameraPaths: payload.cameraPaths.map((item) => normalizeComparablePath(item)),
    analyzeSeconds: Number(payload.analyzeSeconds || 0),
    maxLagSeconds: Number(payload.maxLagSeconds || 0),
  });
}

function reusableMulticamMeasuredCameras(payload) {
  if (!lastMulticamResult || !Array.isArray(lastMulticamResult.cameras)) {
    return [];
  }
  if (lastMulticamResult.analysisSignature !== multicamAnalysisSignature(payload)) {
    return [];
  }
  return lastMulticamResult.cameras;
}

function currentAISettings() {
  return {
    editMode: document.getElementById("editMode").value,
    assemblyAiKey: document.getElementById("assemblyAiKey").value.trim(),
    aiProvider: document.getElementById("aiProvider").value,
    aiKey: resolveSelectedAIKey(document.getElementById("aiProvider").value),
    aiPrompt: document.getElementById("shortsPrompt")?.value.trim() || "",
  };
}

function resolveSelectedAIKey(provider) {
  const normalized = String(provider || "").trim().toLowerCase();
  if (normalized === "gemini") {
    return document.getElementById("geminiAiKey")?.value.trim() || "";
  }
  if (normalized === "openai") {
    return document.getElementById("openAiKey")?.value.trim() || "";
  }
  return "";
}

function defaultShortsOutputDir() {
  const videoPath = document.getElementById("shortsVideoPath")?.value.trim() || "";
  if (!videoPath) {
    return "";
  }
  const normalized = videoPath.replace(/\//g, "\\");
  const baseName = normalized.replace(/.*\\/, "").replace(/\.[^.]+$/, "") || "interview";
  if (normalized.includes("\\")) {
    return `${normalized.replace(/\\[^\\]+$/, "")}\\${baseName}_shorts`;
  }
  return `${baseName}_shorts`;
}

function currentShortsPayload() {
  return {
    videoPath: document.getElementById("shortsVideoPath").value.trim(),
    audioPath: document.getElementById("shortsAudioPath").value.trim(),
    analyzeSeconds: 180,
    maxLagSeconds: 12,
    shortsCount: Number(document.getElementById("shortsCount").value || 3),
    assemblyAiKey: document.getElementById("assemblyAiKey").value.trim(),
    aiProvider: document.getElementById("aiProvider").value,
    aiKey: resolveSelectedAIKey(document.getElementById("aiProvider").value),
    aiPrompt: document.getElementById("shortsPrompt")?.value.trim() || "",
  };
}

function currentShortsFormats() {
  return Array.from(document.querySelectorAll(".shorts-format-checkbox:checked")).map((node) => node.value);
}

function describeShortsFormat(formatId) {
  if (formatId === "source-original") {
    return currentLanguage === "ru" ? "Исходный формат" : "Original format";
  }
  const preset = SHORTS_FORMAT_PRESETS.find((item) => item.id === formatId);
  return preset ? t(preset.labelKey) : formatId;
}

function injectShortsStyles() {
  if (document.getElementById("singleCamShortsStyles")) {
    return;
  }
  const style = document.createElement("style");
  style.id = "singleCamShortsStyles";
  style.textContent = `
    .shorts-meta-row { display:flex; justify-content:space-between; gap:12px; align-items:center; margin-bottom:10px; }
    .shorts-meta-row .mini-note { margin:0; }
    .shorts-subtitle-style-row { display:grid; grid-template-columns:minmax(220px,2fr) minmax(180px,1fr) minmax(180px,1fr); gap:8px; margin-bottom:10px; }
    .shorts-subtitle-style-row .input-group { margin-bottom:0; }
    .shorts-format-grid { display:grid; grid-template-columns:repeat(auto-fit, minmax(180px, 1fr)); gap:8px; margin-bottom:10px; }
    .shorts-format-option { display:flex; align-items:center; gap:8px; border:1px solid var(--line-2); background:var(--panel-2); border-radius:6px; padding:10px 12px; }
    .shorts-review { border:1px solid var(--line-2); border-radius:6px; background:rgba(0,0,0,0.18); padding:12px; overflow:hidden; }
    .shorts-review-title { font-size:12px; font-weight:600; text-transform:uppercase; letter-spacing:0.03em; color:#fff; margin-bottom:10px; }
    .shorts-review-empty { color:var(--muted); font-size:12px; }
    .shorts-segment-list { display:flex; flex-direction:column; gap:10px; max-height:340px; overflow-y:auto; padding-right:4px; }
    .shorts-segment-card { border:1px solid var(--line-2); border-radius:6px; background:var(--panel-2); padding:10px 12px; }
    .shorts-segment-top { display:flex; gap:10px; align-items:flex-start; }
    .shorts-segment-main { flex:1; min-width:0; }
    .shorts-segment-title { color:#fff; font-weight:600; margin-bottom:4px; }
    .shorts-segment-time { color:#9aa0ac; font-size:12px; margin-bottom:0; }
    .shorts-actions { margin-top:10px; }
    .shorts-output-note { margin-top:-2px; }
    @media (max-width: 900px) { .shorts-subtitle-style-row { grid-template-columns:1fr; } }
  `;
  document.head.appendChild(style);
}

function createNode(tag, options = {}) {
  const node = document.createElement(tag);
  if (options.id) {
    node.id = options.id;
  }
  if (options.className) {
    node.className = options.className;
  }
  if (options.text) {
    node.textContent = options.text;
  }
  if (options.type) {
    node.type = options.type;
  }
  if (options.value !== undefined) {
    node.value = options.value;
  }
  if (options.placeholder) {
    node.placeholder = options.placeholder;
  }
  if (options.dataset) {
    Object.entries(options.dataset).forEach(([key, value]) => {
      node.dataset[key] = value;
    });
  }
  return node;
}

function ensureShortsTabUI() {
  injectShortsStyles();
  const tabsBar = document.querySelector(".tabs");
  const backendTab = document.getElementById("tabBackendBtn");
  const contentArea = document.querySelector(".content-area");
  const backendView = document.getElementById("viewBackend");
  const backendAdvanced = document.querySelector("#viewBackend .advanced-toggle");
  if (!tabsBar || !backendTab || !contentArea || !backendView || !backendAdvanced) {
    return;
  }

  if (!document.getElementById("tabShortsBtn")) {
    const shortsTab = createNode("button", { id: "tabShortsBtn", className: "tab-btn", text: t("tab_shorts"), dataset: { i18n: "tab_shorts" } });
    shortsTab.type = "button";
    tabsBar.insertBefore(shortsTab, backendTab);
  }

  let shortsView = document.getElementById("viewShorts");
  if (!shortsView) {
    shortsView = createNode("section", { id: "viewShorts", className: "view-section" });
    const card = createNode("div", { className: "card" });
    card.style.borderBottom = "none";
    shortsView.appendChild(card);
    contentArea.insertBefore(shortsView, backendView);
  }
  const shortsCard = shortsView.querySelector(".card");

  if (!document.getElementById("shortsVideoPathGroup")) {
    const videoGroup = createNode("div", { id: "shortsVideoPathGroup", className: "input-group droppable" });
    const videoPrefix = createNode("span", { className: "input-prefix", text: t("label_source_video_short") });
    const videoInput = createNode("input", { id: "shortsVideoPath", type: "text", placeholder: t("placeholder_video_path") });
    const videoButton = createNode("button", { id: "pickShortsVideoBtn", className: "browse-btn", text: t("browse_btn") });
    videoButton.type = "button";
    videoGroup.append(videoPrefix, videoInput, videoButton);
    shortsCard.appendChild(videoGroup);
  }

  if (!document.getElementById("shortsAudioPathGroup")) {
    const audioGroup = createNode("div", { id: "shortsAudioPathGroup", className: "input-group droppable" });
    const audioPrefix = createNode("span", { className: "input-prefix", text: t("label_source_audio_short") });
    const audioInput = createNode("input", { id: "shortsAudioPath", type: "text", placeholder: t("placeholder_audio_path") });
    const audioButton = createNode("button", { id: "pickShortsAudioBtn", className: "browse-btn", text: t("browse_btn") });
    audioButton.type = "button";
    audioGroup.append(audioPrefix, audioInput, audioButton);
    shortsCard.appendChild(audioGroup);
  }

  const backendContent = backendAdvanced.querySelector(".advanced-content");
  const aiProviderGroup = document.getElementById("aiProvider")?.closest(".input-group");
  const countGroup = document.getElementById("shortsCount")?.closest(".input-group");
  const promptGroup = document.getElementById("aiPrompt")?.closest(".input-group");
  if (aiProviderGroup && !document.getElementById("shortsAiProviderGroup")) {
    aiProviderGroup.id = "shortsAiProviderGroup";
    shortsCard.appendChild(aiProviderGroup);
  }
  if (promptGroup && !document.getElementById("shortsPromptGroup")) {
    promptGroup.id = "shortsPromptGroup";
    const promptInput = document.getElementById("aiPrompt");
    if (promptInput) {
      promptInput.id = "shortsPrompt";
    }
    shortsCard.appendChild(promptGroup);
  }
  if (countGroup && !document.getElementById("shortsCountGroup")) {
    countGroup.id = "shortsCountGroup";
    shortsCard.appendChild(countGroup);
  }
  const planGroup = document.getElementById("planShortsBtn")?.closest(".input-group");
  if (planGroup) {
    planGroup.remove();
  }

  if (!document.getElementById("shortsCaptionsGroup")) {
    const captionsGroup = createNode("div", { id: "shortsCaptionsGroup", className: "input-group" });
    captionsGroup.append(
      createNode("span", { className: "input-prefix", text: t("label_captions_short") }),
      (() => {
        const select = createNode("select", { id: "shortsCaptionsMode", className: "custom-select" });
        select.style.border = "none";
        select.style.background = "transparent";
        const off = createNode("option", { value: "off", text: t("captions_off") });
        const burned = createNode("option", { value: "burned-in", text: t("captions_burned_in") });
        select.append(off, burned);
        return select;
      })(),
    );
    shortsCard.appendChild(captionsGroup);
  }

  if (!document.getElementById("shortsSubtitleStyleRow")) {
    shortsCard.appendChild(createNode("div", { id: "shortsSubtitleStyleRow", className: "shorts-subtitle-style-row" }));
  }

  if (!document.getElementById("shortsSubtitleFontGroup")) {
    const fontGroup = createNode("div", { id: "shortsSubtitleFontGroup", className: "input-group" });
    fontGroup.append(
      createNode("span", { className: "input-prefix", text: t("label_subtitle_font_short") }),
      (() => {
        const select = createNode("select", { id: "shortsSubtitleFont", className: "custom-select" });
        select.style.border = "none";
        select.style.background = "transparent";
        select.append(
          createNode("option", { value: "segoe-ui", text: t("subtitle_font_segoe") }),
          createNode("option", { value: "montserrat", text: t("subtitle_font_montserrat") }),
          createNode("option", { value: "arial", text: t("subtitle_font_arial") }),
          createNode("option", { value: "verdana", text: t("subtitle_font_verdana") }),
          createNode("option", { value: "tahoma", text: t("subtitle_font_tahoma") }),
          createNode("option", { value: "trebuchet-ms", text: t("subtitle_font_trebuchet") }),
          createNode("option", { value: "georgia", text: t("subtitle_font_georgia") }),
        );
        return select;
      })(),
    );
    document.getElementById("shortsSubtitleStyleRow")?.appendChild(fontGroup);
  }

  if (!document.getElementById("shortsSubtitleBgColorGroup")) {
    const bgGroup = createNode("div", { id: "shortsSubtitleBgColorGroup", className: "input-group" });
    const colorInput = createNode("input", { id: "shortsSubtitleBgColor", type: "color" });
    colorInput.style.width = "100%";
    colorInput.style.height = "100%";
    colorInput.style.minHeight = "44px";
    colorInput.style.background = "transparent";
    colorInput.style.border = "none";
    colorInput.style.padding = "6px 10px";
    bgGroup.append(
      createNode("span", { className: "input-prefix", text: t("label_subtitle_bg_short") }),
      colorInput,
    );
    document.getElementById("shortsSubtitleStyleRow")?.appendChild(bgGroup);
  }

  if (!document.getElementById("shortsSubtitleBgOpacityGroup")) {
    const opacityGroup = createNode("div", { id: "shortsSubtitleBgOpacityGroup", className: "input-group" });
    opacityGroup.append(
      createNode("span", { className: "input-prefix", text: t("label_subtitle_bg_opacity_short") }),
      (() => {
        const select = createNode("select", { id: "shortsSubtitleBgOpacity", className: "custom-select" });
        select.style.border = "none";
        select.style.background = "transparent";
        [0, 10, 25, 50, 75, 100].forEach((value) => {
          select.appendChild(createNode("option", { value: String(value), text: `${value}%` }));
        });
        return select;
      })(),
    );
    document.getElementById("shortsSubtitleStyleRow")?.appendChild(opacityGroup);
  }

  if (!document.getElementById("shortsFormatGrid")) {
    const formatGrid = createNode("div", { id: "shortsFormatGrid", className: "shorts-format-grid" });
    SHORTS_FORMAT_PRESETS.forEach((preset) => {
      const label = createNode("label", { className: "shorts-format-option" });
      const checkbox = createNode("input", { type: "checkbox", value: preset.id });
      checkbox.className = "shorts-format-checkbox";
      const text = createNode("span", { text: t(preset.labelKey) });
      label.append(checkbox, text);
      formatGrid.appendChild(label);
    });
    shortsCard.appendChild(formatGrid);
  }

  if (!document.getElementById("shortsOutputDirGroup")) {
    const outputGroup = createNode("div", { id: "shortsOutputDirGroup", className: "input-group droppable" });
    const prefix = createNode("span", { className: "input-prefix", text: t("label_shorts_output_short") });
    const input = createNode("input", { id: "shortsOutputDir", type: "text", placeholder: t("placeholder_shorts_output") });
    const button = createNode("button", { id: "pickShortsOutputDirBtn", className: "browse-btn", text: t("browse_btn") });
    button.type = "button";
    outputGroup.append(prefix, input, button);
    shortsCard.appendChild(outputGroup);

  }

  if (!document.getElementById("shortsActionRow")) {
    const actionRow = createNode("div", { id: "shortsActionRow", className: "actions-inline shorts-actions" });
    const planButton = document.getElementById("planShortsBtn") || createNode("button", { id: "planShortsBtn", className: "run-btn analyze", text: t("btn_build_plan") });
    planButton.type = "button";
    planButton.className = "run-btn analyze";
    actionRow.appendChild(planButton);

    const renderButton = createNode("button", { id: "renderShortsBtn", className: "run-btn secondary render", text: t("btn_render_selected_shorts") });
    renderButton.type = "button";
    actionRow.appendChild(renderButton);

    const fullButton = createNode("button", { id: "renderFullShortsBtn", className: "run-btn secondary render", text: t("btn_render_full_subtitled") });
    fullButton.type = "button";
    actionRow.appendChild(fullButton);

    const cancelButton = createNode("button", { id: "cancelShortsBtn", className: "run-btn secondary cancel", text: t("btn_cancel") });
    cancelButton.type = "button";
    cancelButton.disabled = true;
    actionRow.appendChild(cancelButton);
    shortsCard.appendChild(actionRow);
  }
  document.getElementById("shortsPlanProgress")?.remove();

  if (!document.getElementById("shortsReviewPanel")) {
    const panel = createNode("div", { id: "shortsReviewPanel", className: "shorts-review" });
    panel.append(
      createNode("div", { id: "shortsReviewTitle", className: "shorts-review-title", text: t("shorts_review_title") }),
      createNode("div", { id: "shortsReviewEmpty", className: "shorts-review-empty", text: t("shorts_review_empty") }),
      createNode("div", { id: "shortsSegmentList", className: "shorts-segment-list" }),
    );
    shortsCard.appendChild(panel);
  }

  if (!shortsOutput) {
    const bottomBar = document.querySelector(".bottom-bar");
    if (bottomBar) {
      shortsOutput = createNode("div", { id: "shortsOutput", className: "log-box", text: t("shorts_output_idle"), dataset: { i18n: "shorts_output_idle" } });
      bottomBar.insertBefore(shortsOutput, multicamOutput);
    }
  }

  const prefs = loadShortsPrefs();
  const hintNode = document.getElementById("shortsHint");
  if (hintNode) {
    hintNode.style.display = "none";
  }
  const sourceHintNode = document.getElementById("shortsSourceHint");
  if (sourceHintNode) {
    sourceHintNode.style.display = "none";
  }
  const formatsSummaryNode = document.getElementById("shortsFormatsGroup");
  if (formatsSummaryNode) {
    formatsSummaryNode.style.display = "none";
  }
  const outputNoteNode = document.getElementById("shortsOutputDirNote");
  if (outputNoteNode) {
    outputNoteNode.style.display = "none";
  }
  const shortsVideoPath = document.getElementById("shortsVideoPath");
  if (shortsVideoPath && !shortsVideoPath.value) {
    shortsVideoPath.value = prefs.videoPath || "";
  }
  const shortsAudioPath = document.getElementById("shortsAudioPath");
  if (shortsAudioPath && !shortsAudioPath.value) {
    shortsAudioPath.value = prefs.audioPath || "";
  }
  const shortsOutputDir = document.getElementById("shortsOutputDir");
  if (shortsOutputDir && !shortsOutputDir.value) {
    shortsOutputDir.value = prefs.outputDir || defaultShortsOutputDir();
  }
  if (document.getElementById("shortsCaptionsMode")) {
    document.getElementById("shortsCaptionsMode").value = prefs.captionsMode || "off";
  }
  if (document.getElementById("shortsSubtitleFont")) {
    document.getElementById("shortsSubtitleFont").value = prefs.subtitleFont || "segoe-ui";
  }
  if (document.getElementById("shortsSubtitleBgColor")) {
    document.getElementById("shortsSubtitleBgColor").value = prefs.subtitleBgColor || "#000000";
  }
  if (document.getElementById("shortsSubtitleBgOpacity")) {
    document.getElementById("shortsSubtitleBgOpacity").value = String(
      Number.isFinite(Number(prefs.subtitleBgOpacity)) ? Number(prefs.subtitleBgOpacity) : 50,
    );
  }
  if (document.getElementById("aiProvider")) {
    document.getElementById("aiProvider").value = prefs.aiProvider || "";
  }
  if (document.getElementById("shortsCount")) {
    document.getElementById("shortsCount").value = String(prefs.shortsCount || 3);
  }
  if (document.getElementById("shortsPrompt") && !document.getElementById("shortsPrompt").value) {
    document.getElementById("shortsPrompt").value = prefs.aiPrompt || "";
  }

  document.querySelectorAll(".shorts-format-checkbox").forEach((checkbox) => {
    checkbox.checked = prefs.formats.includes(checkbox.value);
    if (!checkbox.dataset.boundPrefs) {
      checkbox.addEventListener("change", saveShortsPrefs);
      checkbox.dataset.boundPrefs = "true";
    }
  });
  setShortsActionActive(false);

  [["shortsVideoPath", "input"], ["shortsAudioPath", "input"], ["shortsOutputDir", "input"], ["shortsOutputDir", "change"], ["shortsCaptionsMode", "change"], ["shortsSubtitleFont", "change"], ["shortsSubtitleBgColor", "input"], ["shortsSubtitleBgColor", "change"], ["shortsSubtitleBgOpacity", "change"], ["shortsCount", "input"], ["shortsPrompt", "change"], ["shortsPrompt", "input"], ["aiProvider", "change"]].forEach(([id, eventName]) => {
    const node = document.getElementById(id);
    if (node && !node.dataset[`bound${eventName}`]) {
      node.addEventListener(eventName, saveShortsPrefs);
      node.dataset[`bound${eventName}`] = "true";
    }
  });

  const videoBrowseButton = document.getElementById("pickShortsVideoBtn");
  if (videoBrowseButton && !videoBrowseButton.dataset.bound) {
    videoBrowseButton.addEventListener("click", async () => {
      try {
        const path = await pickFile("video");
        if (path) {
          document.getElementById("shortsVideoPath").value = path;
          const outputNode = document.getElementById("shortsOutputDir");
          if (outputNode && !outputNode.value.trim()) {
            outputNode.value = defaultShortsOutputDir();
          }
          saveShortsPrefs();
        }
      } catch (error) {
        setOutput(shortsOutput || syncOutput, error.message, true);
      }
    });
    videoBrowseButton.dataset.bound = "true";
  }

  const audioBrowseButton = document.getElementById("pickShortsAudioBtn");
  if (audioBrowseButton && !audioBrowseButton.dataset.bound) {
    audioBrowseButton.addEventListener("click", async () => {
      try {
        const path = await pickFile("audio");
        if (path) {
          document.getElementById("shortsAudioPath").value = path;
          saveShortsPrefs();
        }
      } catch (error) {
        setOutput(shortsOutput || syncOutput, error.message, true);
      }
    });
    audioBrowseButton.dataset.bound = "true";
  }

  const browseButton = document.getElementById("pickShortsOutputDirBtn");
  if (browseButton && !browseButton.dataset.bound) {
    browseButton.addEventListener("click", async () => {
      try {
        const path = await pickDirectory();
        if (path) {
          document.getElementById("shortsOutputDir").value = path;
          saveShortsPrefs();
        }
      } catch (error) {
        setOutput(shortsOutput || syncOutput, error.message, true);
      }
    });
    browseButton.dataset.bound = "true";
  }

  wireDropTarget(document.getElementById("shortsVideoPathGroup"));
  wireDropTarget(document.getElementById("shortsAudioPathGroup"));
  wireDropTarget(document.getElementById("shortsOutputDirGroup"));

  if (backendContent) {
    const planButton = document.getElementById("planShortsBtn");
    if (planButton) {
      planButton.textContent = t("btn_build_plan");
    }
  }
}

function renderShortsPlanReview(plan) {
  const empty = document.getElementById("shortsReviewEmpty");
  const list = document.getElementById("shortsSegmentList");
  if (!empty || !list) {
    return;
  }
  list.replaceChildren();
  const segments = Array.isArray(plan?.segments) ? plan.segments : [];
  if (segments.length === 0) {
    empty.style.display = "";
    empty.textContent = t("shorts_review_empty");
    return;
  }

  empty.style.display = "none";
  segments.forEach((segment, index) => {
    const card = createNode("div", { className: "shorts-segment-card" });
    const top = createNode("div", { className: "shorts-segment-top" });
    const checkbox = createNode("input", { type: "checkbox" });
    checkbox.checked = segment.enabled !== false;
    checkbox.addEventListener("change", () => {
      segment.enabled = checkbox.checked;
    });

    const main = createNode("div", { className: "shorts-segment-main" });
    const title = createNode("div", { className: "shorts-segment-title", text: segment.title || `Clip ${index + 1}` });
    main.append(title);
    top.append(checkbox, main);
    card.appendChild(top);
    list.appendChild(card);
  });
}

function hasShortsPlanSegments() {
  return Boolean(
    lastShortsPlan
    && Array.isArray(lastShortsPlan.segments)
    && lastShortsPlan.segments.length > 0
  );
}

function hasBuiltShortsPlan() {
  return Boolean(
    hasShortsPlanSegments()
    && Number(lastShortsPlan.timelineDuration || 0) > 0,
  );
}

function setShortsActionActive(active) {
  shortsActionActive = Boolean(active);
  const planButton = document.getElementById("planShortsBtn");
  const renderButton = document.getElementById("renderShortsBtn");
  const fullButton = document.getElementById("renderFullShortsBtn");
  const cancelButton = document.getElementById("cancelShortsBtn");
  if (planButton) {
    planButton.disabled = shortsActionActive;
  }
  if (renderButton) {
    renderButton.disabled = shortsActionActive;
  }
  if (fullButton) {
    fullButton.disabled = shortsActionActive;
  }
  if (cancelButton) {
    cancelButton.disabled = !shortsActionActive;
  }
}

function currentShortsRenderPayload(segments, formatsOverride) {
  return {
    videoPath: document.getElementById("shortsVideoPath").value.trim(),
    audioPath: document.getElementById("shortsAudioPath").value.trim(),
    outputDir: document.getElementById("shortsOutputDir")?.value.trim() || defaultShortsOutputDir(),
    segments,
    utterances: lastShortsPlan?.utterances || [],
    formats: Array.isArray(formatsOverride) ? formatsOverride : currentShortsFormats(),
    captionsMode: document.getElementById("shortsCaptionsMode")?.value || "off",
    subtitleFont: document.getElementById("shortsSubtitleFont")?.value || "segoe-ui",
    subtitleBgColor: document.getElementById("shortsSubtitleBgColor")?.value || "#000000",
    subtitleBgOpacity: Number(document.getElementById("shortsSubtitleBgOpacity")?.value ?? 50),
    syncDelaySeconds: Number(lastShortsPlan?.syncDelaySeconds || 0),
    crf: Number(document.getElementById("crf").value || 18),
    preset: document.getElementById("preset").value,
    ...currentBackendPayload(),
  };
}

function decodeDroppedURI(value) {
  if (!value) {
    return "";
  }
  if (value.startsWith("file://")) {
    try {
      return decodeURIComponent(new URL(value).pathname);
    } catch (_) {
      return value;
    }
  }
  return value;
}

function collectDroppedPaths(event) {
  const files = Array.from(event.dataTransfer?.files || []);
  const filePaths = files
    .map((file) => file.path || file.name || "")
    .filter(Boolean);
  if (filePaths.length > 0) {
    return filePaths;
  }

  const uriList = (event.dataTransfer?.getData("text/uri-list") || "")
    .split("\n")
    .map((item) => item.trim())
    .filter(Boolean)
    .map(decodeDroppedURI);
  if (uriList.length > 0) {
    return uriList;
  }

  const plainText = event.dataTransfer?.getData("text/plain") || "";
  return plainText
    .split("\n")
    .map((item) => decodeDroppedURI(item.trim()))
    .filter(Boolean);
}

function wireDropTarget(element, { multiple = false } = {}) {
  const targetField = element.matches("input, textarea, select")
    ? element
    : element.querySelector("input, textarea, select");
  if (!targetField) {
    return;
  }
  const enter = (event) => {
    event.preventDefault();
    element.classList.add("drop-active");
  };
  const leave = (event) => {
    event.preventDefault();
    element.classList.remove("drop-active");
  };
  const over = (event) => {
    event.preventDefault();
  };
  const drop = (event) => {
    event.preventDefault();
    element.classList.remove("drop-active");
    const paths = collectDroppedPaths(event);
    if (paths.length === 0) {
      return;
    }
    targetField.value = multiple ? paths.join("\n") : paths[0];
  };

  element.addEventListener("dragenter", enter);
  element.addEventListener("dragleave", leave);
  element.addEventListener("dragover", over);
  element.addEventListener("drop", drop);
}

async function loadSystem(writeOutput = true) {
  try {
    const info = await request("/api/system");
    const displayName = `${info.name} ${info.version}`;
    currentSystemDisplayName = displayName;
    updateTitleFromState();
    const chips = [
      displayName,
      `${t("system_http")}: ${info.address}`,
      `ffmpeg: ${info.ffmpegPath || t("system_ffmpeg_missing")}`,
      `ffprobe: ${info.ffprobePath || t("system_ffprobe_missing")}`,
    ];
    if (info.remoteTools?.installedVersion) {
      chips.push(`ffmpeg-over-ip: ${info.remoteTools.installedVersion}`);
    }
    if (Array.isArray(info.bundledComponents) && info.bundledComponents.length > 0) {
      chips.push(...info.bundledComponents.map((item) => `${item.name}: ${item.version}`));
    }
    const systemInfoNode = document.getElementById("systemInfo");
    if (systemInfoNode) {
      systemInfoNode.replaceChildren(...chips.map((item) => {
        const chip = document.createElement("div");
        chip.className = "chip";
        chip.textContent = item;
        return chip;
      }));
    }
    if (writeOutput) {
      setOutput(
        backendOutput,
        [
          displayName,
          `${t("system_http")}: ${info.address}`,
          `ffmpeg: ${info.ffmpegPath || t("system_ffmpeg_missing")}`,
          `ffprobe: ${info.ffprobePath || t("system_ffprobe_missing")}`,
          info.remoteTools?.installedVersion ? `ffmpeg-over-ip: ${info.remoteTools.installedVersion}` : "",
          "",
          ...((info.bundledComponents || []).map((item) => `${item.name}: ${item.version}`)),
        ].filter(Boolean).join("\n"),
        false,
      );
    }
  } catch (error) {
    currentSystemDisplayName = "";
    updateTitleFromState();
    const systemInfoNode = document.getElementById("systemInfo");
    if (systemInfoNode) {
      const chip = document.createElement("div");
      chip.className = "chip";
      chip.textContent = `${t("system_unavailable")}: ${error.message}`;
      systemInfoNode.replaceChildren(chip);
    }
    if (writeOutput) {
      setOutput(backendOutput, `${t("system_unavailable")}: ${error.message}`, true);
    }
  }
}

document.getElementById("analyzeSyncBtn").addEventListener("click", async () => {
  setOutput(syncOutput, t("status_sync_analyzing"), false);
  try {
    const result = await request("/api/analyze-sync", currentSyncPayload());
    lastDelaySeconds = result.delaySeconds;
    setOutput(
      syncOutput,
      [
        `${t("label_delay")}: ${fmtSeconds(result.delaySeconds)}`,
        `${t("label_confidence")}: ${result.confidence}`,
        `${t("label_video_duration")}: ${result.videoDuration} ${t("unit_seconds_short")}`,
        `${t("label_audio_duration")}: ${result.audioDuration} ${t("unit_seconds_short")}`,
        "",
        result.recommendation,
        result.renderSummary,
      ].join("\n"),
      false,
    );
  } catch (error) {
    setOutput(syncOutput, error.message, true);
  }
});

document.getElementById("renderSyncBtn").addEventListener("click", async () => {
  try {
    const payload = currentSyncPayload();
    const result = await request("/api/analyze-sync", payload);
    lastDelaySeconds = result.delaySeconds;
    const previewSeconds = currentPreviewSeconds();
    if (!confirmPreviewRender(previewSeconds, "sync")) {
      return;
    }

    await runStreamedRender("/api/render-sync-stream", {
      videoPath: payload.videoPath,
      audioPath: payload.audioPath,
      outputPath: document.getElementById("outputPath").value.trim(),
      previewSeconds,
      delaySeconds: lastDelaySeconds,
      crf: Number(document.getElementById("crf").value || 18),
      preset: document.getElementById("preset").value,
      ...currentBackendPayload(),
    }, syncOutput, t("status_sync_rendering"), t("label_render_complete"), (renderResult) => {
      setOutput(
        syncOutput,
        [
          t("label_render_complete"),
          `${t("label_offset_used")}: ${fmtSeconds(lastDelaySeconds)}`,
          previewSeconds > 0 ? `${t("label_preview_render")}: ${previewLabel(previewSeconds)}` : "",
          `${t("label_saved_to")}: ${renderResult.outputPath}`,
          `${t("label_elapsed")}: ${renderResult.duration}`,
          "",
          `${t("label_command")}:`,
          renderResult.command,
        ].join("\n"),
        false,
      );
    });
  } catch (error) {
    setOutput(syncOutput, error.message, true);
  }
});

document.getElementById("analyzeMulticamBtn").addEventListener("click", async () => {
  setOutput(multicamOutput, t("status_multicam_analyzing"), false);
  try {
    const payload = currentMulticamPayload();
    const result = await request("/api/analyze-multicam", payload);
    const exportResult = await request("/api/export-multicam-plan", {
      ...payload,
      measuredCameras: result.cameras,
      outputDir: deriveMulticamAlignedDir(),
      crf: Number(document.getElementById("multicamCrf").value || 18),
      preset: document.getElementById("multicamPreset").value,
      ...currentBackendPayload(),
    });
    lastMulticamResult = {
      ...result,
      exportPlan: exportResult,
      analysisSignature: multicamAnalysisSignature(payload),
    };

    const analysisLines = result.cameras.map((camera, index) => {
      return [
        `${t("label_camera")} ${index + 1}: ${camera.path}`,
        `  ${t("label_delay")}: ${fmtSeconds(camera.delaySeconds)}`,
        `  ${t("label_confidence")}: ${camera.confidence}`,
        `  ${t("label_duration")}: ${camera.duration} ${t("unit_seconds_short")}`,
        `  ${t("label_note")}: ${camera.recommendation}`,
      ].join("\n");
    });

    const exportLines = exportResult.plans.map((plan, index) => {
      return [
        `${t("label_camera")} ${index + 1}: ${plan.path}`,
        `${t("label_delay")}: ${fmtSeconds(plan.delaySeconds)}`,
        `${t("label_confidence")}: ${plan.confidence}`,
        `${t("label_output")}: ${plan.outputPath}`,
        `${t("label_strategy")}: ${plan.strategy}`,
        `${t("label_command")}:`,
        plan.command,
      ].join("\n");
    });

    setOutput(multicamOutput, [...analysisLines, exportResult.note, "", ...exportLines].join("\n\n"), false);
  } catch (error) {
    setOutput(multicamOutput, error.message, true);
  }
});

document.getElementById("renderMulticamBtn").addEventListener("click", async () => {
  const requestedOutputPath = document.getElementById("multicamRenderOutput").value.trim();
  const resolvedOutputPath = resolveMulticamRenderOutputPath();
  const previewSeconds = currentMulticamPreviewSeconds();
  try {
    const payload = currentMulticamPayload();
    if (!confirmPreviewRender(previewSeconds, "multicam")) {
      return;
    }
    const measuredCameras = reusableMulticamMeasuredCameras(payload);
    await runStreamedRender("/api/render-multicam-stream", {
      ...payload,
      measuredCameras,
      outputPath: requestedOutputPath,
      previewSeconds,
      crf: Number(document.getElementById("multicamCrf").value || 18),
      preset: document.getElementById("multicamPreset").value,
      shotWindowSeconds: 1,
      minShotSeconds: 2.5,
      primaryCamera: Number(document.getElementById("primaryCamera").value || 1),
      ...currentAISettings(),
      ...currentBackendPayload(),
    }, multicamOutput, t("status_multicam_rendering"), t("label_multicam_render_complete"), (result) => {
      lastMulticamResult = { ...(lastMulticamResult || {}), ...result };
      const totalTimelineSeconds = typeof result.totalSeconds === "number"
        ? result.totalSeconds
        : result.totalTime;

      const shotLines = result.shots.slice(0, 12).map((shot) => {
        return `${t("label_camera")} ${shot.cameraIndex}: ${fmtSeconds(shot.start)} -> ${fmtSeconds(shot.end)}`;
      });
      const moreShots = result.shots.length > 12
        ? [t("label_more_segments", { count: result.shots.length - 12 })]
        : [];

      setOutput(
        multicamOutput,
        [
          t("label_multicam_render_complete"),
          `${measuredCameras.length > 0 ? t("label_offsets_source_cached") : t("label_offsets_source_fresh")}`,
          previewSeconds > 0 ? `${t("label_preview_render")}: ${previewLabel(previewSeconds)}` : "",
          `${t("label_saved_to")}: ${result.outputPath}`,
          `${t("label_elapsed")}: ${result.duration}`,
          `${t("label_timeline_duration")}: ${totalTimelineSeconds} ${t("unit_seconds_short")}`,
          `${t("label_shots")}: ${result.shots.length}`,
          "",
          `${t("label_command")}:`,
          result.command,
          "",
          `${t("label_shot_plan_preview")}:`,
          ...shotLines,
          ...moreShots,
        ].join("\n"),
        false,
      );
    });
  } catch (error) {
    if (/network error/i.test(String(error.message || "")) && resolvedOutputPath) {
      try {
        if (await pathExists(resolvedOutputPath)) {
          setOutput(
            multicamOutput,
            [
              t("label_multicam_render_complete"),
              `${measuredCameras.length > 0 ? t("label_offsets_source_cached") : t("label_offsets_source_fresh")}`,
              previewSeconds > 0 ? `${t("label_preview_render")}: ${previewLabel(previewSeconds)}` : "",
              `${t("label_saved_to")}: ${resolvedOutputPath}`,
              previewSeconds > 0 ? `${t("label_timeline_duration")}: ${previewLabel(previewSeconds)}` : "",
            ].filter(Boolean).join("\n"),
            false,
          );
          return;
        }
      } catch (_) {
      }
    }
    setOutput(multicamOutput, error.message, true);
  }
});

async function cancelCurrentRender() {
  if (!activeRenderOutput) {
    return;
  }
  try {
    await request("/api/cancel", {});
    setOutput(activeRenderOutput, t("status_render_cancelled"), false);
  } catch (error) {
    setOutput(activeRenderOutput, error.message, true);
  }
}

document.getElementById("cancelSyncBtn").addEventListener("click", cancelCurrentRender);
document.getElementById("cancelMulticamBtn").addEventListener("click", cancelCurrentRender);
document.getElementById("cancelShortsBtn")?.addEventListener("click", async () => {
  const outputNode = shortsOutput || syncOutput;
  try {
    await request("/api/cancel", {});
    setOutput(outputNode, t("status_render_cancelled"), false);
  } catch (error) {
    setOutput(outputNode, error.message, true);
  } finally {
    setShortsActionActive(false);
  }
});

document.getElementById("pickVideoBtn").addEventListener("click", async () => {
  try {
    const path = await pickFile("video");
    if (path) {
      document.getElementById("videoPath").value = path;
    }
  } catch (error) {
    setOutput(syncOutput, error.message, true);
  }
});

document.getElementById("pickAudioBtn").addEventListener("click", async () => {
  try {
    const path = await pickFile("audio");
    if (path) {
      document.getElementById("audioPath").value = path;
      lastDelaySeconds = null;
    }
  } catch (error) {
    setOutput(syncOutput, error.message, true);
  }
});

document.getElementById("pickOutputDirBtn").addEventListener("click", async () => {
  try {
    const path = await pickSavePath("sync-output", resolveSyncRenderOutputPath());
    if (path) {
      document.getElementById("outputPath").value = path;
    }
  } catch (error) {
    setOutput(syncOutput, error.message, true);
  }
});

["videoPath", "audioPath", "analyzeSeconds", "maxLagSeconds"].forEach((id) => {
  const node = document.getElementById(id);
  if (!node) {
    return;
  }
  node.addEventListener("input", () => {
    lastDelaySeconds = null;
  });
  node.addEventListener("change", () => {
    lastDelaySeconds = null;
  });
});

document.getElementById("pickMasterAudioBtn").addEventListener("click", async () => {
  try {
    const path = await pickFile("master-audio");
    if (path) {
      document.getElementById("masterAudioPath").value = path;
    }
  } catch (error) {
    setOutput(multicamOutput, error.message, true);
  }
});

[
  ["pickCamera1Btn", "camera1Path"],
  ["pickCamera2Btn", "camera2Path"],
  ["pickCamera3Btn", "camera3Path"],
  ["pickCamera4Btn", "camera4Path"],
].forEach(([buttonId, inputId]) => {
  document.getElementById(buttonId).addEventListener("click", async () => {
    try {
      const path = await pickFile("video");
      if (path) {
        document.getElementById(inputId).value = path;
      }
    } catch (error) {
      setOutput(multicamOutput, error.message, true);
    }
  });
});

ensureShortsTabUI();

document.getElementById("planShortsBtn").addEventListener("click", async () => {
  const outputNode = shortsOutput || syncOutput;
  setOutput(outputNode, t("status_plan_shorts"), false);
  setShortsActionActive(true);
  try {
    const payload = currentShortsPayload();
    const finalizePlan = (result) => {
      result.segments = Array.isArray(result.segments)
        ? result.segments.map((segment) => ({ ...segment, enabled: segment.enabled !== false }))
        : [];
      lastShortsPlan = result;
      renderShortsPlanReview(result);
      saveShortsPrefs();
      setShortsActionActive(false);
      const renderButton = document.getElementById("renderShortsBtn");
      if (renderButton) {
        renderButton.disabled = false;
      }
      const fullButton = document.getElementById("renderFullShortsBtn");
      if (fullButton) {
        fullButton.disabled = false;
      }

      const sourceLine = result.timelineSource === "master-audio"
      ? t("shorts_plan_source_audio")
      : t("shorts_plan_source_video");
      const lines = result.segments.map((segment, index) => {
        return [
          `${index + 1}. ${segment.title}`,
          `  ${t("label_delay")}: ${fmtSeconds(segment.start)} -> ${fmtSeconds(segment.end)}`,
          `  ${t("label_duration")}: ${segment.duration} ${t("unit_seconds_short")}`,
          `  ${t("label_note")}: ${segment.reason}`,
        ].join("\n");
      });
      setOutput(outputNode, [result.note, sourceLine, "", ...lines].join("\n\n"), false);
    };

    try {
      let finalPlan = null;
      await streamRequest("/api/plan-shorts-stream", payload, (event) => {
        if (event.message) {
          setOutput(outputNode, event.message, false);
        }
        if (event.shortsPlan) {
          finalPlan = event.shortsPlan;
        }
      });
      if (!finalPlan) {
        throw new Error("Shorts plan stream finished without a result");
      }
      finalizePlan(finalPlan);
    } catch (streamError) {
      const result = await request("/api/plan-shorts", payload);
      finalizePlan(result);
    }
  } catch (error) {
    setOutput(outputNode, error.message, true);
  } finally {
    setShortsActionActive(false);
  }
});

document.getElementById("renderShortsBtn")?.addEventListener("click", async () => {
  const outputNode = shortsOutput || syncOutput;
  if (!lastShortsPlan || !Array.isArray(lastShortsPlan.segments) || lastShortsPlan.segments.length === 0) {
    setOutput(outputNode, t("shorts_review_empty"), true);
    return;
  }

  const formats = currentShortsFormats();
  const segments = lastShortsPlan.segments.filter((segment) => segment.enabled !== false);
  if (segments.length === 0) {
    setOutput(outputNode, t("shorts_select_required"), true);
    return;
  }

  const outputDir = document.getElementById("shortsOutputDir")?.value.trim() || defaultShortsOutputDir();
  if (!outputDir) {
    setOutput(outputNode, t("shorts_output_required"), true);
    return;
  }

  saveShortsPrefs();
  const payload = currentShortsRenderPayload(segments, formats);
  payload.outputDir = outputDir;

  setShortsActionActive(true);
  try {
    const handleRenderDone = (result) => {
      const fileLines = (result.files || []).slice(0, 18).map((item) => `${describeShortsFormat(item.format)}: ${item.output}`);
      const moreFiles = (result.files || []).length > 18
        ? [`... ${(result.files || []).length - 18} more files`]
        : [];
      setOutput(
        outputNode,
        [
          t("shorts_render_complete"),
          `${t("label_saved_to")}: ${outputDir}`,
          result.planPath ? `plan: ${result.planPath}` : "",
          `${t("label_elapsed")}: ${result.duration}`,
          `${t("label_shots")}: ${result.rendered || (result.files || []).length}`,
          "",
          ...fileLines,
          ...moreFiles,
          result.failed && result.failed.length > 0 ? "" : "",
          ...(result.failed && result.failed.length > 0 ? ["", "Warnings:", ...result.failed] : []),
        ].filter(Boolean).join("\n"),
        false,
      );
    };

    try {
      await runStreamedRender("/api/render-shorts-stream", payload, outputNode, t("shorts_rendering"), t("shorts_render_complete"), handleRenderDone);
    } catch (streamError) {
      const fallback = await request("/api/render-shorts", payload);
      handleRenderDone({
        files: fallback.files || [],
        failed: fallback.failed || [],
        rendered: fallback.renderedCount || 0,
        duration: fallback.duration || "",
        planPath: fallback.planPath || "",
      });
    }
  } catch (error) {
    setOutput(outputNode, error.message, true);
  } finally {
    setShortsActionActive(false);
  }
});

document.getElementById("renderFullShortsBtn")?.addEventListener("click", async () => {
  const outputNode = shortsOutput || syncOutput;
  const videoPath = document.getElementById("shortsVideoPath")?.value.trim() || "";
  if (!videoPath) {
    setOutput(outputNode, t("shorts_full_requires_video"), true);
    return;
  }
  if ((document.getElementById("shortsCaptionsMode")?.value || "off") !== "burned-in") {
    setOutput(outputNode, t("shorts_full_requires_captions"), true);
    return;
  }
  const assemblyAiKey = document.getElementById("assemblyAiKey")?.value.trim() || "";
  if (!assemblyAiKey) {
    setOutput(outputNode, t("shorts_full_requires_ai_key"), true);
    return;
  }

  const outputDir = document.getElementById("shortsOutputDir")?.value.trim() || defaultShortsOutputDir();
  if (!outputDir) {
    setOutput(outputNode, t("shorts_output_required"), true);
    return;
  }

  saveShortsPrefs();
  const payload = {
    videoPath,
    audioPath: document.getElementById("shortsAudioPath")?.value.trim() || "",
    outputDir,
    captionsMode: document.getElementById("shortsCaptionsMode")?.value || "off",
    assemblyAiKey,
    subtitleFont: document.getElementById("shortsSubtitleFont")?.value || "segoe-ui",
    subtitleBgColor: document.getElementById("shortsSubtitleBgColor")?.value || "#000000",
    subtitleBgOpacity: Number(document.getElementById("shortsSubtitleBgOpacity")?.value ?? 50),
    crf: Number(document.getElementById("crf").value || 18),
    preset: document.getElementById("preset").value,
    ...currentBackendPayload(),
  };

  setShortsActionActive(true);
  try {
    const handleRenderDone = (result) => {
      setOutput(
        outputNode,
        [
          t("shorts_full_render_complete"),
          `${t("label_saved_to")}: ${outputDir}`,
          result.outputPath ? `${t("label_output")}: ${result.outputPath}` : "",
          result.srtPath ? `${t("label_srt_file")}: ${result.srtPath}` : "",
          result.textPath ? `${t("label_transcript_text")}: ${result.textPath}` : "",
          result.assPath ? `${t("label_ass_file")}: ${result.assPath}` : "",
          result.transcriptSource ? `${t("label_transcript_source")}: ${result.transcriptSource}` : "",
          `${t("label_elapsed")}: ${result.duration}`,
        ].filter(Boolean).join("\n"),
        false,
      );
    };

    try {
      await runStreamedRender("/api/render-full-captions-stream", payload, outputNode, t("shorts_full_rendering"), t("shorts_full_render_complete"), handleRenderDone);
    } catch (streamError) {
      const fallback = await request("/api/render-full-captions", payload);
      handleRenderDone({
        outputPath: fallback.outputPath || "",
        srtPath: fallback.srtPath || "",
        textPath: fallback.textPath || "",
        assPath: fallback.assPath || "",
        duration: fallback.duration || "",
        transcriptSource: fallback.transcriptSource || "",
      });
    }
  } catch (error) {
    setOutput(outputNode, error.message, true);
  } finally {
    setShortsActionActive(false);
  }
});

document.getElementById("pickMulticamOutputDirBtn").addEventListener("click", async () => {
  try {
    const path = await pickSavePath("multicam-output", resolveMulticamRenderOutputPath());
    if (path) {
      document.getElementById("multicamRenderOutput").value = path;
    }
  } catch (error) {
    setOutput(multicamOutput, error.message, true);
  }
});

document.getElementById("checkBackendBtn").addEventListener("click", async () => {
  await checkBackendConnection(true);
});

["executionMode", "remoteAddress", "remoteSecret"].forEach((id) => {
  const node = document.getElementById(id);
  if (!node) {
    return;
  }
  node.addEventListener("input", () => {
    resetBackendStatusCards();
    if (id !== "remoteSecret") {
      saveBackendPrefs();
    }
  });
  node.addEventListener("change", () => {
    resetBackendStatusCards();
    if (id !== "remoteSecret") {
      saveBackendPrefs();
    }
  });
});

langRuBtn.addEventListener("click", () => setLanguage("ru"));
langEnBtn.addEventListener("click", () => setLanguage("en"));
tabs.forEach((tab) => {
  document.getElementById(`tab${tab}Btn`).addEventListener("click", () => switchTab(tab));
});

wireDropTarget(document.getElementById("videoPathGroup"));
wireDropTarget(document.getElementById("audioPathGroup"));
wireDropTarget(document.getElementById("masterAudioPathGroup"));
wireDropTarget(document.getElementById("multicamRenderOutputGroup"));
wireDropTarget(document.getElementById("camera1Group"));
wireDropTarget(document.getElementById("camera2Group"));
wireDropTarget(document.getElementById("camera3Group"));
wireDropTarget(document.getElementById("camera4Group"));

normalizeSyncLayout();
ensureRemoteToolsUI();
applyBackendPrefs();
setLanguage(currentLanguage);
switchTab(currentTab);
loadStoredSecrets();
loadSystem();
loadRemoteToolsStatus(false);
resetBackendStatusCards();

document.getElementById("updateRemoteToolsBtn")?.addEventListener("click", async () => {
  await updateRemoteTools(true);
});

