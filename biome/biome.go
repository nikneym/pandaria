package biome

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"pandaria/config"
	"runtime"
	"strings"
)

// Version is Biome version we prefer.
const Version = "2.4.15"

// IsCorrectVersion returns true if Biome is installed and is correct version.
func IsCorrectVersion() bool {
	cmd := exec.Command(config.C.BiomePath, "-V")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return false
	}

	line, _ := buf.ReadBytes('\n')
	v, ok := strings.CutPrefix(strings.TrimSpace(string(line)), "Version: ")
	if !ok {
		return false
	}
	return v == Version
}

// Install installs a fresh Biome executable for desired version.
// Returns an error if installation failed.
func Install() error {
	// Where to install.
	f, err := os.Create(config.C.BiomePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// This is a flaky way to create installation URL; should be revisited
	// if it fails on some systems.
	format := "https://github.com/biomejs/biome/releases/download/@biomejs/biome@%s/biome-%s-%s"
	url := fmt.Sprintf(format, Version, runtime.GOOS, runtime.GOARCH)

	resp, err := http.Get(url)
	if err != nil {
		f.Close()
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		resp.Body.Close()
		return err
	}

	// Make runnable.
	return exec.Command("chmod", "+x", config.C.BiomePath).Run()
}

func Start() error {
	return exec.Command(config.C.BiomePath, "start").Run()
}

func Stop() error {
	return exec.Command(config.C.BiomePath, "stop").Run()
}

// FormatByMedia reads from reader and writes formatted bytes to writer.
func FormatByMedia(mediatype string, dst io.Writer, src io.Reader) error {
	var arg string
	switch mediatype {
	case "application/javascript", "text/javascript", "application/x-javascript":
		arg = "--stdin-file-path=x.js"
	default:
		// Nothing to format.
		_, err := io.Copy(dst, src)
		return err
	}

	cmd := exec.Command(config.C.BiomePath, "format", arg, "--use-server")
	cmd.Stdin = src
	cmd.Stdout = dst
	return cmd.Run()
}
