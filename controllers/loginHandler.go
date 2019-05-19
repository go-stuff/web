package controllers

import (
	"log"
	"errors"
	"net/http"

	"github.com/go-stuff/web/models"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// parse form fields
	err := r.ParseForm()
	if err != nil {
		log.Printf("controllers/loginHandler.go > ERROR > r.ParseForm(): %v\n", err.Error())
	}

	switch r.Method {
	case "GET":
		render(w, r, "login.html", nil)

	case "POST":
		// start a new session
		session, err := store.New(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := models.User{}

		// authenticate login
		if r.FormValue("username") != "test" && r.FormValue("password") != "test" {
			render(w, r, "login.html",
			struct {
				Username string
				Error    error
			}{
				Username: r.FormValue("username"),
				Error:    errors.New("username not found"),
			})
			return
		}

		user.Username = "test"

		// add username to the session
		session.Values["username"] = user.Username

		// save the session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save > %v\n", err.Error())
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	}
}
