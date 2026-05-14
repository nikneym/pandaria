package driver

import (
	"context"
	"pandaria/cli"

	"github.com/chromedp/chromedp"
)

type Lightpanda struct {
	allocatorCtx    context.Context
	cdpCtx          context.Context
	allocatorCancel context.CancelFunc
	cdpCancel       context.CancelFunc
}

func InitLightpanda() *Lightpanda {
	//lp := exec.Command(
	//	os.Getenv("LIGHTPANDA_BIN"),
	//	"serve",
	//	"--http-proxy",
	//	proxy,
	//	"--host",
	//	"127.0.0.1",
	//	"--port",
	//	"9223",
	//	//"--log-level",
	//	//"debug",
	//	"--insecure-disable-tls-host-verification",
	//)
	//lp.Stderr = os.Stderr
	//err := lp.Start()
	//if err != nil {
	//	log.Fatal(err)
	//}

	allocatorCtx, allocatorCancel := chromedp.NewRemoteAllocator(context.Background(), "ws://"+cli.LPAddr())
	cdpCtx, cdpCancel := chromedp.NewContext(allocatorCtx)

	return &Lightpanda{
		allocatorCtx:    allocatorCtx,
		cdpCtx:          cdpCtx,
		allocatorCancel: allocatorCancel,
		cdpCancel:       cdpCancel,
	}
}

func (lp *Lightpanda) Close() {
	lp.cdpCancel()
	lp.allocatorCancel()
}

func (lp *Lightpanda) Run(actions ...chromedp.Action) error {
	return chromedp.Run(lp.cdpCtx, chromedp.Tasks(actions))
}
