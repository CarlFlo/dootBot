package utils

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/CarlFlo/malm"
)

// Clear for windows
func Clear() {

	// Check os and use correct settings
	switch currentOS := runtime.GOOS; currentOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "darwin": // Mac
		fallthrough
	case "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

	default:

		malm.Warn("Currently running on %s. No clear command for this type", currentOS)
	}
}
