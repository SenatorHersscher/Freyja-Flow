package adb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PushFile pushes a local file to the specified Android path and triggers media scanner
func PushFile(serial, localPath, targetDir string) error {
	if targetDir == "" {
		targetDir = "/sdcard/Download/"
	}

	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "push", localPath, targetDir)

	cmd := exec.Command("adb", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb push hatası: %s (%w)", string(out), err)
	}

	// Android Medya Tarayıcısını tetikle ki müzik/galeri anında görsün
	filename := filepath.Base(localPath)
	destPath := filepath.Join(targetDir, filename)

	scanArgs := []string{}
	if serial != "" {
		scanArgs = append(scanArgs, "-s", serial)
	}
	scanArgs = append(scanArgs, "shell", "am", "broadcast", "-a", "android.intent.action.MEDIA_SCANNER_SCAN_FILE", "-d", "file://"+destPath)
	_ = exec.Command("adb", scanArgs...).Run()

	return nil
}

// PullFile downloads a file from Android phone to PC
func PullFile(serial, remotePath, localDest string) error {
	if localDest == "" {
		homeDir, _ := os.UserHomeDir()
		localDest = filepath.Join(homeDir, "Downloads")
	}

	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "pull", remotePath, localDest)

	cmd := exec.Command("adb", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb pull hatası: %s (%w)", string(out), err)
	}
	return nil
}

// PhoneFileItem represents a remote file on Android
type PhoneFileItem struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    string `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

// ListPhoneFiles lists files from a given Android folder (e.g. /sdcard/Download or /sdcard/DCIM/Camera)
func ListPhoneFiles(serial, remoteDir string) ([]PhoneFileItem, error) {
	if remoteDir == "" {
		remoteDir = "/sdcard/Download"
	}

	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	// Son değiştirilen 25 dosyayı listele
	args = append(args, "shell", "ls", "-t", "-p", remoteDir)

	cmd := exec.Command("adb", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var items []PhoneFileItem

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isDir := strings.HasSuffix(line, "/")
		cleanName := strings.TrimSuffix(line, "/")

		items = append(items, PhoneFileItem{
			Name:  cleanName,
			Path:  filepath.Join(remoteDir, cleanName),
			IsDir: isDir,
		})

		if len(items) >= 20 {
			break
		}
	}

	return items, nil
}

// GetLatestPhoto pulls the newest camera photo or screenshot directly to PC Downloads
func GetLatestPhoto(serial string) (string, error) {
	// Önce DCIM/Camera'ya bak, yoksa Pictures/Screenshots'a bak
	dirs := []string{"/sdcard/DCIM/Camera", "/sdcard/DCIM", "/sdcard/Pictures/Screenshots", "/sdcard/Download"}
	
	for _, dir := range dirs {
		files, err := ListPhoneFiles(serial, dir)
		if err == nil {
			for _, f := range files {
				if !f.IsDir && (strings.HasSuffix(strings.ToLower(f.Name), ".jpg") || 
					strings.HasSuffix(strings.ToLower(f.Name), ".png") || 
					strings.HasSuffix(strings.ToLower(f.Name), ".mp4")) {
					
					homeDir, _ := os.UserHomeDir()
					destDir := filepath.Join(homeDir, "Downloads")
					err := PullFile(serial, f.Path, destDir)
					if err == nil {
						return f.Name, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("telefonda fotoğraf bulunamadı")
}
