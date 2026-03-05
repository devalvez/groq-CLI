package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

// tool represents a clipboard CLI tool
type tool struct {
	name    string
	copyCmd []string
}

// supported clipboard tools in priority order
var tools = []tool{
	{"xclip", []string{"xclip", "-selection", "clipboard"}},
	{"xsel", []string{"xsel", "--clipboard", "--input"}},
	{"wl-copy", []string{"wl-copy"}},        // Wayland
	{"pbcopy", []string{"pbcopy"}},           // macOS fallback
}

// Copy writes text to the system clipboard.
// It tries each available tool in order.
func Copy(text string) error {
	for _, t := range tools {
		if _, err := exec.LookPath(t.name); err != nil {
			continue // tool not installed
		}
		cmd := exec.Command(t.copyCmd[0], t.copyCmd[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			continue // tool failed, try next
		}
		return nil
	}
	return fmt.Errorf(
		"no clipboard tool found — install one of: xclip, xsel, wl-clipboard\n" +
			"    Ubuntu/Debian:  sudo apt install xclip\n" +
			"    Fedora:         sudo dnf install xclip\n" +
			"    Arch:           sudo pacman -S xclip",
	)
}

// Available returns the name of the first usable clipboard tool,
// or empty string if none is installed.
func Available() string {
	for _, t := range tools {
		if _, err := exec.LookPath(t.name); err == nil {
			return t.name
		}
	}
	return ""
}
