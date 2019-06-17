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

func routesHandler(w http.ResponseWriter, r *http.Request) {
	routes = nil
	router.Walk(gorillaWalkFunc)
	sort.Strings(routes)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	roleSvc := api.NewRoleServiceClient(apiClient)
	routeSvc := api.NewRouteServiceClient(apiClient)

	// get all roles
	roleReq := new(api.RoleSliceReq)
	roleSlice, err := roleSvc.Slice(ctx, roleReq)
	if err != nil {
		log.Printf("controllers/rolesHandler.go > ERROR > svc.Slice(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get all routes
	routeReq := new(api.RouteSliceReq)
	routeSlice, err := routeSvc.Slice(ctx, routeReq)
	if err != nil {
		log.Printf("controllers/routesHandler.go > ERROR > svc.Slice(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, s := range routes {
		log.Printf("route: %v\n", s)

		for _, role := range roleSlice.Roles {

			var found bool
			for _, route := range routeSlice.Routes {
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
					//ID:   primitive.NewObjectID().Hex(),
					RoleID: role.ID,
					Path:   s,
					//CreatedBy:  "System",
					//CreatedAt:  ptypes.TimestampNow(),
					//ModifiedBy: "System",
					//ModifiedAt: ptypes.TimestampNow(),
				}
				uRes, err := routeSvc.UpdateByRoleIDAndPath(ctx, uReq)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println(uRes.Updated)

				// createReq := new(api.RouteCreateReq)
				// createReq.Route = &api.Route{
				// 	ID:         primitive.NewObjectID().Hex(),
				// 	Name:       s,
				// 	CreatedBy:  "System",
				// 	CreatedAt:  ptypes.TimestampNow(),
				// 	ModifiedBy: "System",
				// 	ModifiedAt: ptypes.TimestampNow(),
				// }
				// _, err := svc.Create(ctx, createReq)
				// if err != nil {
				// 	http.Error(w, err.Error(), http.StatusInternalServerError)
				// 	return
				// }
			}
		}
	}
	// readByNameReq := new(api.RouteReadByNameReq)
	// readByNameReq.Name = "/roles"
	// readByNameRes, err := svc.ReadByName(ctx, readByNameReq)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	//log.Printf("roles: %v\n", readByNameRes)

	routeSlice, err = routeSvc.Slice(ctx, routeReq)
	if err != nil {
		log.Printf("controllers/routesHandler.go > ERROR > svc.Slice(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render(w, r, "routes.html",
		struct {
			Notification string
			Roles        []*api.Role
			Routes       []*api.Route
		}{
			Notification: "",
			Roles:        roleSlice.Roles,
			Routes:       routeSlice.Routes,
		},
	)
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
