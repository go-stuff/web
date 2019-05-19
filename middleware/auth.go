package middleware

import (
	"log"
	"net/http"
	"strings"
)

// Auth middleware authenticates users
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("middleware/auth.go > INFO > Auth() > method: %v, %v\n", r.Method, r.RequestURI)

		// Only process files that are not in the /static/ folder and not the favicon,ico.
		if strings.Contains(r.RequestURI, "/static/") || strings.Contains(r.RequestURI, "/favicon.ico") {
			next.ServeHTTP(w, r)
			return
		}

		session, err := store.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("middleware/auth.go > INFO > Auth() > store.Get > %v %v\n", session.ID, session.Values)

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
				log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save > %v\n", err.Error())
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
