/**
 * Freyja Flow - Dropzone & File Transfer Engine
 * Handles Drag & Drop file transfers to Android device storage via ADB.
 */

document.addEventListener('DOMContentLoaded', () => {
    initDropzone();
});

function initDropzone() {
    const dropzone = document.getElementById('file-dropzone');
    const fileInput = document.getElementById('file-input-hidden');
    const browseBtn = document.getElementById('browse-files-btn');
    const folderSelect = document.getElementById('target-folder-select');

    if (!dropzone || !fileInput) return;

    // Dosya Seç butonuna basıldığında gizli input'u tetikle
    if (browseBtn) {
        browseBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            fileInput.click();
        });
    }

    dropzone.addEventListener('click', () => {
        fileInput.click();
    });

    fileInput.addEventListener('change', () => {
        if (fileInput.files.length > 0) {
            handleFileUpload(fileInput.files, folderSelect.value);
        }
    });

    // Drag & Drop Olayları
    ['dragenter', 'dragover'].forEach(eventName => {
        dropzone.addEventListener(eventName, (e) => {
            e.preventDefault();
            e.stopPropagation();
            dropzone.classList.add('drag-over');
        });
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropzone.addEventListener(eventName, (e) => {
            e.preventDefault();
            e.stopPropagation();
            dropzone.classList.remove('drag-over');
        });
    });

    dropzone.addEventListener('drop', (e) => {
        const dt = e.dataTransfer;
        const files = dt.files;
        if (files && files.length > 0) {
            handleFileUpload(files, folderSelect ? folderSelect.value : '/sdcard/Download/');
        }
    });
}

async function handleFileUpload(files, targetFolder) {
    const progressBox = document.getElementById('transfer-progress-box');
    const filenameEl = document.getElementById('transfer-filename');
    const pctEl = document.getElementById('transfer-pct');
    const fillEl = document.getElementById('transfer-fill');

    if (progressBox) progressBox.style.display = 'flex';

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        if (filenameEl) filenameEl.innerText = `${file.name} (${formatBytes(file.size)})`;
        if (pctEl) pctEl.innerText = '0%';
        if (fillEl) fillEl.style.width = '0%';

        showToast(`📤 '${file.name}' telefona gönderiliyor...`);

        const formData = new FormData();
        formData.append('file', file);
        formData.append('target', targetFolder || '/sdcard/Download/');

        try {
            // XMLHttpRequest ile gerçek yükleme ilerlemesini (Progress) takip et
            await uploadWithProgress(formData, (pct) => {
                if (pctEl) pctEl.innerText = `${pct}%`;
                if (fillEl) fillEl.style.width = `${pct}%`;
            });

            showToast(`✅ '${file.name}' başarıyla telefona aktarıldı!`);
        } catch (err) {
            console.error('Yükleme hatası:', err);
            showToast(`❌ '${file.name}' gönderilemedi: ${err.message || 'Hata'}`);
        }
    }

    setTimeout(() => {
        if (progressBox) progressBox.style.display = 'none';
    }, 2500);
}

function uploadWithProgress(formData, onProgress) {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/upload');

        xhr.upload.onprogress = (e) => {
            if (e.lengthComputable) {
                const percent = Math.round((e.loaded / e.total) * 100);
                onProgress(percent);
            }
        };

        xhr.onload = () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve(xhr.response);
            } else {
                reject(new Error(`HTTP ${xhr.status}`));
            }
        };

        xhr.onerror = () => reject(new Error('Ağ Hatası'));
        xhr.send(formData);
    });
}

function formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}
