import { FastSync, RunSync, CompressBatch, StartServer, StopServer, StartTgBot, StopTgBot, SelectVideo, SelectMultipleVideos, SelectAudio, SelectDirectory, CancelProcess } from '../wailsjs/go/main/App';

let currentTab = 'FastSync';
let tabStates = { FastSync: false, Multicam: false, AiCombine: false, Server: false, TgBot: false };
const allTabs = ['FastSync', 'Multicam', 'AiCombine', 'Server', 'TgBot'];

function updateGlobalButtons() {
    let startBtn = document.getElementById('globalStartBtn');
    let cancelBtn = document.getElementById('globalCancelBtn');
    let isRunning = tabStates[currentTab];

    startBtn.disabled = isRunning;
    cancelBtn.disabled = !isRunning;

    if (currentTab === 'Server') {
        cancelBtn.innerText = isRunning ? "ОСТАНОВИТЬ СЕРВЕР" : "ОТМЕНА";
        startBtn.innerText = "ЗАПУСТИТЬ СЕРВЕР";
        startBtn.classList.remove('tg-theme');
    } else if (currentTab === 'TgBot') {
        cancelBtn.innerText = isRunning ? "ОСТАНОВИТЬ БОТА" : "ОТМЕНА";
        startBtn.innerText = "ЗАПУСТИТЬ БОТА";
        startBtn.classList.add('tg-theme');
    } else {
        cancelBtn.innerText = "ОТМЕНА";
        startBtn.classList.remove('tg-theme');
        if (currentTab === 'FastSync') startBtn.innerText = "БЫСТРЫЙ СИНХРОН";
        if (currentTab === 'Multicam') startBtn.innerText = "НАЧАТЬ МОНТАЖ";
        if (currentTab === 'AiCombine') startBtn.innerText = "СЖАТЬ И НАРЕЗАТЬ";
    }
}

function switchTab(tabName) {
    currentTab = tabName;
    allTabs.forEach(id => {
        document.getElementById(`tab${id}Btn`).classList.remove('active');
        document.getElementById(`view${id}`).classList.remove('active');
        document.getElementById(`logBox${id}`).classList.remove('active');
    });
    
    document.getElementById(`tab${tabName}Btn`).classList.add('active');
    document.getElementById(`view${tabName}`).classList.add('active');
    document.getElementById(`logBox${tabName}`).classList.add('active');
    
    let outGroup = document.getElementById('globalOutGroup');
    outGroup.style.display = (tabName === 'Server' || tabName === 'TgBot') ? 'none' : 'flex';
    
    updateGlobalButtons();
}

allTabs.forEach(id => {
    document.getElementById(`tab${id}Btn`).addEventListener('click', () => switchTab(id));
});

document.addEventListener("DOMContentLoaded", () => {
    let savedPrompt = localStorage.getItem("ai_prompt"); if (savedPrompt) document.getElementById('aiPrompt').value = savedPrompt;
    let savedTgToken = localStorage.getItem("tg_token"); if (savedTgToken) document.getElementById('tgToken').value = savedTgToken;
    let savedTgAiKey = localStorage.getItem("tg_aikey"); if (savedTgAiKey) document.getElementById('tgAiKey').value = savedTgAiKey;
    let savedAiKey = localStorage.getItem("api_key"); if (savedAiKey) document.getElementById('apiKey').value = savedAiKey;
    let savedGemini = localStorage.getItem("gemini_key"); if (savedGemini) document.getElementById('geminiKey').value = savedGemini;
    let hostPort = localStorage.getItem("host_port"); if (hostPort) document.getElementById('hostPort').value = hostPort;
    let hostSecret = localStorage.getItem("host_secret"); if (hostSecret) document.getElementById('hostSecret').value = hostSecret;
});

function autoSetGlobalOutDir(filePath) {
    let outInput = document.getElementById('globalOut');
    if (!outInput.value && filePath) {
        let lastSep = Math.max(filePath.lastIndexOf("\\"), filePath.lastIndexOf("/"));
        if (lastSep > 0) outInput.value = filePath.substring(0, lastSep);
    }
}

document.getElementById('btnFsVideo').addEventListener('click', () => { SelectVideo().then(p => { if (p) { document.getElementById('fsVideoPath').value = p; autoSetGlobalOutDir(p); } }); });
document.getElementById('btnFsAudio').addEventListener('click', () => { SelectAudio().then(p => { if (p) document.getElementById('fsAudioPath').value = p; }); });

