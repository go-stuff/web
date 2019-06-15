package controllers

import (
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

func serversHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render to template
	render(w, r, "servers.html", nil)
}

func serverCreateHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render to template
	render(w, r, "serversUpsert.html",
		struct {
			CSRF  template.HTML
			Title string
			// Role   *models.Role
			Action string
		}{
			CSRF:  csrf.TemplateField(r),
			Title: "Create Server",
			// Role:   new(models.Role),
			Action: "Create",
		},
	)
}
