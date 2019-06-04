package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-stuff/web/models"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func usersHandler(w http.ResponseWriter, r *http.Request) {

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// find all roles
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := client.Database(os.Getenv("MONGO_DB_NAME")).Collection("users").Find(ctx,
		bson.D{},
		&options.FindOptions{
			Sort: bson.D{
				{Key: "name", Value: 1}, // acending
				// { Key: "name", Value: -1}, // descending
			},
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// initialize a slice of roles
	var users []*models.User

	// itterate each document returned
	for cursor.Next(ctx) {
		//var result bson.M
		var user = new(models.User)
		err := cursor.Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// time is stored in UTC but we want to display local time
		user.CreatedAt = user.CreatedAt.Local()
		user.ModifiedAt = user.ModifiedAt.Local()

		// append the current role to the slice
		users = append(users, user)
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
		return
	}

	// get notifications if there are any
	notification, err := getNotification(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render(w, r, "users.html",
		struct {
			Notification string
			Users        []*models.User
		}{
			Notification: notification,
			Users:        users,
		},
	)
}
