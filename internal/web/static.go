// Package web serves the static frontend assets alongside the API.
package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Handler serves static files from dir, falling back to index.html for SPA routing.
func Handler(dir string) http.Handler {
	index := filepath.Join(dir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		if name, ok := resolve(dir, r.URL.Path); ok {
			// Cache fingerprinted assets indefinitely.
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeFile(w, r, name)
			return
		}

		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, index)
	})
}

// resolve maps a URL path to a file inside dir, preventing directory traversal.
func resolve(dir, urlPath string) (string, bool) {
	clean := filepath.Clean("/" + strings.TrimPrefix(urlPath, "/"))
	if clean == "/" {
		return "", false
	}

	name := filepath.Join(dir, clean)
	info, err := os.Stat(name)
	if err != nil || info.IsDir() {
		return "", false
	}
	return name, true
}
