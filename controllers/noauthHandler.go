package controllers

import (
	"fmt"
	"log"
	"net/http"
	//"github.com/gorilla/mux"
)

func noauthHandler(w http.ResponseWriter, r *http.Request) {
	//currentRoute := mux.CurrentRoute(r)
	//pathTemplate, _ := currentRoute.GetPathTemplate()

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > middleware/Permissions.go > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	noauth := fmt.Sprintf("%v", session.Values["pathtemplate"])

	render(w, r, "noauth.html", noauth)
}
