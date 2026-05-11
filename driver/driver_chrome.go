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
		//chromedp.UserDataDir(os.TempDir()+"/chrome-tmp"),
		chromedp.ProxyServer(proxy),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent("Chrome"),
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
