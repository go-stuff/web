package controllers

import (
	"log"
	"net/http"

	"github.com/go-stuff/grpc/api"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("controllers/homeHandler.go > ERROR > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get data from session.Values
	user := &api.User{
		Username: session.Values["username"].(string),
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("controllers/homeHandler.go > ERROR > sessions.Save: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render to template
	render(w, r, "home.html",
		struct {
			User  *api.User
			Error error
		}{
			User:  user,
			Error: nil,
		})
}
