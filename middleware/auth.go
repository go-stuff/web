package middleware

import (
	"log"
	"net/http"
	"strings"
)

// Auth is middleware to handle user authentication.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("middleware/auth.go > INFO > Auth() > method: %v, %v\n", r.Method, r.RequestURI)

		// Only process files that are not in the /static/ folder and not the favicon,ico.
		if strings.Contains(r.RequestURI, "/static/") || strings.Contains(r.RequestURI, "/favicon.ico") {
			next.ServeHTTP(w, r)
			return
		}

		// var err error
		// var session *sessions.Session

		session, err := store.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("middleware/auth.go > INFO > Auth() > g.Store.Get > %v %v\n", session.ID, session.Values)

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

		// Create a session.
		// if session.IsNew {
		// 	session, err = g.Store.New(r, "session")
		// 	if err != nil {
		// 		log.Printf("middleware/auth.go > INFO > Auth() > g.Store.New: %v\n", err.Error())
		// 	}
		// 	err = session.Save(r, w)
		// 	if err != nil {
		// 		log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save > %v\n", err.Error())
		// 	}
		// }
		// log.Printf("middleware/auth.go > INFO > Auth() > session.IsNew > %v %v\n", session.ID, session.Values)

		// log.Printf("middleware/auth.go > INFO > Auth() > g.Store.Get > %v\n", session)
		// spew.Dump(session.Values)
		// spew.Dump(session.ID)

		// If there is a POST to the /login route.
		// if r.Method == "POST" && r.RequestURI == "/login" {
		// 	log.Printf("middleware/auth.go > INFO > Auth() > method: %v, %v\n", r.Method, r.RequestURI)

		// 	// Parse the /login form fields.
		// 	err = r.ParseForm()
		// 	if err != nil {
		// 		log.Printf("middleware/auth.go > ERROR > Auth() > POST/Login > %v\n", err.Error())
		// 		next.ServeHTTP(w, r)
		// 		return
		// 	}

		// 	user := &ldap.User{}
		// 	if r.FormValue("username") == "test" && r.FormValue("password") == "test" {
		// 		user.Username = "test"
		// 	} else {
		// 		// Authenticate with ldap.
		// 		user, err = ldap.Auth("svc-goldap", "g0L@ngLd@p", r.FormValue("username"), r.FormValue("password"))
		// 		if err != nil {
		// 			log.Printf("middleware/auth.go > ERROR > Auth() > ldap.Auth > %v\n", err.Error())
		// 			// "unable to connect to ldap"
		// 			// "unable to bind to ldap"
		// 			// "user not found"
		// 			// "user does not exist"
		// 			// "too many users returned"
		// 			// "authentication failed"

		// 			render(w, r, "login.html",
		// 				struct {
		// 					Username string
		// 					Error    error
		// 				}{
		// 					Username: r.FormValue("username"),
		// 					Error:    err,
		// 				},
		// 			)
		// 			return
		// 		}
		// 	}
		// 	// Session is authenticated, start the session.
		// 	session, err := g.Store.New(r, "session")
		// 	if err != nil {
		// 		log.Printf("middleware/auth.go > INFO > Auth() > g.Store.New: %v\n", err.Error())
		// 	}
		// 	log.Printf("middleware/auth.go > INFO > Auth() > Authenticated Session: %v\n", user.Username)

		// 	//session.AddFlash("username", user.Username)
		// 	// Add username to the session.
		// 	session.Values["username"] = user.Username

		// 	// Save the session.
		// 	err = session.Save(r, w)
		// 	if err != nil {
		// 		log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save > %v\n", err.Error())
		// 	}

		// 	// Redirect user to their home url.
		// 	http.Redirect(w, r, "/home", http.StatusSeeOther)
		// 	return
		// }

		// Save the session.
		// err = session.Save(r, w)
		// if err != nil {
		// 	log.Printf("middleware/auth.go > ERROR > Auth() > session.Save > %v\n", err.Error())
		// }

		next.ServeHTTP(w, r)
	})
}
