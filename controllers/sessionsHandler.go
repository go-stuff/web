package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-stuff/web/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("controllers/sessionsHandler.go > INFO > session: %v\n", session.Values)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := client.Database("test").Collection("sessions").Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var arr []models.Session

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		cr := result["createdAt"].(primitive.DateTime)
		ex := result["expiresAt"].(primitive.DateTime)

		arr = append(arr, models.Session{
			Username:   result["username"].(string),
			RemoteAddr: result["remoteaddr"].(string),
			Host:       result["host"].(string),
			CreatedAt:  time.Unix(int64(cr)/1000, int64(cr)%1000*1000000).Format(time.UnixDate),
			ExpiresAt:  time.Unix(int64(ex)/1000, int64(ex)%1000*1000000).Format(time.UnixDate),
		})
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("arr: %v\n", arr)

	// Save the session.
	err = session.Save(r, w)
	if err != nil {
		log.Printf("controllers/sessionsHandler.go > ERROR > sessions.Save > %v\n", err.Error())
	}

	render(w, r, "sessions.html", arr)
}
