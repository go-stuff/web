package controllers

import "net/http"

func noauthHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "noauth.html", nil)
}
