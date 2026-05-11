package driver

import "github.com/chromedp/chromedp"

type Driver interface {
	Close()
	Run(actions ...chromedp.Action) error
}
