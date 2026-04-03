const STORAGE_LANGUAGE_KEY = "autosyncrenderserver.language";
const DEFAULT_LANGUAGE = "ru";

const I18N = {
  ru: {
    document_title: "AutoSync Render Server",
    hero_title: "AutoSync Render Server",
    hero_subtitle: "Отдельный менеджер render-node для Windows. Для обычного запуска достаточно адреса и секрета доступа.",
    config_title: "Быстрый запуск",
    config_help: "Оставьте пути пустыми, если хотите использовать встроенные бинарники приложения.",
    label_server_binary: "Путь к бинарнику сервера",
    label_ffmpeg_path: "Путь к patched FFmpeg",
    label_address: "Адрес",
    label_auth_secret: "Секрет доступа",
    label_log_mode: "Режим логирования",
    label_debug: "Отладка",
    label_rewrites: "Rewrites, по одной JSON-строке на правило",
    btn_start: "Запустить сервер",
    btn_stop: "Остановить сервер",
    btn_refresh: "Обновить",
    advanced_title: "Расширенные настройки",
    server_binary_help: "Необязательно. Если поле пустое, будет использован встроенный ffmpeg-over-ip-server.exe.",
    ffmpeg_path_help: "Необязательно. Если поле пустое, будет использован встроенный ffmpeg.exe.",
    status_title: "Статус",
    card_server: "Сервер",
    card_clients: "Подключённые клиенты",
    card_jobs: "Активные задачи",
    meta_runtime: "Сессия",
    meta_config: "Конфиг",
    section_clients: "Подключённые клиенты",
    section_jobs: "Активные задачи",
    section_logs: "Логи",
    option_true: "true",
    option_false: "false",
    placeholder_server_binary: "C:\\render-node\\ffmpeg-over-ip-server.exe",
    placeholder_ffmpeg_path: "C:\\render-node\\ffmpeg.exe",
    placeholder_secret_saved: "Сохранённый секрет будет использован",
    placeholder_secret_new: "Введите секрет доступа",
    unknown_request_error: "Неизвестная ошибка запроса",
    bundle_ready: "Встроенные инструменты готовы",
    bundle_empty: "Встроенные инструменты недоступны",
    bundle_platform: "Платформа",
    server_online: "онлайн",
    server_offline: "офлайн",
    no_clients: "Подключённых клиентов пока нет.",
    no_jobs: "Активных задач пока нет.",
    waiting_logs: "Ожидание логов...",
    flash_title_info: "Статус",
    flash_title_success: "Готово",
    flash_title_error: "Ошибка",
    flash_ready: "Можно запускать сервер.",
    flash_refreshed: "Состояние сервера обновлено.",
    flash_started: "Сервер запущен.",
    flash_stopped: "Сервер остановлен.",
    runtime_online: "Время работы: {uptime}",
    runtime_offline: "Сервер не запущен",
    runtime_last_exit: "Последний выход: {lastExit}",
    config_meta: "Адрес: {address}\nФайл: {path}",
    config_meta_no_path: "Адрес: {address}",
    not_available: "недоступно",
  },
  en: {
    document_title: "AutoSync Render Server",
    hero_title: "AutoSync Render Server",
    hero_subtitle: "A standalone Windows render-node manager. For normal startup, the address and access secret are enough.",
    config_title: "Quick Start",
    config_help: "Leave paths empty to use the application's bundled binaries.",
    label_server_binary: "Server binary path",
    label_ffmpeg_path: "Patched FFmpeg path",
    label_address: "Address",
    label_auth_secret: "Auth secret",
    label_log_mode: "Log mode",
    label_debug: "Debug",
    label_rewrites: "Rewrites, one JSON line per rule",
    btn_start: "Start server",
    btn_stop: "Stop server",
    btn_refresh: "Refresh",
    advanced_title: "Advanced settings",
    server_binary_help: "Optional. If empty, the bundled ffmpeg-over-ip-server.exe will be used.",
    ffmpeg_path_help: "Optional. If empty, the bundled ffmpeg.exe will be used.",
    status_title: "Status",
    card_server: "Server",
    card_clients: "Connected Clients",
    card_jobs: "Active Jobs",
    meta_runtime: "Session",
    meta_config: "Config",
    section_clients: "Connected Clients",
    section_jobs: "Active Jobs",
    section_logs: "Logs",
    option_true: "true",
    option_false: "false",
    placeholder_server_binary: "C:\\render-node\\ffmpeg-over-ip-server.exe",
    placeholder_ffmpeg_path: "C:\\render-node\\ffmpeg.exe",
    placeholder_secret_saved: "Saved secret will be reused",
    placeholder_secret_new: "Enter access secret",
    unknown_request_error: "Unknown request error",
    bundle_ready: "Bundled tools are ready",
    bundle_empty: "Bundled tools are unavailable",
    bundle_platform: "Platform",
    server_online: "online",
    server_offline: "offline",
    no_clients: "No clients yet.",
    no_jobs: "No active jobs.",
    waiting_logs: "Waiting for logs...",
    flash_title_info: "Status",
    flash_title_success: "Done",
    flash_title_error: "Error",
    flash_ready: "Server is ready to start.",
    flash_refreshed: "Server state refreshed.",
    flash_started: "Server started.",
    flash_stopped: "Server stopped.",
    runtime_online: "Uptime: {uptime}",
    runtime_offline: "Server is not running",
    runtime_last_exit: "Last exit: {lastExit}",
    config_meta: "Address: {address}\nFile: {path}",
    config_meta_no_path: "Address: {address}",
    not_available: "not available",
  },
};

