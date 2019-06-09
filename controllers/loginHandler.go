package controllers

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-stuff/ldap"
	"github.com/go-stuff/web/models"
	"github.com/gorilla/csrf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// parse form fields
	err := r.ParseForm()
	if err != nil {
		log.Printf("controllers/loginHandler.go > ERROR > r.ParseForm(): %v\n", err.Error())
	}

	switch r.Method {
	case "GET":
		render(w, r, "login.html",
			struct {
				CSRF     template.HTML
				Username string
				Error    error
			}{
				CSRF:     csrf.TemplateField(r),
				Username: r.FormValue("username"),
				Error:    nil,
			})

	case "POST":
		// start a new session
		session, err := store.New(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// this is just an example, you can swap out authentication
		// with AD, LDAP, oAuth, etc...
		authenticatedUser := make(map[string]string)
		authenticatedUser["test"] = "test"
		authenticatedUser["user1"] = "password"
		authenticatedUser["user2"] = "password"
		authenticatedUser["user3"] = "password"

		user := models.User{}

		var found bool
		for k, v := range authenticatedUser {
			if r.FormValue("username") == k && r.FormValue("password") == v {
				user.Username = k
				found = true
			}
		}

		// if local account was not found check ldap
		if !found {
			username, groups, err := ldap.Auth(
				os.Getenv("LDAP_SERVER"),
				os.Getenv("LDAP_PORT"),
				os.Getenv("LDAP_BIND_DN"),
				os.Getenv("LDAP_BIND_PASS"),
				os.Getenv("LDAP_USER_BASE_DN"),
				os.Getenv("LDAP_USER_SEARCH_ATTR"),
				os.Getenv("LDAP_GROUP_BASE_DN"),
				os.Getenv("LDAP_GROUP_OBJECT_CLASS"),
				os.Getenv("LDAP_GROUP_SEARCH_ATTR"),
				os.Getenv("LDAP_GROUP_SEARCH_FULL"),
				r.FormValue("username"),
				r.FormValue("password"),
			)
			if err != nil {
				render(w, r, "login.html",
					struct {
						CSRF     template.HTML
						Username string
						Error    error
					}{
						CSRF:     csrf.TemplateField(r),
						Username: r.FormValue("username"),
						Error:    err,
					})
				return
			}

			user.Username = username
			user.Groups = groups

			found = true
		}

		// user not found
		if !found {
			render(w, r, "login.html",
				struct {
					CSRF     template.HTML
					Username string
					Error    error
				}{
					CSRF:     csrf.TemplateField(r),
					Username: r.FormValue("username"),
					Error:    errors.New("username not found"),
				})
			return
		}

		// add important values to the session
		session.Values["remoteaddr"] = r.RemoteAddr
		session.Values["host"] = r.Host
		session.Values["username"] = user.Username

		// save the session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save > %v\n", err.Error())
		}

		// update user and groups in mongo to use with permissions middleware

		// initialize a new role
		var findUser = new(models.User)

		// find a role
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = client.Database(os.Getenv("MONGO_DB_NAME")).Collection("users").FindOne(ctx,
			bson.D{
				{Key: "username", Value: user.Username},
			},
		).Decode(findUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// If they dont exist, add them

				// prepare a role to insert
				addUser := &models.User{
					ID:         primitive.NewObjectID().Hex(),
					Username:   user.Username,
					Groups:     user.Groups,
					CreatedBy:  "System",
					CreatedAt:  time.Now().UTC(),
					ModifiedBy: "System",
					ModifiedAt: time.Now().UTC(),
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				// insert role into mongo
				_, err = client.Database(os.Getenv("MONGO_DB_NAME")).Collection("users").InsertOne(ctx, addUser)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// If they do exist, update their groups
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_, err := client.Database(os.Getenv("MONGO_DB_NAME")).Collection("users").UpdateOne(ctx,
				bson.D{
					{Key: "username", Value: user.Username},
				},
				bson.D{
					{Key: "$set", Value: bson.D{
						{Key: "groups", Value: user.Groups},
						{Key: "modifiedBy", Value: "System"},
						{Key: "modifiedAt", Value: time.Now().UTC()},
					}},
				},
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	}
}
