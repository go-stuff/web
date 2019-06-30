package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-stuff/grpc/api"
	"github.com/gorilla/mux"
)

// Permissions allows or denies access to routes
func Permissions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only process files that are not in the /static/ folder and not the favicon,ico.
		if strings.Contains(r.RequestURI, "/static/") || strings.Contains(r.RequestURI, "/favicon.ico") {
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("INFO > middleware/permission.go > Permissions() > method: %v, %v\n", r.Method, r.RequestURI)

		currentRoute := mux.CurrentRoute(r)
		pathTemplate, err := currentRoute.GetPathTemplate()
		if err != nil {
			log.Printf("ERROR > middleware/permission.go > currentRoute.GetPathTemplate(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("INFO > middleware/permission.go > Permissions() > pathTemplate: %s\n", pathTemplate)

		// get session
		session, err := store.Get(r, "session")
		if err != nil {
			log.Printf("ERROR > middleware/Permissions.go > store.Get(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if pathTemplate != "/noauth" &&
			pathTemplate != "/login" &&
			pathTemplate != "/logout" {

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			routeSvc := api.NewRouteServiceClient(apiClient)

			// use the api to find a role
			routeReq := new(api.RouteReadByRoleIDAndPathReq)

			if session.Values["roleid"] == nil || session.Values["roleid"] == "" {
				log.Println("ERROR > middleware/Permissions.go > no role")
				http.Error(w, "account has no role", http.StatusInternalServerError)
				return
			}

			roleid := fmt.Sprintf("%v", session.Values["roleid"])

			routeReq.Route = new(api.Route)
			routeReq.Route.RoleID = roleid
			routeReq.Route.Path = pathTemplate

			routeRes, err := routeSvc.ReadByRoleIDAndPath(ctx, routeReq)
			if err != nil {
				log.Printf("ERROR > controllers/permissions.go > permissionFM() > routeSvc.RouteReadByRoleIDAndPath(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			log.Printf("INFO > controllers/controllers.go > permissionFM() > pathTemplate = permission: %s = %v\n", pathTemplate, routeRes.Route.Permission)

			if routeRes.Route.Permission == false {
				log.Printf("WARN > middleware/permission.go > Permissions() > The role: %v has no permissions to route: %v\n", roleid, pathTemplate)

				session.Values["pathtemplate"] = pathTemplate
				// save session
				err = session.Save(r, w)
				if err != nil {
					log.Printf("ERROR > middleware/Permissions.go > session.Save(): %s\n", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, "/noauth", http.StatusTemporaryRedirect)
			}
		}

		// save session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("ERROR > middleware/Permissions.go > session.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the results of this http request to the next handler.
		next.ServeHTTP(w, r)
		return
	})
}
