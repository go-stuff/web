package controllers

import (
	"log"
	"net/http"

	"github.com/go-stuff/web/models"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get data from session.Values
	user := &models.User{
		Username: session.Values["username"].(string),
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("controllers/homeHandler.go > ERROR > sessions.Save > %v\n", err.Error())
	}

	// render to template
	render(w, r, "home.html",
		struct {
			User  *models.User
			Error error
		}{
			User:  user,
			Error: nil,
		})
}
