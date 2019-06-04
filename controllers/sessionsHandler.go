package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-stuff/web/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("controllers/sessionsHandler.go > INFO > session: %v\n", session.Values["_id"])

	// initialize a slice of sessions
	var sessions []models.Session

	// find all sessions
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := client.Database(os.Getenv("MONGO_DB_NAME")).Collection("sessions").Find(ctx, bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// itterate each document returned
	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// assert datetime values
		cr := result["createdAt"].(primitive.DateTime)
		ex := result["expiresAt"].(primitive.DateTime)

		// append result to slice
		sessions = append(sessions,
			models.Session{
				Username:   result["username"].(string),
				RemoteAddr: result["remoteaddr"].(string),
				Host:       result["host"].(string),
				CreatedAt:  time.Unix(int64(cr)/1000, int64(cr)%1000*1000000).Format(time.UnixDate),
				ExpiresAt:  time.Unix(int64(ex)/1000, int64(ex)%1000*1000000).Format(time.UnixDate),
			},
		)
	}

	// handle any errors with the cursor
	if err := cursor.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	render(w, r, "sessions.html", sessions)
}
