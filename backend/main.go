package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"freyja-flow/adb"
	"freyja-flow/deck"
	flowSync "freyja-flow/sync"

	"github.com/gorilla/websocket"
)

//go:embed all:frontend_embed
var embeddedFrontend embed.FS

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

type WSMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func main() {
	openBrowser := flag.Bool("open", false, "Başlangıçta tarayıcıda otomatik aç")
	customPort := flag.Int("port", 8090, "Sunucu portu")
	flag.Parse()

	listener, port, err := findAvailableListener(*customPort)
	if err != nil {
		log.Fatalf("Port açılamadı: %v", err)
	}

	localIP := getLocalIP()
	serverURL := fmt.Sprintf("http://localhost:%d", port)
	lanURL := fmt.Sprintf("http://%s:%d", localIP, port)

	fmt.Printf("\n🔮 ===================================================\n")
	fmt.Printf("   FREYJA FLOW - Ekosistem & Cihaz Köprüsü v1.0.0\n")
	fmt.Printf("   Yerel Adres (PC):  %s\n", serverURL)
	fmt.Printf("   Ağ Adresi (Mobil): %s\n", lanURL)
	fmt.Printf("   Platform:          %s / %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("===================================================\n\n")

	// Statik Arayüz (Embedded / Gömülü Dosya Sistemi)
	var fileServer http.Handler
	subFS, err := fs.Sub(embeddedFrontend, "frontend_embed")
	if err == nil {
		fileServer = http.FileServer(http.FS(subFS))
	} else {
		// Geliştirme modu yedeği
		fileServer = http.FileServer(http.Dir("../frontend"))
	}
	http.Handle("/", fileServer)

	// WebSocket Uç Noktası
	http.HandleFunc("/ws", handleWebSocket)

	// REST API Uç Noktaları
	http.HandleFunc("/api/devices", handleGetDevices)
	http.HandleFunc("/api/connect", handleConnectWireless)
	http.HandleFunc("/api/mirror", handleLaunchMirror)
	http.HandleFunc("/api/upload", handleFileUpload)
	http.HandleFunc("/api/input/key", handleInjectKey)
	http.HandleFunc("/api/clipboard/sync", handleClipboardSync)
	http.HandleFunc("/api/music/status", handleMusicStatus)
	http.HandleFunc("/api/music/control", handleMusicControl)
	http.HandleFunc("/api/phone/files", handleListPhoneFiles)
	http.HandleFunc("/api/phone/pull", handlePullPhoneFile)
	http.HandleFunc("/api/phone/quick-photo", handleQuickPhoto)

	// Arka plan cihaz durumu, bildirimleri ve telemetri yayıncısı
	go startDeviceTelemetryBroadcast()

	// Otomatik tarayıcı açma
	if *openBrowser {
		go func() {
			time.Sleep(300 * time.Millisecond)
			openURL(serverURL)
		}()
	}

	log.Fatal(http.Serve(listener, nil))
}

func findAvailableListener(startPort int) (net.Listener, int, error) {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			return ln, port, nil
		}
	}
	return nil, 0, fmt.Errorf("uygun port bulunamadı")
}

/* ==================== WEBSOCKET YÖNETİCİSİ ==================== */

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	// İlk bağlantıda mevcut cihazları hemen ilet
	devs, _ := adb.GetDevices()
	if len(devs) > 0 {
		_ = conn.WriteJSON(WSMessage{Event: "DEVICE_STATE", Data: devs[0]})
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			break
		}
	}
}

func broadcast(msg WSMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func startDeviceTelemetryBroadcast() {
	var lastSerial string
	var lastLevel int
	var notifiedFullBattery bool

	for {
		time.Sleep(3 * time.Second)
		devs, err := adb.GetDevices()
		if err != nil || len(devs) == 0 {
			if lastSerial != "" {
				broadcast(WSMessage{Event: "DEVICE_DISCONNECTED", Data: nil})
				flowSync.SendDesktopNotification("Freyja Flow", "📱 Cihaz bağlantısı kesildi.")
				lastSerial = ""
				notifiedFullBattery = false
			}
			continue
		}

		currentDev := devs[0]
		if currentDev.Serial != lastSerial {
			broadcast(WSMessage{Event: "DEVICE_CONNECTED", Data: currentDev})
			flowSync.SendDesktopNotification("Freyja Flow", fmt.Sprintf("📱 %s başarıyla bağlandı (%s)!", currentDev.Model, currentDev.ConnectedVia))
			lastSerial = currentDev.Serial
			lastLevel = currentDev.Battery.Level
		} else if currentDev.Battery.Level != lastLevel {
			broadcast(WSMessage{Event: "BATTERY_UPDATE", Data: currentDev.Battery})
			lastLevel = currentDev.Battery.Level

			// %100 Şarj Bildirimi
			if currentDev.Battery.Level >= 100 && currentDev.Battery.Charging && !notifiedFullBattery {
				flowSync.SendDesktopNotification("Freyja Flow", fmt.Sprintf("⚡ %s şarjı tamamen doldu (%%100)!", currentDev.Model))
				notifiedFullBattery = true
			}
		}
	}
}

/* ==================== REST API İŞLEYİCİLERİ ==================== */

func handleGetDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	devs, err := adb.GetDevices()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error(), "devices": []adb.DeviceState{}})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "devices": devs})
}

