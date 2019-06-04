package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-stuff/ldap"
	"github.com/go-stuff/web/models"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// parse form fields
	err := r.ParseForm()
	if err != nil {
		log.Printf("controllers/loginHandler.go > ERROR > r.ParseForm(): %v\n", err.Error())
	}

	switch r.Method {
	case "GET":
		render(w, r, "login.html", nil)

	case "POST":
		// start a new session
		session, err := store.New(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// this is just an example, you can swap out authentication
		// with AD, LDAP, oAuth, etc...
		authenticatedUser := make(map[string]string)
		authenticatedUser["test"] = "test"
		authenticatedUser["user1"] = "password"
		authenticatedUser["user2"] = "password"
		authenticatedUser["user3"] = "password"

		user := models.User{}

		var found bool
		for k, v := range authenticatedUser {
			if r.FormValue("username") == k && r.FormValue("password") == v {
				user.Username = k
				found = true
			}
		}

		// if local account was not found check ldap
		if !found {
			username, groups, err := ldap.Auth(
				"192.168.1.100",               // LDAP_SERVER
				"636",                         // LDAP_PORT
				"cn=admin,dc=go-stuff,dc=ca",  // LDAP_BIND_DN
				"password",                    // LDAP_BIND_PASS
				"ou=people,dc=go-stuff,dc=ca", // LDAP_USER_BASE_DN
				"uid",                         // LDAP_USER_SEARCH_ATTR
				"ou=group,dc=go-stuff,dc=ca",  // LDAP_GROUP_BASE_DN
				"posixGroup",                  // LDAP_GROUP_OBJECT_CLASS
				"memberUid",                   // LDAP_GROUP_SEARCH_ATTR
				"false",                       // LDAP_GROUP_SEARCH_FULL
				r.FormValue("username"),       // Username
				r.FormValue("password"),       // Password
			)
			if err != nil {
				render(w, r, "login.html",
					struct {
						Username string
						Error    error
					}{
						Username: r.FormValue("username"),
						Error:    err,
					})
				return
			}

			user.Username = username
			user.Groups = groups

			found = true
		}

		// user not found
		if !found {
			render(w, r, "login.html",
				struct {
					Username string
					Error    error
				}{
					Username: r.FormValue("username"),
					Error:    errors.New("username not found"),
				})
			return
		}

		// add important values to the session
		session.Values["remoteaddr"] = r.RemoteAddr
		session.Values["host"] = r.Host
		session.Values["username"] = user.Username

		// save the session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save > %v\n", err.Error())
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	}
}
