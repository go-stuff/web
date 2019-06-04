package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/go-stuff/web/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func rolesHandler(w http.ResponseWriter, r *http.Request) {

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// find all roles
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := client.Database(os.Getenv("MONGO_DB_NAME")).Collection("roles").Find(ctx,
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
	var roles []*models.Role

	// itterate each document returned
	for cursor.Next(ctx) {
		//var result bson.M
		var role = new(models.Role)
		err := cursor.Decode(&role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// time is stored in UTC but we want to display local time
		role.CreatedAt = role.CreatedAt.Local()
		role.ModifiedAt = role.ModifiedAt.Local()

		// append the current role to the slice
		roles = append(roles, role)
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

	render(w, r, "roles.html",
		struct {
			Notification string
			Roles        []*models.Role
		}{
			Notification: notification,
			Roles:        roles,
		},
	)
}

func roleCreateHandler(w http.ResponseWriter, r *http.Request) {

	// parse form fields
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "POST":

		// prepare a role to insert
		role := &models.Role{
			ID:          primitive.NewObjectID().Hex(),
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			CreatedBy:   session.Values["username"].(string),
			CreatedAt:   time.Now().UTC(),
			ModifiedBy:  session.Values["username"].(string),
			ModifiedAt:  time.Now().UTC(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// insert role into mongo
		_, err = client.Database(os.Getenv("MONGO_DB_NAME")).Collection("roles").InsertOne(ctx, role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was added
		addNotification(session, fmt.Sprintf("Role '%s' has been created!", role.Name))
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render or redirect
	switch r.Method {
	case "GET":
		render(w, r, "rolesUpsert.html",
			struct {
				Title  string
				Role   *models.Role
				Action string
			}{
				Title:  "Create Role",
				Role:   new(models.Role),
				Action: "Create",
			},
		)
	case "POST":
		http.Redirect(w, r, "/roles", http.StatusSeeOther)
	}
}

func roleReadHandler(w http.ResponseWriter, r *http.Request) {

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)

	// initialize a new role
	var role = new(models.Role)

	// find a role
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Database(os.Getenv("MONGO_DB_NAME")).Collection("roles").FindOne(ctx,
		bson.D{
			{Key: "_id", Value: vars["id"]},
		},
	).Decode(role)
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

	render(w, r, "rolesRead.html",
		struct {
			Role *models.Role
		}{

			Role: role,
		},
	)
}

func roleUpdateHandler(w http.ResponseWriter, r *http.Request) {

	// get variables from uri
	vars := mux.Vars(r)

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// initialize a new role
	var role = new(models.Role)

	// find a role
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Database(os.Getenv("MONGO_DB_NAME")).Collection("roles").FindOne(ctx,
		bson.D{
			{Key: "_id", Value: vars["id"]},
		},
	).Decode(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		// time is stored in UTC but we want to display local time
		role.CreatedAt = role.CreatedAt.Local()
		role.ModifiedAt = role.ModifiedAt.Local()

	case "POST":
		// update values on the form and modified fields
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := client.Database(os.Getenv("MONGO_DB_NAME")).Collection("roles").UpdateOne(ctx,
			bson.D{
				{Key: "_id", Value: vars["id"]},
			},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "name", Value: r.FormValue("name")},
					{Key: "description", Value: r.FormValue("description")},
					{Key: "modifiedBy", Value: session.Values["username"]},
					{Key: "modifiedAt", Value: time.Now().UTC()},
				}},
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was updated
		addNotification(session, fmt.Sprintf("Role '%s' has been updated!", r.FormValue("name")))
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		render(w, r, "rolesUpsert.html",
			struct {
				Title  string
				Role   *models.Role
				Action string
			}{
				Title:  "Update Role",
				Role:   role,
				Action: "Update",
			},
		)
	case "POST":
		http.Redirect(w, r, "/roles", http.StatusSeeOther)
	}

}

func roleDeleteHandler(w http.ResponseWriter, r *http.Request) {

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)

	// initialize a new role
	var role = new(models.Role)

	// find a role
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Database(os.Getenv("MONGO_DB_NAME")).Collection("roles").FindOne(ctx,
		bson.D{
			{Key: "_id", Value: vars["id"]},
		},
	).Decode(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// delete the ObjectID from roles
	_, err = client.Database(os.Getenv("MONGO_DB_NAME")).Collection("roles").DeleteOne(ctx,
		bson.D{
			{Key: "_id", Value: vars["id"]},
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// put a notification in the session.Values that a role was deleted
	addNotification(session, fmt.Sprintf("Role '%s' was deleted!", role.Name))

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/roles", http.StatusTemporaryRedirect)
}

// addNotification adds a notification message to session.Values
func addNotification(session *sessions.Session, notification string) {
	session.Values["notification"] = notification
}

// getNotification returns a notification from session.Values if
// one exists, otherwise it returns an empty string
// if a notification was returned, the notification session.Value
// is emptied
func getNotification(w http.ResponseWriter, r *http.Request) (string, error) {

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	var notification string

	if session.Values["notification"] == nil {
		notification = ""
	} else {
		notification = session.Values["notification"].(string)
	}

	session.Values["notification"] = ""

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	return notification, nil
}