func handleConnectWireless(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek", http.StatusBadRequest)
		return
	}

	err := adb.ConnectWireless(req.IP)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleLaunchMirror(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Mode   string `json:"mode"`
		Serial string `json:"serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek", http.StatusBadRequest)
		return
	}

	err := adb.LaunchScrcpy(req.Serial, req.Mode)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseMultipartForm(500 << 20) // 500 MB max
	if err != nil {
		http.Error(w, "Dosya okunamadı", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Dosya bulunamadı", http.StatusBadRequest)
		return
	}
	defer file.Close()

	targetDir := r.FormValue("target")
	if targetDir == "" {
		targetDir = "/sdcard/Download/"
	}

	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, header.Filename)
	out, err := os.Create(tempFilePath)
	if err != nil {
		http.Error(w, "Geçici dosya oluşturulamadı", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	defer os.Remove(tempFilePath)

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Dosya yazılamadı", http.StatusInternalServerError)
		return
	}

	err = adb.PushFile("", tempFilePath, targetDir)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	flowSync.SendDesktopNotification("Freyja Flow", fmt.Sprintf("📁 '%s' başarıyla telefona aktarıldı!", header.Filename))
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "filename": header.Filename})
}

func handleInjectKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Action string `json:"action"`
		Serial string `json:"serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek", http.StatusBadRequest)
		return
	}

	err := adb.InjectKey(req.Serial, req.Action)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleClipboardSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clipText := flowSync.ReadDesktopClipboard()
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "text": clipText})
}

func handleMusicStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := deck.GetYTMusicStatus()
	json.NewEncoder(w).Encode(status)
}

func handleMusicControl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	action := r.URL.Query().Get("action")
	err := deck.ControlYTMusic(action)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleListPhoneFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir := r.URL.Query().Get("dir")
	serial := r.URL.Query().Get("serial")
	files, err := adb.ListPhoneFiles(serial, dir)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error(), "files": []adb.PhoneFileItem{}})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "files": files})
}

func handlePullPhoneFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		RemotePath string `json:"remote_path"`
		Serial     string `json:"serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek", http.StatusBadRequest)
		return
	}

	err := adb.PullFile(req.Serial, req.RemotePath, "")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	flowSync.SendDesktopNotification("Freyja Flow", fmt.Sprintf("📥 '%s' bilgisayarınıza indirildi (~/Downloads)!", filepath.Base(req.RemotePath)))
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "filename": filepath.Base(req.RemotePath)})
}

func handleQuickPhoto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	serial := r.URL.Query().Get("serial")
	filename, err := adb.GetLatestPhoto(serial)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	flowSync.SendDesktopNotification("Freyja Flow", fmt.Sprintf("📷 Son fotoğraf '%s' bilgisayarınıza indirildi!", filename))
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "filename": filename})
}

/* ==================== YARDIMCI FONKSİYONLAR ==================== */

func openURL(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows'ta Edge veya Chrome App-Mode (Bağımsız Masaüstü Penceresi)
		edgePath := `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
		if _, err := os.Stat(edgePath); err == nil {
			cmd = exec.Command(edgePath, fmt.Sprintf("--app=%s", url), "--window-size=1280,820")
		} else {
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		}
	case "darwin":
		cmd = exec.Command("open", url)
	default: // "linux"
		// 1. Tercih: Yerel WebKit2 / GTK Penceresi (Sıfır tarayıcı sekmesi, gerçek yerel pencere)
		exePath, _ := os.Executable()
		rootDir := filepath.Dir(filepath.Dir(exePath))
		guiPy := filepath.Join(rootDir, "gui.py")

		if _, err := os.Stat(guiPy); err == nil {
			cmd = exec.Command("python3", guiPy, url)
		} else if _, err := os.Stat("gui.py"); err == nil {
			cmd = exec.Command("python3", "gui.py", url)
		} else {
			// 2. Tercih: Chromium tabanlı tarayıcılarda App-Mode
			browsers := []string{"google-chrome", "chromium-browser", "chromium", "brave-browser", "microsoft-edge"}
			var chosenBrowser string
			for _, b := range browsers {
				if path, err := exec.LookPath(b); err == nil {
					chosenBrowser = path
					break
				}
			}

			if chosenBrowser != "" {
				cmd = exec.Command(chosenBrowser, fmt.Sprintf("--app=%s", url), "--window-size=1280,820", "--class=FreyjaFlow")
			} else {
				cmd = exec.Command("xdg-open", url)
			}
		}
	}

	if cmd != nil {
		_ = cmd.Start()
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && (strings.HasPrefix(ipnet.IP.String(), "192.168.") || strings.HasPrefix(ipnet.IP.String(), "10.")) {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
