package cli

type FetchOptions struct {
	URL             string `arg:"" help:"Where to fetch."`
	HTTPProxyAddr   string `help:"Address of HTTP proxy server." default:"127.0.0.1:3000"`
	LPAddr          string `help:"Address of Lightpanda CDP server." default:"127.0.0.1:9222"`
	EnableHTTPProxy bool   `help:"Whether run an HTTP proxy server." default:"true"`
	EnableLP        bool   `help:"Whether connect to a Lightpanda instance." default:"true"`
	EnableChrome    bool   `help:"Whether run a Chrome instance." default:"true"`
	ChromeHeadless  bool   `help:"Whether run Chrome in a headless mode." default:"true"`
}

type SearchOptions struct {
	Pattern string `arg:"" help:"What to search."`
}

type Commands struct {
	Fetch  FetchOptions  `cmd:"" help: "Fetch from a URL."`
	Search SearchOptions `cmd:"" help: "Search for a pattern in MITM'd files."`
}

var CLI Commands

func URL() string {
	return CLI.Fetch.URL
}

func HTTPProxyAddr() string {
	return CLI.Fetch.HTTPProxyAddr
}

func LPAddr() string {
	return CLI.Fetch.LPAddr
}

func EnableHTTPProxy() bool {
	return CLI.Fetch.EnableHTTPProxy
}

func EnableLP() bool {
	return CLI.Fetch.EnableLP
}

func EnableChrome() bool {
	return CLI.Fetch.EnableChrome
}

func ChromeHeadless() bool {
	return CLI.Fetch.ChromeHeadless
}

func SearchPattern() string {
	return CLI.Search.Pattern
}
