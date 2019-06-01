package controllers

import (
	"log"
	"net/http"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("controllers/testHandler.go > INFO > session: %v\n", session.Values)

	// Save the session.
	err = session.Save(r, w)
	if err != nil {
		log.Printf("controllers/testHandler.go > ERROR > sessions.Save > %v\n", err.Error())
	}

	render(w, r, "test.html", nil)
}
