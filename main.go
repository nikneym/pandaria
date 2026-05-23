package main

import (
	"context"
	"log"
	"os/signal"
	"pandaria/biome"
	"pandaria/cli"
	"pandaria/config"
	"pandaria/driver"
	"pandaria/httpproxy"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/chromedp/chromedp"
)

func fetch() {
	// Start Biome daemon.
	err := biome.Start()
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
	err = biome.Stop()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("stopped Biome daemon")

	log.Println("exit signal")
}

func main() {
	// Init config.
	err := config.Setup()
	if err != nil {
		log.Fatal(err)
	}

	ctx := kong.Parse(&cli.CLI)
	switch ctx.Command() {
	case "fetch <url>":
		// Check if we have to update Biome.
		if !biome.IsCorrectVersion() {
			log.Printf("updating Biome")
			err := biome.Install()
			if err != nil {
				log.Fatal(err)
			}
		}

		fetch()
	default:
		panic(ctx.Command())
	}
}
