/**
 * Freyja Flow - Main Frontend Application Logic
 * Manages WebSocket, device lifecycle, themes, and mirroring engine.
 */

// Durum Değişkenleri
let currentTheme = localStorage.getItem('freyja_flow_theme') || 'dark';
let activeDevice = null;
let socket = null;

document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    initQRCode();
    initWebSocket();
    initEventListeners();
    fetchDevices();

    // 10 saniyede bir cihaz durumunu tazele
    setInterval(fetchDevices, 10000);
});

/* ==================== TEMA YÖNETİMİ (YTMusic Uyumlu) ==================== */
function initTheme() {
    applyTheme(currentTheme);

    const themeBtn = document.getElementById('theme-toggle-btn');
    if (themeBtn) {
        themeBtn.addEventListener('click', () => {
            currentTheme = currentTheme === 'dark' ? 'light' : 'dark';
            localStorage.setItem('freyja_flow_theme', currentTheme);
            applyTheme(currentTheme);
        });
    }
}

function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    const logoImg = document.getElementById('brand-logo');
    const themeIcon = document.getElementById('theme-icon');

    if (logoImg) {
        logoImg.src = theme === 'light' ? 'assets/freyja_logo_light.svg' : 'assets/freyja_logo_dark.svg';
    }

    if (themeIcon) {
        themeIcon.setAttribute('data-lucide', theme === 'light' ? 'moon' : 'sun');
        if (window.lucide) lucide.createIcons();
    }
}

/* ==================== DİNAMİK QR KOD EŞLEŞTİRME ==================== */
function initQRCode() {
    const qrContainer = document.getElementById('qrcode-target');
    if (!qrContainer || !window.QRCode) return;

    // Mevcut sunucu adresini al veya varsayılanı kullan
    const host = window.location.host || '192.168.1.45:8080';
    const serverUrl = `http://${host}`;
    
    const displaySpan = document.getElementById('server-address-display');
    if (displaySpan) displaySpan.innerText = serverUrl;

    qrContainer.innerHTML = '';
    new QRCode(qrContainer, {
        text: serverUrl,
        width: 165,
        height: 165,
        colorDark: "#090614",
        colorLight: "#ffffff"
    });
}

