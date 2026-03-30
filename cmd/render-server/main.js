const STORAGE_LANGUAGE_KEY = "autosyncrenderserver.language";
const DEFAULT_LANGUAGE = "ru";

const I18N = {
  ru: {
    document_title: "AutoSync Render Server",
    hero_title: "AutoSync Render Server",
    hero_subtitle: "Отдельный менеджер для `ffmpeg-over-ip-server` на Windows: запускает render-node, показывает активные задачи, подключённых клиентов и логи в реальном времени.",
    config_title: "Конфигурация сервера",
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
    status_title: "Статус",
    card_server: "Сервер",
    card_clients: "Подключённые клиенты",
    card_jobs: "Активные задачи",
    section_clients: "Подключённые клиенты",
    section_jobs: "Активные задачи",
    section_logs: "Логи",
    clients_idle: "Подключённых клиентов пока нет.",
    jobs_idle: "Активных задач пока нет.",
    logs_idle: "Ожидание логов...",
    option_true: "true",
    option_false: "false",
    placeholder_server_binary: "C:\\render-node\\ffmpeg-over-ip-server.exe",
    placeholder_ffmpeg_path: "C:\\render-node\\ffmpeg.exe",
    unknown_request_error: "Неизвестная ошибка запроса",
    bundle_manifest: "Встроенные компоненты",
    bundle_manifest_empty: "Список встроенных компонентов пуст.",
    server_online: "онлайн",
    server_offline: "офлайн",
    no_clients: "Подключённых клиентов пока нет.",
    no_jobs: "Активных задач пока нет.",
    waiting_logs: "Ожидание логов...",
  },
  en: {
    document_title: "AutoSync Render Server",
    hero_title: "AutoSync Render Server",
    hero_subtitle: "A standalone Windows manager for `ffmpeg-over-ip-server`: starts the render node, shows active jobs, connected clients, and live logs.",
    config_title: "Server Config",
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
    status_title: "Status",
    card_server: "Server",
    card_clients: "Connected Clients",
    card_jobs: "Active Jobs",
    section_clients: "Connected Clients",
    section_jobs: "Active Jobs",
    section_logs: "Logs",
    clients_idle: "No clients yet.",
    jobs_idle: "No active jobs.",
    logs_idle: "Waiting for logs...",
    option_true: "true",
    option_false: "false",
    placeholder_server_binary: "C:\\render-node\\ffmpeg-over-ip-server.exe",
    placeholder_ffmpeg_path: "C:\\render-node\\ffmpeg.exe",
    unknown_request_error: "Unknown request error",
    bundle_manifest: "Bundled manifest",
    bundle_manifest_empty: "Bundled manifest is empty.",
    server_online: "online",
    server_offline: "offline",
    no_clients: "No clients yet.",
    no_jobs: "No active jobs.",
    waiting_logs: "Waiting for logs...",
  },
};

let currentLanguage = loadLanguage();

function loadLanguage() {
  const saved = localStorage.getItem(STORAGE_LANGUAGE_KEY);
  return saved && I18N[saved] ? saved : DEFAULT_LANGUAGE;
}

function t(key) {
  return I18N[currentLanguage]?.[key] ?? I18N[DEFAULT_LANGUAGE][key] ?? key;
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

function currentConfig() {
  return {
    serverBinary: document.getElementById("serverBinary").value.trim(),
    ffmpegPath: document.getElementById("ffmpegPath").value.trim(),
    address: document.getElementById("serverAddress").value.trim(),
    authSecret: document.getElementById("authSecret").value.trim(),
    logMode: document.getElementById("logMode").value,
    debug: document.getElementById("debugFlag").value === "true",
    rewrites: document.getElementById("rewrites").value.split("\n").map((item) => item.trim()).filter(Boolean),
  };
}

function applyStatus(status) {
  const bundleLines = (status.bundledComponents || []).map((item) => `${item.name}: ${item.version}`).join(" | ");
  document.getElementById("bundleInfo").textContent = bundleLines
    ? `${t("bundle_manifest")} (${status.bundledPlatform}): ${bundleLines}`
    : t("bundle_manifest_empty");
  document.getElementById("serverState").textContent = status.running
    ? `${t("server_online")} (PID ${status.pid})`
    : t("server_offline");
  document.getElementById("serverState").className = `card-value ${status.running ? "ok" : "warn"}`;
  document.getElementById("clientCount").textContent = String(status.connectedClients.length);
  document.getElementById("jobCount").textContent = String(status.activeJobs.length);

  document.getElementById("clientsBox").textContent =
    status.connectedClients.length === 0
      ? t("no_clients")
      : status.connectedClients.map((client) => `${client.remoteAddress} (${client.state})`).join("\n");

  document.getElementById("jobsBox").textContent =
    status.activeJobs.length === 0
      ? t("no_jobs")
      : status.activeJobs.map((job) => `${job.pid} | ${job.name}\n${job.commandLine || ""}`).join("\n\n");

  document.getElementById("logsBox").textContent = (status.logTail || []).join("\n") || t("waiting_logs");

  const cfg = status.serverConfig || {};
  document.getElementById("serverBinary").value = cfg.serverBinary || "";
  document.getElementById("ffmpegPath").value = cfg.ffmpegPath || "";
  document.getElementById("serverAddress").value = cfg.address || "0.0.0.0:5050";
  document.getElementById("authSecret").value = cfg.authSecret || "";
  document.getElementById("logMode").value = cfg.logMode || "stdout";
  document.getElementById("debugFlag").value = String(cfg.debug ?? true);
  document.getElementById("rewrites").value = (cfg.rewrites || []).join("\n");
}

async function refreshStatus() {
  const status = await request("/api/status");
  applyStatus(status);
}

document.getElementById("startBtn").addEventListener("click", async () => {
  try {
    const status = await request("/api/start", currentConfig());
    applyStatus(status);
  } catch (error) {
    alert(error.message);
  }
});

document.getElementById("stopBtn").addEventListener("click", async () => {
  try {
    const status = await request("/api/stop", {});
    applyStatus(status);
  } catch (error) {
    alert(error.message);
  }
});

document.getElementById("refreshBtn").addEventListener("click", refreshStatus);
document.getElementById("langRuBtn").addEventListener("click", () => setLanguage("ru"));
document.getElementById("langEnBtn").addEventListener("click", () => setLanguage("en"));

setLanguage(currentLanguage);
refreshStatus();
setInterval(refreshStatus, 3000);
