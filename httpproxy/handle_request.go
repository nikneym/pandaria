package httpproxy

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/elazarl/goproxy"
)

func handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	dir := sniffDir(req.UserAgent())

	payload := &Payload{dir: dir}
	ctx.UserData = payload

	localPath, extension := urlToLocal(dir, req.URL)
	if isFile(localPath) {
		content, err := os.ReadFile(localPath)
		if err == nil {
			payload.intercepted = true
			return req, goproxy.NewResponse(req, getContentType(extension), http.StatusOK, string(content))
		}
		log.Printf("Failed to read local file: %s - %v", localPath, err)
	}

	return req, nil
}

func sniffDir(ua string) string {
	switch {
	case strings.Contains(ua, "Lightpanda"):
		return "/tmp/pandaria/lp"
	case strings.Contains(ua, "Chrome"):
		return "/tmp/pandaria/chrome"
	default:
		panic("unknown user agent")
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Local file error: %s - %v", path, err)
		}
		return false
	}
	return !info.IsDir()
}
