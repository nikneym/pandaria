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
	"os/exec"
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

	basePath, urlExt := urlToBase(baseDir, req.URL)
	cachedPath := findCached(basePath, urlExt)
	if cachedPath != "" {
		content, err := os.ReadFile(cachedPath)
		if err == nil {
			data.intercepted = true
			return req, goproxy.NewResponse(req, contentTypeFor(filepath.Ext(cachedPath)), http.StatusOK, string(content))
		}
		log.Printf("Failed to read local file: %s - %v", cachedPath, err)
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

	basePath, urlExt := urlToBase(data.baseDir, req.URL)
	ext := extensionForContentType(res.Header.Get("Content-Type"))
	if ext == "" {
		ext = urlExt
	}
	localPath := basePath + ext

	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		log.Printf("Failed to create dir: %s - %v", dir, err)
	}

	switch ext {
	case ".js":
		f, err := os.Create(localPath)
		if err != nil {
			log.Fatalf("Failed to create file: %s - %v", localPath, err)
		}
		defer f.Close()

		// Run Biome to format.
		cmd := exec.Command("tools/biome", "format", "--stdin-file-path=x.js", "--use-server")
		// Read from response body.
		cmd.Stdin = bytes.NewReader(body)
		// Drain to file.
		cmd.Stdout = f

		err = cmd.Run()
		if err != nil {
			f.Close()
			log.Fatalf("Failed to format file: %s", localPath)
		}
	default:
		if err := os.WriteFile(localPath, body, 0640); err != nil {
			log.Printf("Failed to save locally: %s - %v", localPath, err)
		}
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

// urlToBase returns the cache-file base path (without final extension) and
// the extension derivable from the URL path (e.g. ".js" for "/foo.js", or
// ".html" for directory-like paths). The final on-disk extension is decided
// by the writer using the response's Content-Type, falling back to urlExt.
func urlToBase(baseDir string, u *url.URL) (string, string) {
	path := u.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	var urlExt string
	if len(path) == 0 || path[len(path)-1] == '/' {
		urlExt = ".html"
		path = path + "__index"
	} else {
		urlExt = filepath.Ext(path)
		path = path[0 : len(path)-len(urlExt)]
	}

	hash := ""
	if u.RawQuery != "" {
		hasher := sha1.New()
		hasher.Write([]byte(u.RawQuery))
		hash = "." + hex.EncodeToString(hasher.Sum(nil))
	}
	return filepath.Join(baseDir, "sites", u.Host, path+hash), urlExt
}

// extensionForContentType returns ".js" when the Content-Type indicates
// JavaScript, otherwise "". JS is the only language we care about tagging
// for Biome; everything else is left to the URL-derived extension.
func extensionForContentType(ct string) string {
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil || mt == "" {
		return ""
	}
	switch mt {
	case "application/javascript", "text/javascript", "application/x-javascript":
		return ".js"
	}
	return ""
}

// findCached locates a previously written cache file for basePath, tolerating
// whatever extension the writer chose. Tries basePath+urlExt first as a fast
// path, then scans the directory for any `basePath` or `basePath.<ext>`.
func findCached(basePath, urlExt string) string {
	if urlExt != "" {
		if p := basePath + urlExt; isFile(p) {
			return p
		}
	}
	dir, base := filepath.Split(basePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == base || strings.HasPrefix(name, base+".") {
			return filepath.Join(dir, name)
		}
	}
	return ""
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
