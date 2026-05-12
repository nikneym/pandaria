package driver

import (
	"context"

	"github.com/chromedp/chromedp"
)

type Chrome struct {
	allocatorCtx    context.Context
	cdpCtx          context.Context
	allocatorCancel context.CancelFunc
	cdpCancel       context.CancelFunc
}

// InitChrome initializes a Chrome driver.
func InitChrome(proxy string) *Chrome {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(proxy),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent("Chrome"),
		chromedp.Flag("enable-logging", true),
		chromedp.Flag("v", "1"),
		chromedp.UserDataDir("/tmp/pandaria/chrome-userdata-dir"),
		chromedp.WindowSize(1200, 800),
		chromedp.NoSandbox,
		//chromedp.Flag("remote-debugging-port", "8080"),
	)

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

func (c *Chrome) Run(actions ...chromedp.Action) error {
	return chromedp.Run(c.cdpCtx, chromedp.Tasks(actions))
}
