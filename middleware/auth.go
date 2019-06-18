package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// Auth middleware authenticates users
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// init err so that errors from other handlers are not passed on to this middleware
		var err error

		path := r.URL.Path[1:]
		var contentType string

		//log.Debug(path)

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

		//log.Printf("middleware/header.go > INFO > Header() > method: %v, %v, content type: %v\n", r.Method, r.RequestURI, contentType)

		//log.Debug(contentType)

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

		currentRoute := mux.CurrentRoute(r)
		pathTemplate, _ := currentRoute.GetPathTemplate()

		_ = pathTemplate
		// get variables from uri
		//vars := mux.Vars(r)
		//log.Printf("\n\nr.RequestURI: %v\npathTemplate: %v\n\n", r.RequestURI, pathTemplate)

		//log.Printf("middleware/auth.go > INFO > Auth() > method: %v, %v\n", r.Method, r.RequestURI)

		// Only process files that are not in the /static/ folder and not the favicon,ico.
		if strings.Contains(r.RequestURI, "/static/") || strings.Contains(r.RequestURI, "/favicon.ico") {
			next.ServeHTTP(w, r)
			return
		}

		session, err := store.Get(r, "session")
		if err != nil {
			log.Printf("middleware/auth.go > ERROR > Auth() > store.Get(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("middleware/auth.go > INFO > Auth() > store.Get(): %v %v\n", session.ID, session.Values["username"])

		// If this is a new session redirect to the login screen.
		if session.IsNew && r.RequestURI != "/login" {
			log.Println("middleware/auth.go > INFO > Auth() > Redirect to /login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If a session exists and the logout uri was requested, expire the session.
		if session.IsNew == false && r.RequestURI == "/logout" {
			log.Println("middleware/auth.go > INFO > Auth() > /logout expire session")

			// Set MaxAge to -1 to delete the session.
			session.Options.MaxAge = -1

			// Save the session.
			err = store.Save(r, w, session)
			if err != nil {
				log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
