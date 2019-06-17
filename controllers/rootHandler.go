package controllers

import "net/http"

func rootHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "home.html", nil)
}
