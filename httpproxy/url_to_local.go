package httpproxy

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var unsafeCharsRegex = regexp.MustCompile(`[\\?%*:|"<>\x00-\x1f]`)

func urlToLocal(baseDir string, u *url.URL) (string, string) {
	path := u.Path
	// Strip leading `/`.
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

	if u.RawQuery != "" {
		path += "__" + strings.ReplaceAll(u.RawQuery, "/", "-")
	}
	safe := unsafeCharsRegex.ReplaceAllString(path, "_")
	if len(safe) > 125 {
		safe = safe[0:125] + fmt.Sprintf("_%x", sha1.Sum([]byte(safe)))
	}
	return filepath.Join(baseDir, "sites", u.Host, safe+extension), extension
}
