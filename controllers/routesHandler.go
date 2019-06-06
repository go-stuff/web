package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func routesHandler(w http.ResponseWriter, r *http.Request) {

	router.Walk(gorillaWalkFunc)

	render(w, r, "routes.html",
		struct {
			Notification string
			Routes       []string
		}{
			Notification: "",
			Routes:       routes,
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
