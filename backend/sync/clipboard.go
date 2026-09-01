package sync

import (
	"os/exec"
	"strings"
)

// ReadDesktopClipboard reads the current text from Linux clipboard (Wayland / X11)
func ReadDesktopClipboard() string {
	// Wayland
	if cmdExists("wl-paste") {
		out, err := exec.Command("wl-paste", "--no-newline").Output()
		if err == nil {
			return string(out)
		}
	}

	// X11 / Fallback
	if cmdExists("xclip") {
		out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
		if err == nil {
			return string(out)
		}
	}

	return ""
}

// WriteDesktopClipboard writes text to Linux clipboard (Wayland / X11)
func WriteDesktopClipboard(text string) error {
	if cmdExists("wl-copy") {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	if cmdExists("xclip") {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	return nil
}

func cmdExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
