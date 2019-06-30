package controllers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"

	"github.com/go-stuff/grpc/api"
)

func routeSeed() error {
	// walk and get the routes and sort them
	routes = nil
	router.Walk(gorillaWalkFunc)
	sort.Strings(routes)

	// create a context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// gRPC role and route services
	roleSvc := api.NewRoleServiceClient(apiClient)
	routeSvc := api.NewRouteServiceClient(apiClient)

	// gRPC get all roles
	roleReq := new(api.RoleListReq)
	roleRes, err := roleSvc.List(ctx, roleReq)
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > routeSeed() > roleSvc.List(): %s\n", err.Error())
		return err
	}

	// get all routes
	routeReq := new(api.RouteListReq)
	routeRes, err := routeSvc.List(ctx, routeReq)
	if err != nil {
		log.Printf("ERROR > controllers/routeHandler.go > routeSeed() > routeSvc.List(): %s\n", err.Error())
		return err
	}

	// iterate over roles and routes and delete any routes for roles that do not exist
	for _, route := range routeRes.Routes {
		var found bool
		for _, role := range roleRes.Roles {
			if role.ID == route.RoleID {
				found = true
			}
		}
		if !found {
			deleteReq := new(api.RouteDeleteReq)
			deleteReq.ID = route.ID
			deleteRes, err := routeSvc.Delete(ctx, deleteReq)
			if err != nil {
				log.Printf("ERROR > controllers/routeHandler.go > routeSeed() > routeSvc.Delete(): %s\n", err.Error())
				return err
			}
			if deleteRes.Deleted > 0 {
				log.Printf("INFO > controllers/routeHandler.go > routeSeed(): - delete %v %v\n", route.RoleID, route.Path)
			}
		}
	}

	// iterate over roles and routes and create any that are missing
	for _, s := range routes {
		for _, role := range roleRes.Roles {
			var found bool
			for _, route := range routeRes.Routes {
				if role.ID == route.RoleID && s == route.Path {
					found = true
				}
			}
			if !found {
				updateReq := new(api.RouteUpdateByRoleIDAndPathReq)
				updateReq.RoleID = role.ID
				updateReq.Path = s

				if role.Name == "Admin" {
					updateReq.Permission = true
				}
				if role.Name == "Read Only" {
					if s == "/" || s == "/home" || s == "/server/list" {
						updateReq.Permission = true
					}
				}
				updateRes, err := routeSvc.UpdateByRoleIDAndPath(ctx, updateReq)
				if err != nil {
					log.Printf("ERROR > controllers/routeHandler.go > routeSeed() > routeSvc.UpdateByRoleIDAndPath(): %s\n", err.Error())
					return err
				}
				if updateRes.Updated > 0 {
					log.Printf("INFO > controllers/routeHandler.go > routeSeed(): - update %v %v\n", role.ID, s)
				}
			}
		}
	}

	return nil
}

func gorillaWalkFunc(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	pathTemplate, err := route.GetPathTemplate()
	if err != nil {
		return err
	}
	switch pathTemplate {
	case
		"/login",
		"/logout",
		"/noauth",
		"/static/":
		// do not add public routes to the list
	default:
		routes = append(routes, pathTemplate)
	}
	return nil
}

func routeListHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/routeHandler.go > routeListHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC role and route services
		roleSvc := api.NewRoleServiceClient(apiClient)
		routeSvc := api.NewRouteServiceClient(apiClient)

		// get all roles
		roleReq := new(api.RoleListReq)
		roleRes, err := roleSvc.List(ctx, roleReq)
		if err != nil {
			log.Printf("ERROR > controllers/rolesHandler.go > roleSvc.List(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get all routes
		routeReq := new(api.RouteListReq)
		routeRes, err := routeSvc.List(ctx, routeReq)
		if err != nil {
			log.Printf("ERROR > controllers/routesHandler.go > routeSvc.List(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render(w, r, "routeList.html",
			struct {
				CSRF         template.HTML
				Notification string
				Roles        []*api.Role
				Routes       []*api.Route
			}{
				CSRF:         csrf.TemplateField(r),
				Notification: "",
				Roles:        roleRes.Roles,
				Routes:       routeRes.Routes,
			},
		)

	case "POST":
		// parse form fields
		err := r.ParseForm()
		if err != nil {
			log.Printf("ERROR > controllers/routeUpdateHandler.go > r.ParseForm(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC route service
		routeSvc := api.NewRouteServiceClient(apiClient)

		// get all routes
		routeReq := new(api.RouteListReq)
		routeRes, err := routeSvc.List(ctx, routeReq)
		if err != nil {
			log.Printf("ERROR > controllers/routesHandler.go > routeSvc.List(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// range over hidden checkbox fields
		for _, route := range routeRes.Routes {
			cbp := fmt.Sprintf("%v", r.Form.Get("hidden"+route.ID))

			if cbp == "" {
				route.Permission = false
			} else {
				route.Permission = true
			}

			routeReq := new(api.RouteUpdateByRoleIDAndPathReq)
			routeReq.RoleID = route.RoleID
			routeReq.Path = route.Path
			routeReq.Permission = route.Permission
			routeReq.ModifiedBy = "System"

			routeRes, err := routeSvc.UpdateByRoleIDAndPath(ctx, routeReq)
			if err != nil {
				log.Printf("ERROR > controllers/routesHandler.go > routeSvc.UpdateByRoleIDAndPath(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("INFO > controllers/routesHandler.go > routeSvc.UpdateByRoleIDAndPath(): %v\n", routeRes.Updated)
		}

		http.Redirect(w, r, "/route/list", http.StatusSeeOther)
	}

	// save session
	err = store.Save(r, w, session)
	if err != nil {
		log.Printf("ERROR > controllers/routeHandler.go > routeListHandler() > sessions.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
