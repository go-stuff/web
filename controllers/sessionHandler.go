package controllers

import (
	"context"
	"log"
	"net/http"

	"time"

	"github.com/go-stuff/grpc/api"
)

func sessionListHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/sessionsHandler.go > sessionListHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
		// display session
		log.Printf("INFO > controllers/sessionsHandler.go > sessionListHandler() > session: %v %v\n", session.Values["_id"], session.Values["username"])

		// call api to get a slice of sessions
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		sessionSvc := api.NewSessionServiceClient(apiClient)

		sessionReq := new(api.SessionListReq)
		sessionRes, err := sessionSvc.List(ctx, sessionReq)
		if err != nil {
			log.Printf("ERROR > controllers/sessionsHandler.go > sessionListHandler() > sessionSvc.List(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render(w, r, "sessionList.html",
			struct {
				Sessions []*api.Session
			}{
				Sessions: sessionRes.Sessions,
			},
		)
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("ERROR > controllers/sessionsHandler.go > sessionListHandler() > session.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
