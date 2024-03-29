package svg

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func Visualize(ctx context.Context, suffix string, input io.Reader) error {
	tempFile, err := newTempFile(os.TempDir(), "pprof", "."+suffix)
	if err != nil {
		return err
	}
	deferDeleteTempFile(tempFile.Name())
	if _, err := io.Copy(tempFile, input); err != nil {
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	// Try visualizers until one is successful
	for _, v := range browsers() {
		// Separate command and arguments for exec.Command.
		args := strings.Split(v, " ")
		if len(args) == 0 {
			continue
		}
		viewer := exec.Command(args[0], append(args[1:], tempFile.Name())...)
		viewer.Stderr = os.Stderr
		if err = viewer.Start(); err == nil {
			return nil
		}
	}
	return err
}

// browsers returns a list of commands to attempt for web visualization.
func browsers() []string {
	var cmds []string
	if userBrowser := os.Getenv("BROWSER"); userBrowser != "" {
		cmds = append(cmds, userBrowser)
	}
	switch runtime.GOOS {
	case "darwin":
		cmds = append(cmds, "/usr/bin/open")
	case "windows":
		cmds = append(cmds, "cmd /c start")
	default:
		// Commands opening browsers are prioritized over xdg-open, so browser()
		// command can be used on linux to open the .svg file generated by the -web
		// command (the .svg file includes embedded javascript so is best viewed in
		// a browser).
		cmds = append(cmds, []string{"chrome", "google-chrome", "chromium", "firefox", "sensible-browser"}...)
		if os.Getenv("DISPLAY") != "" {
			// xdg-open is only for use in a desktop environment.
			cmds = append(cmds, "xdg-open")
		}
	}
	return cmds
}

// newTempFile returns a new output file in dir with the provided prefix and suffix.
func newTempFile(dir, prefix, suffix string) (*os.File, error) {
	for index := 1; index < 10000; index++ {
		switch f, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("%s%03d%s", prefix, index, suffix)), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666); {
		case err == nil:
			return f, nil
		case !os.IsExist(err):
			return nil, err
		}
	}
	// Give up
	return nil, fmt.Errorf("could not create file of the form %s%03d%s", prefix, 1, suffix)
}

var tempFiles []string
var tempFilesMu = sync.Mutex{}

// deferDeleteTempFile marks a file to be deleted by next call to Cleanup()
func deferDeleteTempFile(path string) {
	tempFilesMu.Lock()
	tempFiles = append(tempFiles, path)
	tempFilesMu.Unlock()
}
