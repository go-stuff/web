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
		log.Printf("controllers/sessionsHandler.go > ERROR > sessionListHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// display session
	log.Printf("controllers/sessionsHandler.go > INFO > sessionListHandler() > session: %s %s\n", session.Values["_id"].(string), session.Values["username"].(string))

	// call api to get a slice of sessions
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sessionSvc := api.NewSessionServiceClient(apiClient)

	sessionReq := new(api.SessionListReq)
	sessionRes, err := sessionSvc.List(ctx, sessionReq)
	if err != nil {
		log.Printf("controllers/sessionsHandler.go > ERROR > sessionListHandler() > sessionSvc.List(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// // initialize a slice of sessions
	// var sessions []*models.Session

	// // find all sessions
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	// cursor, err := client.Database(os.Getenv("MONGO_DB_NAME")).Collection("sessions").Find(ctx, bson.D{})
	// if err != nil {
	// 	log.Printf("controllers/sessionsHandler.go > ERROR > client.Database(): %s\n", err.Error())
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// defer cursor.Close(ctx)

	// // itterate each document returned
	// for cursor.Next(ctx) {
	// 	var session = new(models.Session)
	// 	err := cursor.Decode(&session)
	// 	if err != nil {
	// 		log.Printf("controllers/sessionsHandler.go > ERROR > cursor.Decode(): %s\n", err.Error())
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	// get local time from UTC dates
	// 	session.CreatedAt = session.CreatedAt.Local()
	// 	session.ExpiresAt = session.ExpiresAt.Local()

	// 	// append result to slice
	// 	sessions = append(sessions, session)
	// }

	// // handle any errors with the cursor
	// if err := cursor.Err(); err != nil {
	// 	log.Printf("controllers/sessionsHandler.go > ERROR > cursor.Err(): %s\n", err.Error())
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("controllers/sessionsHandler.go > ERROR > sessionListHandler() > session.Save(): %s\n", err.Error())
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
