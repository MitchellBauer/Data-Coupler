package ui

import (
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
)

// openInFileBrowser opens the OS file manager with the given path selected.
func openInFileBrowser(path string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("explorer", "/select,"+path).Start() //nolint:errcheck
	case "darwin":
		exec.Command("open", "-R", path).Start() //nolint:errcheck
	default:
		exec.Command("xdg-open", filepath.Dir(path)).Start() //nolint:errcheck
	}
}

// uriToFilePath converts a Fyne URI to an OS-native file path.
// On Windows, Fyne file dialog URIs can carry a leading path separator before
// the drive letter (e.g. \C:\file.csv). This trims that prefix so the path is
// usable directly with os.Open and friends.
func uriToFilePath(u fyne.URI) string {
	p := filepath.FromSlash(u.Path())
	// Windows: trim leading separator when followed by a drive letter, e.g. \C:\...
	if len(p) > 2 && p[0] == filepath.Separator && p[2] == ':' {
		p = p[1:]
	}
	return p
}