document.getElementById('btnV1').addEventListener('click', () => { SelectVideo().then(p => { if (p) { document.getElementById('vPath').value = p; autoSetGlobalOutDir(p); } }); });
document.getElementById('btnV2').addEventListener('click', () => { SelectVideo().then(p => { if (p) document.getElementById('vPath2').value = p; }); });
document.getElementById('btnVWide').addEventListener('click', () => { SelectVideo().then(p => { if (p) document.getElementById('vPathWide').value = p; }); });
document.getElementById('btnV3').addEventListener('click', () => { SelectVideo().then(p => { if (p) document.getElementById('vPath3').value = p; }); });
document.getElementById('btnV4').addEventListener('click', () => { SelectVideo().then(p => { if (p) document.getElementById('vPath4').value = p; }); });
document.getElementById('btnA').addEventListener('click', () => { SelectAudio().then(p => { if (p) document.getElementById('aPath').value = p; }); });

let selectedCompressFiles = [];
document.getElementById('btnC').addEventListener('click', () => { 
    SelectMultipleVideos().then(paths => { if (paths && paths.length > 0) { selectedCompressFiles = paths; document.getElementById('cPath').value = `Выбрано: ${paths.length}`; autoSetGlobalOutDir(paths[0]); } }); 
});

document.getElementById('btnGlobalOut').addEventListener('click', () => { SelectDirectory().then(p => { if (p) document.getElementById('globalOut').value = p; }); });

let crfSlider = document.getElementById('crfSlider');
let crfLabel = document.getElementById('crfLabel');
if(crfSlider) crfSlider.addEventListener('input', () => { crfLabel.innerText = `${crfSlider.value}`; });

let mcCrfSlider = document.getElementById('mcCrfSlider');
let mcCrfLabel = document.getElementById('mcCrfLabel');
if(mcCrfSlider) mcCrfSlider.addEventListener('input', () => { mcCrfLabel.innerText = `${mcCrfSlider.value}`; });

if (window.runtime) {
    const handleProgress = (data) => {
        let box = document.getElementById(`logBox${currentTab}`);
        if (box) box.innerText = `[${data.percent}%] ${data.message}`;
    };
    window.runtime.EventsOn("fastsync_progress", handleProgress);
    window.runtime.EventsOn("sync_progress", handleProgress);
    window.runtime.EventsOn("compress_progress", handleProgress);
    window.runtime.EventsOn("server_log", data => { let box = document.getElementById('logBoxServer'); box.innerText += data.message + "\n"; box.scrollTop = box.scrollHeight; });
    window.runtime.EventsOn("tg_log", data => { let box = document.getElementById('logBoxTgBot'); box.innerText += data.message + "\n"; box.scrollTop = box.scrollHeight; });
}

function setTabProcessing(isProcessing) {
    tabStates[currentTab] = isProcessing;
    updateGlobalButtons();
}

document.getElementById('globalCancelBtn').addEventListener('click', () => { 
    if (currentTab === 'Server') { StopServer().then(() => { setTabProcessing(false); }); } 
    else if (currentTab === 'TgBot') { StopTgBot().then(() => { document.getElementById('logBoxTgBot').innerText += "\n🛑 Бот остановлен.\n"; setTabProcessing(false); }); } 
    else { CancelProcess(); setTabProcessing(false); }
});

// 🔥 УМНАЯ НАРЕЗКА С ОТПРАВКОЙ ПУТИ К КАМЕРЕ 1
window.splitWideCam = function() {
    let v1Path = document.getElementById('vPath').value;
    if (!v1Path) { alert("⚠️ Сначала выберите Камеру 1 (эталон размера)!"); return; }
    
    let widePath = document.getElementById('vPathWide').value;
    if (!widePath) { alert("⚠️ Выберите файл Общего плана в поле над кнопкой!"); return; }
    
    let outDir = document.getElementById('globalOut').value;
    if(!outDir) { alert("⚠️ Выберите папку сохранения (OUT) в самом низу!"); return; }
    
    let useGPU = document.getElementById('mcUseGPU') ? document.getElementById('mcUseGPU').checked : false;
    let logBox = document.getElementById('logBoxMulticam');
    
    setTabProcessing(true); 
    if(logBox) logBox.innerText = `✂️ Запуск: Режем Общий план...\nКлонируем формат с Камеры 1...\nРежим: ${useGPU ? 'Ускорение видеокартой (GPU) 🚀' : 'Процессор (CPU)'}`;
    
    window.go.main.App.SplitWideCamera(widePath, v1Path, outDir, useGPU).then(res => {
        if(logBox) logBox.innerText = res;
        setTabProcessing(false); 
    }).catch(err => {
        if(logBox) logBox.innerText = "❌ Ошибка: " + err;
        setTabProcessing(false); 
    });
};

