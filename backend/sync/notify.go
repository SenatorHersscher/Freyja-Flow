package sync

import (
	"fmt"
	"os/exec"
	"runtime"
)

// SendDesktopNotification sends a native OS desktop notification
func SendDesktopNotification(title, message string) {
	switch runtime.GOOS {
	case "linux":
		if cmdExists("notify-send") {
			_ = exec.Command("notify-send", "-a", "Freyja Flow", "-i", "phone", title, message).Run()
		}
	case "windows":
		// Windows PowerShell BurntToast veya msg komutu
		psCmd := fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; $template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); $textNodes = $template.GetElementsByTagName("text"); $textNodes.Item(0).AppendChild($template.CreateTextNode("%s")) > $null; $textNodes.Item(1).AppendChild($template.CreateTextNode("%s")) > $null; $notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Freyja Flow"); $notifier.Show([Windows.UI.Notifications.ToastNotification]::new($template))`, title, message)
		_ = exec.Command("powershell", "-NoProfile", "-Command", psCmd).Start()
	}
}
