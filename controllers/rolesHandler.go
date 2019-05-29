package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	// initialize a slice of roles
	var roles []*models.Role

	options := options.FindOptions{}

	// sort by name
	options.Sort = bson.D{
		{Key: "name", Value: 1}, // acending
		// { Key: "name", Value: -1}, // descending
	}

	// Limit by 100 documents only
	// limit := int64(100)
	// options.Limit = &limit

	// find all roles
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := client.Database("test").Collection("roles").Find(ctx,
		bson.D{},
		&options,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

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

		roles = append(roles, role)
		// assert datetime values
		//cr := result["createdAt"].(primitive.DateTime)
		//mo := result["modifiedAt"].(primitive.DateTime)

		// append result to slice
		// roles = append(roles,
		// 	models.Role{
		// 		ID:          result["_id"].(primitive.ObjectID),
		// 		Name:        result["name"].(string),
		// 		Description: result["description"].(string),
		// 		CreatedBy:   result["createdBy"].(string),
		// 		CreatedAt:   primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond)),
		// 		//CreatedAt:   time.Unix(int64(cr)/1000, int64(cr)%1000*1000000).Format(time.UnixDate),
		// 		ModifiedBy: result["modifiedBy"].(string),
		// 		ModifiedAt: primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond)),
		// 		//ModifiedAt:  time.Unix(int64(mo)/1000, int64(mo)%1000*1000000).Format(time.UnixDate),
		// 	},
		// )
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

func rolesCreateHandler(w http.ResponseWriter, r *http.Request) {

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
		role := &models.Role{
			ID:          primitive.NewObjectID().Hex(),
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			CreatedBy:   session.Values["username"].(string),
			CreatedAt:   time.Now().UTC(), //primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond)),
			ModifiedBy:  session.Values["username"].(string),
			ModifiedAt:  time.Now().UTC(), //primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond)),
		}

		// load session.Values into a bson.D object
		// var insert bson.D

		// insert = append(insert, bson.E{Key: "name", Value: r.FormValue("name")})
		// insert = append(insert, bson.E{Key: "description", Value: r.FormValue("description")})

		// insert = append(insert, bson.E{Key: "createdBy", Value: session.Values["username"]})
		// insert = append(insert, bson.E{Key: "createdAt", Value: primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond))})
		// insert = append(insert, bson.E{Key: "modifiedBy", Value: session.Values["username"]})
		// insert = append(insert, bson.E{Key: "modifiedAt", Value: primitive.DateTime(time.Now().Truncate(time.Millisecond).UnixNano() / int64(time.Millisecond))})

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// insert session.Values into mongo and get the returned ObjectID
		_, err = client.Database("test").Collection("roles").InsertOne(ctx,
			role,
			// bson.D{
			// 	{Key: "_id", Value: primitive.NewObjectID().Hex()},
			// 	{Key: "name", Value: r.FormValue("name")},
			// 	{Key: "description", Value: r.FormValue("description")},
			// 	{Key: "createdBy", Value: session.Values["username"].(string)},
			// 	{Key: "createdAt", Value: time.Now().UTC()},
			// 	{Key: "modifiedBy", Value: session.Values["modifiedBy"].(string)},
			// 	{Key: "modifiedAt", Value: time.Now().UTC()},
			// },
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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

func rolesUpdateHandler(w http.ResponseWriter, r *http.Request) {

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
	err = client.Database("test").Collection("roles").FindOne(ctx,
		bson.D{
			{Key: "_id", Value: vars["id"]},
		},
	).Decode(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("role: %v\n", role)

	log.Printf("role: %v\n", role)

	switch r.Method {
	case "GET":
		// time is stored in UTC but we want to display local time
		role.CreatedAt = role.CreatedAt.Local()
		role.ModifiedAt = role.ModifiedAt.Local()

	case "POST":
		// role.Name = r.FormValue("name")
		// role.Description = r.FormValue("description")
		// role.ModifiedAt = time.Now().UTC()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := client.Database("test").Collection("roles").UpdateOne(ctx,
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

func rolesDeleteHandler(w http.ResponseWriter, r *http.Request) {

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
	err = client.Database("test").Collection("roles").FindOne(ctx,
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
	_, err = client.Database("test").Collection("roles").DeleteOne(ctx,
		bson.D{
			{Key: "_id", Value: vars["id"]},
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addNotification(session, fmt.Sprintf("Role '%s' was deleted!", role.Name))

	// save session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/roles", http.StatusTemporaryRedirect)
}

// objectIDToHex converts an ObjectID string to ObjectID Hex string
// example: ObjectID("5cedc3faf23dd22dcf869789") to 5cedc3faf23dd22dcf869789
// func objectIDToHex(id string) string {
// 	hexID := strings.TrimPrefix(id, "ObjectID(\"")
// 	hexID = strings.TrimSuffix(hexID, "\")")
// 	return hexID
// }

// func formatDateTime(dt primitive.DateTime) string {
// 	return time.Unix(int64(dt)/1000, int64(dt)%1000*1000000).Format(time.UnixDate)
// }

func addNotification(session *sessions.Session, notification string) {
	session.Values["notification"] = notification
}

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