document.getElementById('globalStartBtn').addEventListener('click', () => {
    let outDir = document.getElementById('globalOut').value;
    
    if (currentTab === 'FastSync') {
        let logBox = document.getElementById('logBoxFastSync');
        let v = document.getElementById('fsVideoPath').value;
        let a = document.getElementById('fsAudioPath').value;
        
        if (!v || !a) { logBox.innerText = "⚠️ Выберите Видео и Аудио!"; return; }
        if (!outDir) { logBox.innerText = "⚠️ Выберите папку для сохранения!"; return; }
        
        setTabProcessing(true);
        logBox.innerText = "🚀 Запуск Быстрого Синхрона...";
        
        FastSync(v, a, outDir)
            .then(res => { logBox.innerText = res; setTabProcessing(false); })
            .catch(err => { logBox.innerText = "❌ Ошибка: " + err; setTabProcessing(false); });
            
    } else if (currentTab === 'Multicam') {
        let logBox = document.getElementById('logBoxMulticam');
        
        let v1 = document.getElementById('vPath').value;
        let v2 = document.getElementById('vPath2').value;
        let v3 = document.getElementById('vPath3').value; 
        let v4 = document.getElementById('vPath4').value; 
        let a = document.getElementById('aPath').value;
        
        let apiKey = document.getElementById('apiKey').value.trim();
        let mainCam = document.querySelector('input[name="mainCam"]:checked').value;

        if (!v1 || !a) { logBox.innerText = "⚠️ Выберите как минимум Камеру 1 и Мастер-аудио!"; return; }
        if (!outDir) { logBox.innerText = "⚠️ Выберите папку для сохранения!"; return; }

        setTabProcessing(true);
        logBox.innerText = "🚀 Начинаем монтаж камер...\n";

        let slider = document.getElementById('mcCrfSlider');
        let crf = slider ? parseInt(slider.value) : 23; 
        let testDuration = document.getElementById('mcTestDuration') ? parseInt(document.getElementById('mcTestDuration').value) : 0;

        window.go.main.App.RunSync(v1, a, v2, v3, v4, apiKey, parseInt(mainCam), crf, outDir, testDuration)
            .then(res => { logBox.innerText = res; setTabProcessing(false); })
            .catch(err => { logBox.innerText = "❌ Ошибка: " + err; setTabProcessing(false); });

    } else if (currentTab === 'AiCombine') {
        let logBox = document.getElementById('logBoxAiCombine');
        let geminiKey = document.getElementById('geminiKey').value.trim();
        let aiPrompt = document.getElementById('aiPrompt').value.trim();
        
        if (selectedCompressFiles.length === 0) { logBox.innerText = "⚠️ Выберите готовый подкаст!"; return; }
        if (!outDir) { logBox.innerText = "⚠️ Выберите папку для сохранения!"; return; }
        
        localStorage.setItem("gemini_key", geminiKey);
        localStorage.setItem("ai_prompt", aiPrompt);
        
        setTabProcessing(true);
        logBox.innerText = "🚀 Начинаем сжатие...";
        
        CompressBatch(selectedCompressFiles, parseInt(document.getElementById('crfSlider').value), document.getElementById('cResolution').value, false, false, document.getElementById('cUseGPU').checked, false, "", "", outDir)
            .then(res => { logBox.innerText = res; setTabProcessing(false); })
            .catch(err => { logBox.innerText = "❌ Ошибка: " + err; setTabProcessing(false); });
            
    } else if (currentTab === 'Server') {
        let logBox = document.getElementById('logBoxServer');
        setTabProcessing(true);
        logBox.innerText = "Запуск сервера...\n";
        StartServer(document.getElementById('hostPort').value.trim(), document.getElementById('hostSecret').value.trim())
            .then(res => { logBox.innerText += res + "\n"; })
            .catch(err => { logBox.innerText += "❌ Ошибка: " + err + "\n"; setTabProcessing(false); });
            
    } else if (currentTab === 'TgBot') {
        let logBox = document.getElementById('logBoxTgBot');
        let token = document.getElementById('tgToken').value.trim();
        let tgAiKey = document.getElementById('tgAiKey').value.trim();
        
        if (!token) { logBox.innerText = "⚠️ Введите токен бота!"; return; }
        localStorage.setItem("tg_token", token);
        localStorage.setItem("tg_aikey", tgAiKey);
        
        setTabProcessing(true);
        logBox.innerText = "Подключение к Telegram...\n";
        
        StartTgBot(token, tgAiKey)
            .then(res => { logBox.innerText += res + "\n"; })
            .catch(err => { logBox.innerText += "❌ Ошибка: " + err + "\n"; setTabProcessing(false); });
    }
});

switchTab('FastSync');