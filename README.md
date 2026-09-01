# 🔮 Freyja Flow

<div align="center">
  <img src="frontend/assets/freyja_logo_dark.svg" width="120" alt="Freyja Flow Logo"/>
  <h3>Android Cihaz & Ekosistem Köprüsü (Freyja OS Ecosystem)</h3>
  <p>Ultra hafif, sıfır gecikmeli ekran yansıtma, çift yönlü sürükle-bırak dosya transferi, evrensel pano eşitleme ve medya kontrol merkezi.</p>
</div>

---

## 🌟 Öne Çıkan Özellikler

- 📱 **Sıfır Kurulumlu Ekran Yansıtma (Scrcpy 4.0 Motoru):** 60 FPS ultra düşük gecikme, tam fare ve klavye etkileşimi.
- 🌙 **Gizli Mod (Screen-Off Mode):** Telefon/tablet ekranı fiziksel olarak kararırken tüm kontrolü PC monitöründen sağlama.
- 📁 **Freyja Dropzone:** Masaüstünden dosyaları doğrudan telefonun `/sdcard/Download`, `/sdcard/Music` vb. klasörlerine sürükleyip bırakma.
- 📋 **Evrensel Pano & 2FA / OTP Yakalayıcı:** PC ve mobil cihaz arasında anlık metin/link kopyalama.
- ⚡ **Otomatik Cihaz Telemetrisi:** Canlı pil seviyesi, şarj durumu, dahili depolama ve Android sürüm göstergesi.
- 🔔 **Masaüstü Bildirimleri:** Cihaz bağlandığında veya şarjı %100'e ulaştığında yerel Linux (`notify-send`) ve Windows bildirimleri.
- 🎵 **Freyja Music Player Köprüsü:** `YTMusic_Engine` ile doğrudan oynatma ve durum senkronizasyonu.
- 🌓 **Lüks Cam (Glassmorphic) Tasarım:** Açık/Koyu tema ve GPU Aurora ışık efektleri.

---

## 🚀 Kurulum ve Çalıştırma

### 1. Hazır Binary ile Çalıştırma

#### 🐧 Linux (Freyja OS / Fedora / Ubuntu / Arch)
```bash
# Çalıştırma izni verin ve başlatın
chmod +x bin/freyja-flow-linux-x64
./bin/freyja-flow-linux-x64 --open
```

#### 🪟 Windows
`bin/freyja-flow-windows-x64.exe` dosyasına çift tıklayarak anında başlatabilirsiniz.

---

### 2. Kaynak Koddan Derleme

Gereksinimler: `go >= 1.22`, `adb`, `scrcpy`

```bash
# Otomatik derleme betiğini çalıştırın:
chmod +x build.sh
./build.sh
```

---

## 🏗️ Mimari Yapı

```
Frejya Flow/
├── bin/                          # Derlenmiş tek parça (Single-Binary) çıktılar
│   ├── freyja-flow-linux-x64     # Linux bağımsız çalıştırılabilir dosya
│   └── freyja-flow-windows-x64.exe # Windows çalıştırılabilir dosya
├── backend/                      # Go Backend (REST API + WebSocket + ADB Driver)
│   ├── adb/                      # Cihaz tarama, Scrcpy başlatıcı, tuş enjektörü
│   ├── deck/                     # Freyja Deck & Music Player entegrasyonu
│   ├── sync/                     # Pano ve yerel masaüstü bildirim motoru
│   └── main.go                   # HTTP/WS sunucu & Gömülü frontend
├── frontend/                     # Modern Glassmorphic Web Dashboard
│   ├── css/style.css             # Freyja OS renk ve tema motoru
│   ├── js/                       # WebSocket istemcisi, Dropzone ve Deck yönetimi
│   └── assets/                   # Freyja OS logoları ve simgeler
├── FreyjaFlow.desktop            # Linux uygulama menüsü kısayolu
└── build.sh                      # Otomatik çapraz derleme betiği
```

---

## 📄 Lisans
Bu proje **Freyja OS** ekosisteminin bir parçası olarak açık kaynak lisansıyla sunulmuştur.
