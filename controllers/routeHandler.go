package controllers

import (
	"context"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"

	"github.com/go-stuff/grpc/api"
)

func CompileRoutes() error {
	routes = nil
	router.Walk(gorillaWalkFunc)
	sort.Strings(routes)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	roleSvc := api.NewRoleServiceClient(apiClient)
	routeSvc := api.NewRouteServiceClient(apiClient)

	// get all roles
	roleReq := new(api.RoleListReq)
	roleRes, err := roleSvc.List(ctx, roleReq)
	if err != nil {
		log.Printf("controllers/rolesHandler.go > ERROR > roleSvc.List(): %s\n", err.Error())
		return err
	}

	// get all routes
	routeReq := new(api.RouteListReq)
	routeRes, err := routeSvc.List(ctx, routeReq)
	if err != nil {
		log.Printf("controllers/routesHandler.go > ERROR > routeSvc.List(): %s\n", err.Error())
		return err
	}

	// iterate over roles and routes
	for _, s := range routes {
		log.Printf("route: %v\n", s)

		for _, role := range roleRes.Roles {

			var found bool
			for _, route := range routeRes.Routes {
				if role.ID == route.RoleID && s == route.Path {
					found = true
				}
			}
			if found {
				log.Println("found")
			} else {
				log.Println("not found")

				uReq := new(api.RouteUpdateByRoleIDAndPathReq)
				uReq.Route = &api.Route{
					RoleID: role.ID,
					Path:   s,
				}
				uRes, err := routeSvc.UpdateByRoleIDAndPath(ctx, uReq)
				if err != nil {
					return err
				}
				log.Println(uRes.Updated)
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
		"/",
		"/home",
		"/login",
		"/logout",
		"/static/":
		// do not add public routes to the list
	default:
		routes = append(routes, pathTemplate)
	}
	return nil
}

func routeListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	roleSvc := api.NewRoleServiceClient(apiClient)
	routeSvc := api.NewRouteServiceClient(apiClient)

	// get all roles
	roleReq := new(api.RoleListReq)
	roleRes, err := roleSvc.List(ctx, roleReq)
	if err != nil {
		log.Printf("controllers/rolesHandler.go > ERROR > roleSvc.List(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get all routes
	routeReq := new(api.RouteListReq)
	routeRes, err := routeSvc.List(ctx, routeReq)
	if err != nil {
		log.Printf("controllers/routesHandler.go > ERROR > routeSvc.List(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render(w, r, "routeList.html",
		struct {
			Notification string
			Roles        []*api.Role
			Routes       []*api.Route
		}{
			Notification: "",
			Roles:        roleRes.Roles,
			Routes:       routeRes.Routes,
		},
	)
}

func routeUpdateHandler(w http.ResponseWriter, r *http.Request) {

}