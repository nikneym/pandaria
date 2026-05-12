package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"pandaria/driver"
	"pandaria/httpproxy"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/chromedp/chromedp"
)

func start(d driver.Driver, wg *sync.WaitGroup) {
	defer wg.Done()

	err := d.Run(
		chromedp.Navigate("https://www.airbnb.com/"),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func fetch(options FetchOptions) {
	httpProxyAddr := options.HttpProxyAddr

	// Start HTTP proxy.
	go httpproxy.Run(httpProxyAddr)

	var wg sync.WaitGroup

	// Start Chrome.
	if options.EnableChrome {
		chrome := driver.InitChrome(httpProxyAddr)
		defer chrome.Close()
		wg.Add(1)
		go start(chrome, &wg)
	}

	// Start LP.
	if options.EnableLP {
		lp := driver.InitLightpanda(httpProxyAddr)
		defer lp.Close()
		wg.Add(1)
		go start(lp, &wg)
	}

	wg.Wait()
	fmt.Println("fetch completed")
}

func chromeDebugLog() {
	f, err := os.Open("/tmp/pandaria/chrome-userdata-dir/chrome_debug.log")
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(os.Stdout, f)
	if err != nil {
		log.Fatal(err)
	}
}

type FetchOptions struct {
	Url                       string `arg:"" help:"Where to fetch from."`
	HttpProxyAddr             string `help:"Address of HTTP proxy server." default:"127.0.0.1:3000"`
	LPAddr                    string `help:"Address of Lightpanda CDP server." default:"127.0.0.1:9222"`
	ChromeRemoteDebuggingPort uint16 `help:"Port for remote debugging Chrome (--remote-debugging-port)." default:"9223"`
	EnableChrome              bool   `help:"Whether run a Chrome instance." default:"true"`
	EnableLP                  bool   `help:"Whether connect to a Lightpanda instance." default:"true"`
}

var CLI struct {
	Fetch          FetchOptions `cmd:"" help: "Fetch from a URL."`
	ChromeDebugLog struct{}     `cmd:"" help: "Writes \"chrome_debug.log\" to stdout."`
}

func main() {
	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}

	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "fetch <url>":
		fetch(CLI.Fetch)
	case "chrome-debug-log":
		chromeDebugLog()
	default:
		panic(ctx.Command())
	}

}
