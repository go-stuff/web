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
		log.Printf("ERROR > controllers/roleHandler.go > CompileRoutes() > roleSvc.List(): %s\n", err.Error())
		return err
	}

	// get all routes
	routeReq := new(api.RouteListReq)
	routeRes, err := routeSvc.List(ctx, routeReq)
	if err != nil {
		log.Printf("ERROR > controllers/routeHandler.go > CompileRoutes() > routeSvc.List(): %s\n", err.Error())
		return err
	}

	// iterate over roles and routes
	for _, s := range routes {
		log.Printf("INFO > controllers/routeHandler.go > CompileRoutes(): - %v\n", s)

		for _, role := range roleRes.Roles {

			var found bool
			for _, route := range routeRes.Routes {
				if role.ID == route.RoleID && s == route.Path {
					found = true
				}
			}
			if found {
				// log.Println("found")
			} else {
				// log.Println("not found")

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	roleSvc := api.NewRoleServiceClient(apiClient)
	routeSvc := api.NewRouteServiceClient(apiClient)

	switch r.Method {
	case "GET":

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

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//roleSvc := api.NewRoleServiceClient(apiClient)
		routeSvc := api.NewRouteServiceClient(apiClient)

		// // get all roles
		// roleReq := new(api.RoleListReq)
		// roleRes, err := roleSvc.List(ctx, roleReq)
		// if err != nil {
		// 	log.Printf("controllers/rolesHandler.go > ERROR > roleSvc.List(): %s\n", err.Error())
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// get all routes
		routeReq := new(api.RouteListReq)
		routeRes, err := routeSvc.List(ctx, routeReq)
		if err != nil {
			log.Printf("controllers/routesHandler.go > ERROR > routeSvc.List(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// range over hidden checkbox fields
		for _, route := range routeRes.Routes {
			cbp := fmt.Sprintf("%v", r.Form.Get("hidden"+route.ID))
			log.Printf("%v: %v\n", route.ID, cbp)
			if cbp == "" {
				route.Permission = false
			} else {
				route.Permission = true
			}

			routeReq := new(api.RouteUpdateByRoleIDAndPathReq)
			log.Printf("routeReq: %v\n", routeReq.Route)

			routeReq.Route = route

			// routeReq.Route.ID = route.ID
			// routeReq.Route.RoleID = route.RoleID
			// routeReq.Route.Path = route.Path
			// routeReq.Route.Permission = route.Permission

			routeRes, err := routeSvc.UpdateByRoleIDAndPath(ctx, routeReq)
			if err != nil {
				log.Printf("controllers/routesHandler.go > ERROR > routeSvc.List(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("update: %v", routeRes.Updated)

			//log.Printf("route: %v, permission: %v\n", route.Path, route.Permission)
		}

		// for i := 0; i < len(routeRes.Routes); i++ {
		// 	cbp := fmt.Sprintf("%v", r.Form.Get("hidden"+routeRes.Routes[i].ID))
		// 	log.Printf("hidden%v: %v\n", routeRes.Routes[i].ID, cbp)
		// 	if cbp == "" {
		// 		routeRes.Routes[i].Permission = false
		// 	} else {
		// 		routeRes.Routes[i].Permission = true
		// 	}
		// 	//log.Printf("route: %v, permission: %v\n", routeRes.Routes[i].Path, routeRes.Routes[i].Permission)
		// }

		http.Redirect(w, r, "/route/list", http.StatusSeeOther)
	}
}
