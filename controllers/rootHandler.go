package controllers

import "net/http"

func rootHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "test.html", nil)
}
