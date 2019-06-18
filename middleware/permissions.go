package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Permissions allows or denies access to routes
func Permissions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// init err so that errors from other handlers are not passed on to this middleware
		var err error

		// Only process files that are not in the /static/ folder and not the favicon,ico.
		if strings.Contains(r.RequestURI, "/static/") || strings.Contains(r.RequestURI, "/favicon.ico") {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("middleware/permission.go > INFO > Permissions() > method: %v, %v\n", r.Method, r.RequestURI)

		currentRoute := mux.CurrentRoute(r)
		pathTemplate, err := currentRoute.GetPathTemplate()
		if err != nil {
			log.Printf("middleware/permission.go > ERROR > currentRoute.GetPathTemplate(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("middleware/permission.go > INFO > Permissions() > pathTemplate: %s\n", pathTemplate)

		// Get the regex version of the RequestURI.
		// route := controllers.RouteToRegex(r.RequestURI)

		// //log.WithFields(log.Fields{"func": "ServeHTTP", "file": "middleware/permission.go"}).Debugf("User: %v, Role: %v, Route: %v", s.Values["UserID"], s.Values["RoleTypeID"], route)
		// if route != "/noauth" &&
		// 	route != "/login" &&
		// 	route != "/login/noauth" &&
		// 	route != "/logout" &&
		// 	route != "/en" &&
		// 	route != "/fr" {

		// 	session := getSession(r)
		// 	roleid, err := stringToUint(fmt.Sprintf("%v", session.Values["RoleID"]))
		// 	if err != nil {
		// 		//panic(err)
		// 		log.Fatalf("middleware/permission.go > WARN > Permissions() > %v\n", err.Error())
		// 	}
		// get session
		session, err := store.Get(r, "session")
		if err != nil {
			log.Printf("middleware/Permissions.go > ERROR > store.Get(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 	if controllers.PermissionsMap()[roleid][route] == false {
		// 		log.Printf("middleware/permission.go > WARN > Permissions() > The role: %v has no permissions to route: %v\n", roleid, route)
		// 		http.Redirect(w, r, "/noauth", http.StatusTemporaryRedirect)
		// 	}
		// }

		// save session
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the results of this http request to the next handler.
		next.ServeHTTP(w, r)
	})
}
