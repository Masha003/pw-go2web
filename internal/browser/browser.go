package browser

import (
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}
	
	return cmd.Start()
}
