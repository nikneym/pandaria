package httpproxy

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"pandaria/cli"
	"path/filepath"
	"strings"

	"github.com/elazarl/goproxy"
)

const (
	dirLightpanda = "/tmp/pandaria/lp"
	dirChrome     = "/tmp/pandaria/chrome"
)

type ctxData struct {
	baseDir     string
	intercepted bool
}

func Run() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(handleRequest)
	proxy.OnResponse().DoFunc(handleResponse)

	log.Fatal(http.ListenAndServe(cli.HTTPProxyAddr(), proxy))
}

func handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	baseDir := classify(req.UserAgent())
	if baseDir == "" {
		return req, nil
	}
	data := &ctxData{baseDir: baseDir}
	ctx.UserData = data

	localPath, extension := urlToLocal(baseDir, req.URL)
	if isFile(localPath) {
		content, err := os.ReadFile(localPath)
		if err == nil {
			data.intercepted = true
			return req, goproxy.NewResponse(req, contentTypeFor(extension), http.StatusOK, string(content))
		}
		log.Printf("Failed to read local file: %s - %v", localPath, err)
	}
	return req, nil
}

func handleResponse(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if res.StatusCode != 200 {
		return res
	}
	data, ok := ctx.UserData.(*ctxData)
	if !ok || data == nil || data.intercepted {
		return res
	}

	req := ctx.Req
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("Failed to read response: %s - %v", req.URL, err)
		return goproxy.NewResponse(req, "text/plain", http.StatusInternalServerError, err.Error())
	}

	localPath, _ := urlToLocal(data.baseDir, req.URL)
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		log.Printf("Failed to create dir: %s - %v", dir, err)
	}
	if err := os.WriteFile(localPath, body, 0640); err != nil {
		log.Printf("Failed to save locally: %s - %v", localPath, err)
	}

	res.Body = io.NopCloser(bytes.NewReader(body))
	return res
}

func classify(ua string) string {
	switch {
	case strings.Contains(ua, "Lightpanda"):
		return dirLightpanda
	case strings.Contains(ua, "Chrome"):
		return dirChrome
	default:
		panic("unknown user agent")
	}
}

func urlToLocal(baseDir string, u *url.URL) (string, string) {
	path := u.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	var extension string
	if len(path) == 0 || path[len(path)-1] == '/' {
		extension = ".html"
		path = path + "__index"
	} else {
		extension = filepath.Ext(path)
		path = path[0 : len(path)-len(extension)]
	}

	hash := ""
	if u.RawQuery != "" {
		hasher := sha1.New()
		hasher.Write([]byte(u.RawQuery))
		hash = "." + hex.EncodeToString(hasher.Sum(nil))
	}
	return filepath.Join(baseDir, "sites", u.Host, path+hash+extension), extension
}

func contentTypeFor(extension string) string {
	if ct := mime.TypeByExtension(extension); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

func isFile(path string) bool {
	fo, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Local file error: %s - %v", path, err)
		}
		return false
	}
	return !fo.IsDir()
}