/* ==================== WEBSOCKET BAĞLANTISI ==================== */
function initWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host || 'localhost:8080'}/ws`;

    try {
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            console.log('✅ [Freyja Flow] WebSocket Köprüsü Bağlandı.');
            updateConnectionPill(true, 'Sistem Aktif (Canlı)');
        };

        socket.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                handleSocketMessage(msg);
            } catch (err) {
                console.error('Socket mesajı ayrıştırılamadı:', err);
            }
        };

        socket.onclose = () => {
            console.warn('⚠️ [Freyja Flow] WebSocket Bağlantısı koptu. 3s sonra yeniden denenecek...');
            updateConnectionPill(false, 'Bağlantı Bekleniyor...');
            setTimeout(initWebSocket, 3000);
        };

        socket.onerror = (err) => {
            console.error('WebSocket Hatası:', err);
        };
    } catch (e) {
        console.log('WebSocket başlatılamadı (Henüz backend açık olmayabilir):', e);
    }
}

function handleSocketMessage(msg) {
    switch (msg.event) {
        case 'DEVICE_CONNECTED':
        case 'DEVICE_STATE':
            renderDevice(msg.data);
            showToast(`📱 ${msg.data.model || 'Cihaz'} bağlandı!`);
            break;
        case 'DEVICE_DISCONNECTED':
            renderDisconnectedState();
            showToast('⚠️ Cihaz bağlantısı kesildi.');
            break;
        case 'BATTERY_UPDATE':
            updateBatteryUI(msg.data.level, msg.data.charging);
            break;
        case 'CLIPBOARD_UPDATE':
            updateClipboardUI(msg.data.text);
            break;
        case 'OTP_DETECTED':
            updateOTPUI(msg.data.code, msg.data.sender);
            showToast(`🔑 Doğrulama Kodu: ${msg.data.code}`);
            break;
        default:
            console.log('Alınan olay:', msg);
    }
}

/* ==================== CİHAZ BİLGİSİ ÇEKME & ARAYÜZ ==================== */
async function fetchDevices() {
    try {
        const res = await fetch('/api/devices');
        if (!res.ok) throw new Error('API Hatası');
        const data = await res.json();
        
        if (data && data.devices && data.devices.length > 0) {
            activeDevice = data.devices[0];
            renderDevice(activeDevice);
            updateConnectionPill(true, `${activeDevice.model || 'Android Cihaz'} Bağlı`);
        } else {
            renderDisconnectedState();
            updateConnectionPill(false, 'Cihaz Bekleniyor...');
        }
    } catch (err) {
        // Backend henüz açık değilse örnek verilerle arayüzü diri tut
        console.log('API henüz canlı değil, beklemede...');
    }
}

function renderDevice(dev) {
    if (!dev) return;
    const nameEl = document.getElementById('device-name');
    const serialEl = document.getElementById('device-serial');
    const transportBadge = document.getElementById('transport-badge');
    const modelBadge = document.getElementById('device-model-badge');

    if (nameEl) nameEl.innerText = dev.model || 'Android Cihaz';
    if (serialEl) serialEl.innerText = `${dev.product || ''} (${dev.serial || 'USB'})`;
    if (modelBadge) modelBadge.innerText = dev.android_version || 'Android 14';

    if (transportBadge) {
        const isWifi = dev.serial && dev.serial.includes(':');
        transportBadge.className = isWifi ? 'tag-badge wifi' : 'tag-badge usb';
        transportBadge.innerHTML = isWifi 
            ? '<i data-lucide="wifi"></i> Wi-Fi 5GHz' 
            : '<i data-lucide="zap"></i> USB 3.0';
    }

    if (dev.battery) {
        updateBatteryUI(dev.battery.level, dev.battery.charging);
    }

    if (dev.storage) {
        const storageText = document.getElementById('storage-text-display');
        const storageBar = document.getElementById('storage-progress-bar');
        if (storageText) storageText.innerText = `${dev.storage.used_gb} GB / ${dev.storage.total_gb} GB`;
        if (storageBar) storageBar.style.width = `${(dev.storage.used_gb / dev.storage.total_gb) * 100}%`;
    }

    if (window.lucide) lucide.createIcons();
}

function renderDisconnectedState() {
    const nameEl = document.getElementById('device-name');
    const serialEl = document.getElementById('device-serial');
    if (nameEl) nameEl.innerText = 'Cihaz Algılanmadı';
    if (serialEl) serialEl.innerText = 'USB takın veya Wi-Fi ile bağlanın';
}

function updateBatteryUI(level, isCharging) {
    const textEl = document.getElementById('battery-pct-text');
    const barEl = document.getElementById('battery-progress-bar');
    if (textEl) textEl.innerText = `${isCharging ? '⚡ ' : ''}%${level} ${isCharging ? '(Şarj Oluyor)' : ''}`;
    if (barEl) barEl.style.width = `${level}%`;
}

function updateClipboardUI(text) {
    const el = document.getElementById('clipboard-content-display');
    if (el && text) {
        el.innerText = text;
        showToast('📋 Pano senkronize edildi');
    }
}

function updateOTPUI(code, sender) {
    const el = document.getElementById('otp-code-display');
    if (el) el.innerText = code;
}

function updateConnectionPill(isConnected, text) {
    const pill = document.getElementById('connection-status-pill');
    const textEl = document.getElementById('connection-status-text');
    if (pill) {
        if (isConnected) pill.classList.add('connected');
        else pill.classList.remove('connected');
    }
    if (textEl) textEl.innerText = text;
}

/* ==================== BUTON ETKİLEŞİMLERİ & SCRCPY TETİKLEME ==================== */
function initEventListeners() {
    // Manuel IP ile Wi-Fi ADB Bağlantısı
    const connectIpBtn = document.getElementById('connect-ip-btn');
    const ipInput = document.getElementById('manual-ip-input');
    if (connectIpBtn && ipInput) {
        connectIpBtn.addEventListener('click', async () => {
            const ip = ipInput.value.trim();
            if (!ip) {
                showToast('Lütfen geçerli bir IP girin (örn: 192.168.1.50:5555)');
                return;
            }
            showToast(`🌐 ${ip} adresine bağlanılıyor...`);
            try {
                const res = await fetch('/api/connect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ip })
                });
                const data = await res.json();
                if (data.success) {
                    showToast('✅ Kablosuz bağlantı başarılı!');
                    fetchDevices();
                } else {
                    showToast('❌ Bağlantı başarısız: ' + (data.error || 'Bilinmiyor'));
                }
            } catch (err) {
                showToast('❌ Sunucuya erişilemedi');
            }
        });
    }

    // USB Cihazları Yeniden Tara
    const scanUsbBtn = document.getElementById('scan-usb-btn');
    const refreshBtn = document.getElementById('refresh-devices-btn');
    const doScan = () => {
        showToast('🔍 ADB cihazları taranıyor...');
        fetchDevices();
    };
    if (scanUsbBtn) scanUsbBtn.addEventListener('click', doScan);
    if (refreshBtn) refreshBtn.addEventListener('click', doScan);

    // Scrcpy Yansıtma Butonları
    setupMirrorButton('btn-mirror-normal', 'normal');
    setupMirrorButton('btn-mirror-screenoff', 'screenoff');
    setupMirrorButton('btn-mirror-audio', 'audio');
    setupMirrorButton('btn-mirror-cam', 'camera');

    // Telefondan Dosya Çekme
    initPhoneFileExplorer();
}

function initPhoneFileExplorer() {
    const refreshFilesBtn = document.getElementById('btn-refresh-phone-files');
    const folderSelect = document.getElementById('phone-source-folder');
    const quickPhotoBtn = document.getElementById('btn-quick-photo');

    if (refreshFilesBtn && folderSelect) {
        refreshFilesBtn.addEventListener('click', () => loadPhoneFiles(folderSelect.value));
        folderSelect.addEventListener('change', () => loadPhoneFiles(folderSelect.value));
    }

    if (quickPhotoBtn) {
        quickPhotoBtn.addEventListener('click', async () => {
            showToast('📷 Telefondaki son fotoğraf çekiliyor...');
            try {
                const serial = activeDevice ? activeDevice.serial : '';
                const res = await fetch(`/api/phone/quick-photo?serial=${serial}`, { method: 'POST' });
                const data = await res.json();
                if (data.success) {
                    showToast(`✅ '${data.filename}' İndirilenler klasörünüze kaydedildi!`);
                } else {
                    showToast('❌ Fotoğraf alınamadı: ' + (data.error || 'Bilinmiyor'));
                }
            } catch (err) {
                showToast('❌ Sunucu hatası');
            }
        });
    }
}

async function loadPhoneFiles(dir) {
    const listEl = document.getElementById('phone-file-list');
    if (!listEl) return;

    listEl.innerHTML = '<div class="file-empty-state"><p>Dosyalar okunuyor...</p></div>';
    try {
        const serial = activeDevice ? activeDevice.serial : '';
        const res = await fetch(`/api/phone/files?dir=${encodeURIComponent(dir || '/sdcard/Download')}&serial=${serial}`);
        const data = await res.json();

        if (data.success && data.files && data.files.length > 0) {
            listEl.innerHTML = '';
            data.files.forEach(f => {
                const card = document.createElement('div');
                card.className = 'file-card-mini';
                const isMedia = f.name.match(/\.(jpg|jpeg|png|mp4|mkv|mp3)$/i);
                const iconName = f.is_dir ? 'folder' : (isMedia ? 'image' : 'file-text');

                card.innerHTML = `
                    <div class="file-card-icon"><i data-lucide="${iconName}"></i></div>
                    <div class="file-card-info">
                        <span class="file-card-name" title="${f.name}">${f.name}</span>
                        <span class="file-card-action">⬇️ PC'ye İndir</span>
                    </div>
                `;

                card.addEventListener('click', async () => {
                    showToast(`📥 '${f.name}' PC'ye aktarılıyor...`);
                    try {
                        const pullRes = await fetch('/api/phone/pull', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ remote_path: f.path, serial: serial })
                        });
                        const pullData = await pullRes.json();
                        if (pullData.success) {
                            showToast(`✅ '${f.name}' İndirilenler klasörüne kaydedildi!`);
                        } else {
                            showToast('❌ İndirme başarısız: ' + (pullData.error || 'Hata'));
                        }
                    } catch (e) {
                        showToast('❌ Aktarım hatası');
                    }
                });

                listEl.appendChild(card);
            });
            if (window.lucide) lucide.createIcons();
        } else {
            listEl.innerHTML = '<div class="file-empty-state"><p>Bu klasörde dosya bulunamadı.</p></div>';
        }
    } catch (err) {
        listEl.innerHTML = '<div class="file-empty-state"><p>Dosyalar listelenemedi.</p></div>';
    }
}

function setupMirrorButton(elementId, mode) {
    const btn = document.getElementById(elementId);
    if (!btn) return;
    btn.addEventListener('click', async () => {
        showToast(`🚀 Yansıtma Başlatılıyor (${mode.toUpperCase()})...`);
        try {
            const res = await fetch('/api/mirror', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ mode: mode, serial: activeDevice ? activeDevice.serial : '' })
            });
            const data = await res.json();
            if (data.success) {
                showToast('✅ Yansıtma penceresi açıldı!');
            } else {
                showToast('❌ Yansıtma başlatılamadı: ' + (data.error || 'Cihaz bulunamadı'));
            }
        } catch (e) {
            showToast('❌ Backend servisi yanıt vermedi.');
        }
    });
}

/* ==================== BİLDİRİM TOAST YÖNETİCİSİ ==================== */
function showToast(message) {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.innerHTML = `<i data-lucide="info" style="width: 18px; color: var(--aurora-lavender);"></i> <span>${message}</span>`;
    container.appendChild(toast);

    if (window.lucide) lucide.createIcons();

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateY(15px)';
        setTimeout(() => toast.remove(), 300);
    }, 3500);
}
