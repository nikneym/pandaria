package driver

import (
	"context"
	"log"
	"pandaria/cli"

	"github.com/chromedp/chromedp"
)

type Chrome struct {
	allocatorCtx    context.Context
	cdpCtx          context.Context
	allocatorCancel context.CancelFunc
	cdpCancel       context.CancelFunc
}

// InitChrome initializes a Chrome driver.
func InitChrome() *Chrome {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Chrome"),
		chromedp.Flag("enable-logging", true),
		chromedp.Flag("v", "1"),
		chromedp.UserDataDir("/tmp/pandaria/chrome-userdata-dir"),
		chromedp.Flag("remote-debugging-port", "8080"),
		chromedp.Flag("headless", cli.ChromeHeadless()),
		chromedp.NoSandbox,
	)

	if cli.EnableHTTPProxy() {
		opts = append(
			opts,
			chromedp.ProxyServer(cli.HTTPProxyAddr()),
			chromedp.Flag("disable-web-security", true),
			chromedp.Flag("ignore-certificate-errors", true),
		)
	}

	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	cdpCtx, cdpCancel := chromedp.NewContext(allocatorCtx)

	return &Chrome{
		allocatorCtx:    allocatorCtx,
		cdpCtx:          cdpCtx,
		allocatorCancel: allocatorCancel,
		cdpCancel:       cdpCancel,
	}
}

func (c *Chrome) Close() {
	c.cdpCancel()
	c.allocatorCancel()
}

func (c *Chrome) Run(actions ...chromedp.Action) {
	err := chromedp.Run(c.cdpCtx, chromedp.Tasks(actions))
	if err != nil {
		log.Fatal(err)
	}
}
