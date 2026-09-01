/**
 * Freyja Flow - Deck & Quick Action Controller
 * Handles remote hardware key injection, OTP actions, and clipboard syncing.
 */

document.addEventListener('DOMContentLoaded', () => {
    initDeckShortcuts();
    initClipboardActions();
});

function initDeckShortcuts() {
    const keyActions = {
        'deck-home': 'HOME',
        'deck-back': 'BACK',
        'deck-recents': 'RECENTS',
        'deck-notifications': 'NOTIFICATIONS',
        'deck-screenshot': 'SCREENSHOT',
        'deck-vol-up': 'VOLUME_UP',
        'deck-vol-down': 'VOLUME_DOWN'
    };

    Object.keys(keyActions).forEach(btnId => {
        const btn = document.getElementById(btnId);
        if (btn) {
            btn.addEventListener('click', () => {
                sendDeckKey(keyActions[btnId]);
            });
        }
    });
}

async function sendDeckKey(action) {
    showToast(`⚡ Eylem: ${action}`);
    try {
        const res = await fetch('/api/input/key', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: action })
        });
        const data = await res.json();
        if (!data.success) {
            console.warn('Tuş gönderilemedi:', data.error);
        }
    } catch (e) {
        console.error('Tuş API Hatası:', e);
    }
}

function initClipboardActions() {
    // 2FA / OTP Kopyalama
    const copyOtpBtn = document.getElementById('copy-otp-btn');
    const otpCodeEl = document.getElementById('otp-code-display');

    if (copyOtpBtn && otpCodeEl) {
        copyOtpBtn.addEventListener('click', () => {
            const code = otpCodeEl.innerText.replace(/\s+/g, '');
            if (code) {
                navigator.clipboard.writeText(code);
                showToast(`🔑 Kod panoya kopyalandı: ${code}`);
            }
        });
    }

    // Pano İçeriğini Kopyala
    const copyClipBtn = document.getElementById('copy-clipboard-btn');
    const clipContentEl = document.getElementById('clipboard-content-display');

    if (copyClipBtn && clipContentEl) {
        copyClipBtn.addEventListener('click', () => {
            const text = clipContentEl.innerText;
            if (text && !text.includes('bekleniyor')) {
                navigator.clipboard.writeText(text);
                showToast('📋 Metin panoya kopyalandı!');
            }
        });
    }

    // Manuel Pano Senkronizasyonu
    const syncClipBtn = document.getElementById('sync-clipboard-btn');
    if (syncClipBtn) {
        syncClipBtn.addEventListener('click', async () => {
            showToast('🔄 Pano senkronize ediliyor...');
            try {
                const res = await fetch('/api/clipboard/sync', { method: 'POST' });
                const data = await res.json();
                if (data.success && data.text) {
                    if (clipContentEl) clipContentEl.innerText = data.text;
                    showToast('✅ Pano eşitlendi!');
                }
            } catch (e) {
                showToast('❌ Pano eşitlenemedi');
            }
        });
    }
}
