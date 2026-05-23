package httpproxy

import (
	"log"
	"net/http"
	"pandaria/cli"

	"github.com/elazarl/goproxy"
)

func Run() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(handleRequest)
	proxy.OnResponse().DoFunc(handleResponse)

	log.Fatal(http.ListenAndServe(cli.HTTPProxyAddr(), proxy))
}

func getContentType(extension string) string {
	switch extension {
	case ".js":
		return "text/javascript"
	case ".json":
		return "application/json"
	case ".html", ".htm":
		return "text/html"
	}
	return "text/html"
}