let currentLanguage = loadLanguage();
let lastStatus = null;
let authSecretDirty = false;

function loadLanguage() {
  const saved = localStorage.getItem(STORAGE_LANGUAGE_KEY);
  return saved && I18N[saved] ? saved : DEFAULT_LANGUAGE;
}

function t(key, params = {}) {
  const template = I18N[currentLanguage]?.[key] ?? I18N[DEFAULT_LANGUAGE][key] ?? key;
  return Object.entries(params).reduce(
    (acc, [name, value]) => acc.replaceAll(`{${name}}`, String(value)),
    template,
  );
}

function request(url, payload) {
  return fetch(url, {
    method: payload ? "POST" : "GET",
    headers: payload ? { "Content-Type": "application/json" } : undefined,
    body: payload ? JSON.stringify(payload) : undefined,
  }).then(async (response) => {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || t("unknown_request_error"));
    }
    return data;
  });
}

function setLanguage(language) {
  if (!I18N[language]) {
    return;
  }
  currentLanguage = language;
  localStorage.setItem(STORAGE_LANGUAGE_KEY, language);
  document.documentElement.lang = language;
  document.title = t("document_title");

  document.querySelectorAll("[data-i18n]").forEach((node) => {
    node.textContent = t(node.dataset.i18n);
  });

  document.querySelectorAll("[data-i18n-placeholder]").forEach((node) => {
    node.setAttribute("placeholder", t(node.dataset.i18nPlaceholder));
  });

  document.querySelectorAll("[data-i18n-option]").forEach((node) => {
    node.textContent = t(node.dataset.i18nOption);
  });

  document.getElementById("langRuBtn").classList.toggle("active", language === "ru");
  document.getElementById("langEnBtn").classList.toggle("active", language === "en");

  if (lastStatus) {
    applyStatus(lastStatus);
  } else {
    setFlash("info", "flash_title_info", "flash_ready");
  }
}

function currentConfig() {
  return {
    serverBinary: document.getElementById("serverBinary").value.trim(),
    ffmpegPath: document.getElementById("ffmpegPath").value.trim(),
    address: document.getElementById("serverAddress").value.trim(),
    authSecret: authSecretDirty ? document.getElementById("authSecret").value.trim() : "",
    logMode: document.getElementById("logMode").value,
    debug: document.getElementById("debugFlag").value === "true",
    rewrites: document.getElementById("rewrites").value.split("\n").map((item) => item.trim()).filter(Boolean),
  };
}

function setFlash(kind, titleKey, bodyKeyOrText) {
  const flash = document.getElementById("flashMessage");
  flash.className = `flash ${kind} visible`;
  document.getElementById("flashTitle").textContent = t(titleKey);
  document.getElementById("flashBody").textContent = I18N[currentLanguage]?.[bodyKeyOrText]
    ? t(bodyKeyOrText)
    : bodyKeyOrText;
}

function renderBundleInfo(status) {
  const container = document.getElementById("bundleInfo");
  container.replaceChildren();
  const bundledComponents = Array.isArray(status?.bundledComponents) ? status.bundledComponents : [];

  const platformChip = document.createElement("div");
  platformChip.className = "bundle-chip";
  const platformLabel = document.createElement("strong");
  platformLabel.textContent = t("bundle_platform");
  platformChip.append(platformLabel, document.createTextNode(` ${status.bundledPlatform || t("not_available")}`));
  container.appendChild(platformChip);

  if (bundledComponents.length === 0) {
    const emptyChip = document.createElement("div");
    emptyChip.className = "bundle-chip";
    emptyChip.textContent = t("bundle_empty");
    container.appendChild(emptyChip);
    return;
  }

  const readyChip = document.createElement("div");
  readyChip.className = "bundle-chip";
  readyChip.textContent = t("bundle_ready");
  container.appendChild(readyChip);

  bundledComponents.slice(0, 4).forEach((item) => {
    const chip = document.createElement("div");
    chip.className = "bundle-chip";
    const title = document.createElement("strong");
    title.textContent = item.name;
    chip.append(title, document.createTextNode(item.version ? ` ${item.version}` : ""));
    container.appendChild(chip);
  });
}

