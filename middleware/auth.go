package middleware

import (
	"log"
	"net/http"
	"strings"
	//"github.com/gorilla/mux"
)

// Auth middleware authenticates users
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only process files that are not in the /static/ folder and not the favicon,ico.
		if strings.Contains(r.RequestURI, "/static/") || strings.Contains(r.RequestURI, "/favicon.ico") {
			next.ServeHTTP(w, r)
			return
		}

		session, err := store.Get(r, "session")
		if err != nil {
			log.Printf("ERROR > middleware/auth.go > Auth() > store.Get(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("INFO > middleware/auth.go > Auth() > store.Get(): %v %v\n", session.ID, session.Values["username"])

		// If this is a new session redirect to the login screen.
		if session.IsNew && r.RequestURI != "/login" {
			log.Println("INFO > middleware/auth.go > Auth() > Redirect to /login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If a session exists and the logout uri was requested, expire the session.
		if session.IsNew == false && r.RequestURI == "/logout" {
			log.Println("INFO > middleware/auth.go > Auth() > /logout expired session")

			// Set MaxAge to -1 to delete the session.
			session.Options.MaxAge = -1

			// Save the session.
			err = store.Save(r, w, session)
			if err != nil {
				log.Printf("ERROR > middleware/auth.go > Auth() > sessions.Save(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
		return
	})
}
