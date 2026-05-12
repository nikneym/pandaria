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
	//wd, err := os.Getwd()
	//if err != nil {
	//	log.Fatal(err)
	//}

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		//chromedp.UserDataDir(os.TempDir()+"/chrome-tmp"),
		chromedp.ProxyServer(proxy),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent("Chrome"),
		//chromedp.Flag("remote-debugging-port", "8080"),
		chromedp.Flag("enable-logging", true),
		chromedp.Flag("v", "0"),
		//chromedp.UserDataDir(wd),
		//chromedp.Flag("log-level", "0"),
		chromedp.WindowSize(1200, 800),
		chromedp.NoSandbox,
	)

	//f, err := os.Create("chrome-log.txt")
	//if err != nil {
	//	log.Fatal(err)
	//}

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