function runtimeMeta(status) {
  if (status.running) {
    return t("runtime_online", { uptime: status.uptime || "0s" });
  }
  if (status.lastExit) {
    return `${t("runtime_offline")}\n${t("runtime_last_exit", { lastExit: status.lastExit })}`;
  }
  return t("runtime_offline");
}

function configMeta(status) {
  const address = status.serverConfig?.address || "0.0.0.0:5050";
  const path = status.serverConfigPath || "";
  if (!path) {
    return t("config_meta_no_path", { address });
  }
  return t("config_meta", { address, path });
}

function updateSecretPlaceholder(hasAuthSecret) {
  const secretInput = document.getElementById("authSecret");
  if (authSecretDirty && secretInput.value.trim() !== "") {
    return;
  }
  secretInput.placeholder = hasAuthSecret ? t("placeholder_secret_saved") : t("placeholder_secret_new");
}

function syncButtons(status) {
  document.getElementById("startBtn").disabled = Boolean(status.running);
  document.getElementById("stopBtn").disabled = !status.running;
}

function applyStatus(status) {
  lastStatus = status;
  const connectedClients = Array.isArray(status?.connectedClients) ? status.connectedClients : [];
  const activeJobs = Array.isArray(status?.activeJobs) ? status.activeJobs : [];
  const logTail = Array.isArray(status?.logTail) ? status.logTail : [];
  renderBundleInfo(status);

  document.getElementById("serverState").textContent = status.running
    ? `${t("server_online")} (PID ${status.pid})`
    : t("server_offline");
  document.getElementById("serverState").className = `stat-value ${status.running ? "ok" : "warn"}`;
  document.getElementById("clientCount").textContent = String(connectedClients.length);
  document.getElementById("jobCount").textContent = String(activeJobs.length);
  document.getElementById("serverMeta").textContent = runtimeMeta(status);
  document.getElementById("configMeta").textContent = configMeta(status);

  document.getElementById("clientsBox").textContent =
    connectedClients.length === 0
      ? t("no_clients")
      : connectedClients.map((client) => `${client.remoteAddress} (${client.state})`).join("\n");

  document.getElementById("jobsBox").textContent =
    activeJobs.length === 0
      ? t("no_jobs")
      : activeJobs.map((job) => `${job.pid} | ${job.name}\n${job.commandLine || ""}`).join("\n\n");

  document.getElementById("logsBox").textContent = logTail.join("\n") || t("waiting_logs");

  const cfg = status.serverConfig || {};
  document.getElementById("serverBinary").value = cfg.serverBinary || "";
  document.getElementById("ffmpegPath").value = cfg.ffmpegPath || "";
  document.getElementById("serverAddress").value = cfg.address || "0.0.0.0:5050";
  document.getElementById("logMode").value = cfg.logMode || "stdout";
  document.getElementById("debugFlag").value = String(cfg.debug ?? true);
  document.getElementById("rewrites").value = (cfg.rewrites || []).join("\n");

  if (!authSecretDirty) {
    document.getElementById("authSecret").value = "";
  }
  updateSecretPlaceholder(Boolean(status.hasAuthSecret));
  syncButtons(status);
}

async function refreshStatus(showFlash = false) {
  try {
    const status = await request("/api/status");
    applyStatus(status);
    if (showFlash) {
      setFlash("info", "flash_title_info", "flash_refreshed");
    }
  } catch (error) {
    setFlash("error", "flash_title_error", error.message);
  }
}

document.getElementById("startBtn").addEventListener("click", async () => {
  try {
    const status = await request("/api/start", currentConfig());
    authSecretDirty = false;
    applyStatus(status);
    setFlash("success", "flash_title_success", "flash_started");
  } catch (error) {
    setFlash("error", "flash_title_error", error.message);
  }
});

document.getElementById("stopBtn").addEventListener("click", async () => {
  try {
    const status = await request("/api/stop", {});
    applyStatus(status);
    setFlash("info", "flash_title_info", "flash_stopped");
  } catch (error) {
    setFlash("error", "flash_title_error", error.message);
  }
});

document.getElementById("refreshBtn").addEventListener("click", () => refreshStatus(true));
document.getElementById("langRuBtn").addEventListener("click", () => setLanguage("ru"));
document.getElementById("langEnBtn").addEventListener("click", () => setLanguage("en"));
document.getElementById("authSecret").addEventListener("input", () => {
  authSecretDirty = true;
});

setLanguage(currentLanguage);
refreshStatus();
setInterval(() => {
  refreshStatus(false);
}, 3000);
