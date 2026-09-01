package adb

import (
	"fmt"
	"os/exec"
	"strings"
)

// InjectKey sends an Android keyevent or statusbar command to the device
func InjectKey(serial, action string) error {
	var cmd *exec.Cmd

	baseArgs := []string{}
	if serial != "" {
		baseArgs = append(baseArgs, "-s", serial)
	}

	switch strings.ToUpper(action) {
	case "HOME":
		cmd = exec.Command("adb", append(baseArgs, "shell", "input", "keyevent", "3")...)
	case "BACK":
		cmd = exec.Command("adb", append(baseArgs, "shell", "input", "keyevent", "4")...)
	case "RECENTS":
		cmd = exec.Command("adb", append(baseArgs, "shell", "input", "keyevent", "187")...)
	case "NOTIFICATIONS":
		cmd = exec.Command("adb", append(baseArgs, "shell", "cmd", "statusbar", "expand-notifications")...)
	case "SCREENSHOT":
		cmd = exec.Command("adb", append(baseArgs, "shell", "input", "keyevent", "120")...)
	case "VOLUME_UP":
		cmd = exec.Command("adb", append(baseArgs, "shell", "input", "keyevent", "24")...)
	case "VOLUME_DOWN":
		cmd = exec.Command("adb", append(baseArgs, "shell", "input", "keyevent", "25")...)
	case "POWER":
		cmd = exec.Command("adb", append(baseArgs, "shell", "input", "keyevent", "26")...)
	default:
		return fmt.Errorf("bilinmeyen eylem: %s", action)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tuş enjeksiyonu başarısız: %s (%w)", string(out), err)
	}
	return nil
}

// InjectText sends text directly to the focused input field on the phone
func InjectText(serial, text string) error {
	// ADB shell input text boşlukları %s olarak bekler
	escaped := strings.ReplaceAll(text, " ", "%s")
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "shell", "input", "text", escaped)
	cmd := exec.Command("adb", args...)
	return cmd.Run()
}
