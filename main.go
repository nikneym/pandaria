package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"pandaria/cli"
	"pandaria/driver"
	"pandaria/httpproxy"
	"runtime"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/chromedp/chromedp"
)

func fetch() {
	// Start Biome daemon.
	err := exec.Command(TOOLS_DIR+"biome", "start").Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("started Biome daemon")

	// Start HTTP proxy.
	if cli.EnableHTTPProxy() {
		go httpproxy.Run()
		log.Println("started HTTP proxy server")
	}

	// Start Chrome.
	if cli.EnableChrome() {
		chrome := driver.InitChrome()
		defer chrome.Close()
		log.Println("initialized Chrome")

		go chrome.Run(
			chromedp.Navigate(cli.URL()),
		)
	}

	// Start LP.
	if cli.EnableLP() {
		lp := driver.InitLightpanda()
		defer lp.Close()

		go lp.Run(
			chromedp.Navigate(cli.URL()),
		)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	// Wait until exit signal.
	<-ctx.Done()

	// Stop Biome daemon.
	err = exec.Command(TOOLS_DIR+"biome", "stop").Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("stopped Biome daemon")

	log.Println("exit signal")
}

func search() {
	pattern := fmt.Sprintf("`%s`", cli.SearchPattern())
	cmd := exec.Command(
		TOOLS_DIR+"biome",
		"search",
		pattern,
		"/tmp/pandaria/chrome/sites/",
		"--colors=off",
		"--config-path=./biome.jsonc",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

const TOOLS_DIR = "tools/"
const BIOME_VERSION = "2.4.15"

// isCorrectBiomeVersion returns true if Biome is installed and is correct version.
func isCorrectBiomeVersion() bool {
	// Check if we have Biome installed and running correct version.
	cmd := exec.Command(TOOLS_DIR+"biome", "-V")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf

	err := cmd.Run()
	if err != nil {
		return false
	}

	line, err := buf.ReadBytes('\n')
	if err != nil {
		log.Fatal("executable is not Biome")
	}

	version := string(line[9 : len(line)-1])
	return version == BIOME_VERSION
}

// installBiome installs a fresh Biome executable for desired version.
// Returns an error if installation failed.
func installBiome(v string) error {
	path := TOOLS_DIR + "biome"
	// Where to install.
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	// This is a flaky way to create installation URL; should be revisited
	// if it fails on some systems.
	format := "https://github.com/biomejs/biome/releases/download/@biomejs/biome@%s/biome-%s-%s"
	url := fmt.Sprintf(format, v, runtime.GOOS, runtime.GOARCH)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	// Make runnable.
	err = exec.Command("chmod", "+x", path).Run()
	return err
}

func main() {
	// Create TOOLS_DIR if not exists.
	{
		err := os.Mkdir(TOOLS_DIR, 0755)
		_ = err
	}

	ctx := kong.Parse(&cli.CLI)
	switch ctx.Command() {
	case "fetch <url>":
		// Check if we have to update Biome.
		if !isCorrectBiomeVersion() {
			log.Printf("updating Biome")
			err := installBiome(BIOME_VERSION)
			if err != nil {
				log.Fatal(err)
			}
		}

		fetch()
	case "search <pattern>":
		search()
	default:
		panic(ctx.Command())
	}
}
