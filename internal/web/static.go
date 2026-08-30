// Package web serves the built frontend next to the API.
//
// The frontend is a Vite single-page build. It used to be a Next.js app with
// its own Node server, which meant two processes and a proxy hop between them
// just to keep the session cookie same-origin. Serving the static build from
// this binary collapses that: one origin, one deploy, and nothing to configure
// for the cookie to work.
//
// Enabled by WEB_DIST_DIR. Unset, the API behaves exactly as it did before,
// which is what you want when running the backend alone against `vite dev`.
package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Handler serves files from dir, falling back to index.html.
//
// The fallback is what makes deep links work. A request for /dashboard has no
// file behind it — the route only exists once React Router is running — so the
// shell is served and the router resolves the URL in the browser. Without it,
// every refresh on a real page would 404.
func Handler(dir string) http.Handler {
	index := filepath.Join(dir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			// A non-GET that reached the fallback is a call to an endpoint
			// that does not exist, not a page request. Serving it the shell
			// with a 200 would tell a broken client everything is fine.
			http.NotFound(w, r)
			return
		}

		if name, ok := resolve(dir, r.URL.Path); ok {
			// Vite fingerprints everything under /assets, so those URLs can
			// never go stale and are worth caching hard. index.html is the
			// opposite: it names the current bundles, and caching it is how a
			// deploy leaves people on the old one.
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

// resolve maps a URL path to a file inside dir, or reports that there is none.
//
// The join is done on a cleaned, rooted path so that a request for
// /../../etc/passwd cannot climb out of the build directory.
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
