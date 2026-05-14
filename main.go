package main

import (
	"context"
	"log"
	"os/signal"
	"pandaria/cli"
	"pandaria/driver"
	"pandaria/httpproxy"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/chromedp/chromedp"
)

func fetch() {
	// Start HTTP proxy.
	if cli.EnableHTTPProxy() {
		go httpproxy.Run()
		log.Println("Started HTTP proxy server")
	}

	// Start Chrome.
	if cli.EnableChrome() {
		chrome := driver.InitChrome()
		defer chrome.Close()
		log.Println("Initialized Chrome")

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

	log.Println("Exit signal")
}

func main() {
	ctx := kong.Parse(&cli.CLI)
	switch ctx.Command() {
	case "fetch <url>":
		fetch()
	default:
		panic(ctx.Command())
	}
}
