package server

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// distFS holds the built frontend. `make web` fills dist/; a checkout
// without it only contains .gitkeep, and Static serves a hint page.
//
//go:embed all:dist
var distFS embed.FS

const notBuiltPage = `<!doctype html>
<meta charset="utf-8">
<title>clask</title>
<style>body{font:16px/1.5 system-ui,sans-serif;max-width:40rem;margin:4rem auto;padding:0 1rem}code{background:#eee;padding:.1em .3em;border-radius:3px}</style>
<h1>clask UI not built</h1>
<p>The API is running, but the frontend has not been compiled into this binary.</p>
<ul>
<li>Build it: <code>make web</code>, then rebuild clask.</li>
<li>Or develop live: <code>pnpm dev</code> in <code>web/</code> and open <a href="http://localhost:5173">http://localhost:5173</a>.</li>
</ul>
<p>Try <a href="/api/tasks">/api/tasks</a>.</p>
`

// Static serves the embedded frontend with single-page-app fallback.
func Static() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	files := http.FileServerFS(sub)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name != "" && name != "index.html" {
			if f, err := sub.Open(name); err == nil {
				f.Close()
				files.ServeHTTP(w, r)
				return
			}
		}
		if _, err := fs.Stat(sub, "index.html"); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(notBuiltPage))
			return
		}
		// Unknown extension-less paths and "/" get the app shell.
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		files.ServeHTTP(w, r2)
	})
}
