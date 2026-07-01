// Package dashboard serves the cluster dashboard single-page application.
//
// The SPA assets are embedded at build time via go:embed. Until the real
// build is copied in (see the "dashboard" Makefile target), a small
// placeholder index.html is embedded so the package always compiles.
package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed static
var staticFS embed.FS

// Handler returns an http.Handler that serves the embedded SPA assets with
// client-side routing support: any request path that does not correspond to a
// real embedded file is served index.html so the SPA router can take over.
func Handler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		// This can only happen if the embedded tree is malformed, which is a
		// build-time invariant, so failing loudly is appropriate.
		panic("dashboard: embed static: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if reqPath == "" {
			reqPath = "index.html"
		}

		if _, err := fs.Stat(sub, reqPath); err != nil {
			// Not a real file: fall back to index.html for SPA routing.
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
