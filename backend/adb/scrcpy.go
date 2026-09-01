package adb

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
	scrcpyMutex sync.Mutex
	activeCmd   *exec.Cmd
)

// LaunchScrcpy starts scrcpy with optimized flags for the given mode
func LaunchScrcpy(serial, mode string) error {
	scrcpyMutex.Lock()
	defer scrcpyMutex.Unlock()

	// Eğer zaten çalışan bir scrcpy oturumu varsa temizle
	if activeCmd != nil && activeCmd.Process != nil {
		_ = activeCmd.Process.Kill()
		activeCmd = nil
	}

	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}

	switch mode {
	case "screenoff":
		// Telefonun ekranı kapalı, PC'de canlı kontrol
		args = append(args, "--turn-screen-off", "--stay-awake", "--power-off-on-close")
	case "audio":
		// Sadece ses köprüsü
		args = append(args, "--no-video", "--audio-source=playback")
	case "camera":
		// Arka kamerayı PC'ye 1080p stream et
		args = append(args, "--video-source=camera", "--camera-facing=back")
	default: // "normal"
		// Ultra düşük gecikmeli 60 FPS, ses aktarımı açık
		args = append(args, "--stay-awake", "--max-fps=60")
	}

	cmd := exec.Command("scrcpy", args...)
	
	// Arka planda bağımsız başlat (asenkron)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("scrcpy başlatılamadı: %w", err)
	}

	activeCmd = cmd

	// Proses bittiğinde temizle
	go func() {
		_ = cmd.Wait()
		scrcpyMutex.Lock()
		if activeCmd == cmd {
			activeCmd = nil
		}
		scrcpyMutex.Unlock()
	}()

	return nil
}

// StopScrcpy kills the currently running scrcpy session
func StopScrcpy() {
	scrcpyMutex.Lock()
	defer scrcpyMutex.Unlock()
	if activeCmd != nil && activeCmd.Process != nil {
		_ = activeCmd.Process.Kill()
		activeCmd = nil
	}
}
