package httpproxy

import (
	"bytes"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"pandaria/biome"
	"path/filepath"

	"github.com/elazarl/goproxy"
)

func handleResponse(res *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if res.StatusCode != 200 {
		return res
	}

	payload, ok := ctx.UserData.(*Payload)
	if !ok || payload == nil || payload.intercepted {
		return res
	}

	req := ctx.Req

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response: %s - %v", req.URL, err)
		return goproxy.NewResponse(req, "text/html", http.StatusInternalServerError, err.Error())
	}

	localPath, extension := urlToLocal(payload.dir, req.URL)

	if err := os.MkdirAll(filepath.Dir(localPath), 0750); err != nil {
		log.Printf("Failed to create dir: %s - %v", filepath.Dir(localPath), err)
	}

	f, err := os.Create(localPath)
	if err != nil {
		log.Fatalf("Failed to create file: %s - %v", localPath, err)
	}
	defer f.Close()

	// Format by media type for the on-disk copy.
	mt, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err := biome.FormatByMedia(mt, f, bytes.NewReader(data)); err != nil {
		log.Printf("Failed to save locally: %s - %v", localPath, err)
	}

	return goproxy.NewResponse(req, getContentType(extension), http.StatusOK, string(data))
}
