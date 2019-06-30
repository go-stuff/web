package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
)

// Headers middleware makes sure the files have proper content types.
func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only consider put, post and patch
		switch r.Method {
		case "PUT", "POST", "PATCH":
			path := r.URL.Path[1:]
			var contentType string

			if strings.HasSuffix(path, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(path, ".js") {
				contentType = "application/javascript"
			} else if strings.HasSuffix(path, ".ico") {
				contentType = "image/x-icon"
			} else if strings.HasSuffix(path, ".html") {
				contentType = "text/html"
			} else if strings.HasSuffix(path, ".png") {
				contentType = "image/png"
			} else if strings.HasSuffix(path, ".svg") {
				contentType = "image/svg+xml"
			} else {
				contentType = "text/plain"
			}

			// Add the Content-Type of our files.
			w.Header().Add("Content-Type", contentType)
			// Add X-Content-Type-Options header
			w.Header().Add("X-Content-Type-Options", "nosniff")
			// Add X-XSS-Protection HTTP response header allows the web server to enable or disable the web browser's XSS protection mechanism
			w.Header().Add("X-XSS-Protection", "1; mode=block")
			// Add Cache-Control and Pragma header
			w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate;")
			w.Header().Add("Pragma", "no-cache")
			// Prevent page from being displayed in an iframe
			w.Header().Add("X-Frame-Options", "DENY")

			// Get the token and pass it in the CSRF header. Our JSON-speaking client
			// or JavaScript framework can now read the header and return the token in
			// in its own "X-CSRF-Token" request header on the subsequent POST.
			w.Header().Set("X-CSRF-Token", csrf.Token(r))
		}

		next.ServeHTTP(w, r)
		return
	})
}
