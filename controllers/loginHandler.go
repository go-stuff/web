package controllers

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-stuff/grpc/api"
	"github.com/go-stuff/ldap"
	"github.com/go-stuff/web/models"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/csrf"
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

		// update user and groups in mongo to use with permissions middleware
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		userSvc := api.NewUserServiceClient(apiClient)
		req := new(api.UserByUsernameReq)
		req.Username = user.Username
		foundRes, err := userSvc.ByUsername(ctx, req)
		if err != nil {
			if strings.Contains(err.Error(), mongo.ErrNoDocuments.Error()) {
				// If they dont exist, add them
				req := new(api.UserCreateReq)
				req.User = &api.User{
					ID:         primitive.NewObjectID().Hex(),
					Username:   user.Username,
					Groups:     user.Groups,
					CreatedBy:  "System",
					CreatedAt:  ptypes.TimestampNow(),
					ModifiedBy: "System",
					ModifiedAt: ptypes.TimestampNow(),
				}
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				_, err := userSvc.Create(ctx, req)
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
			req := new(api.UserUpdateReq)
			req.User = new(api.User)
			req.User.ID = foundRes.User.ID
			req.User.Username = user.Username
			req.User.Groups = user.Groups
			req.User.ModifiedBy = "System"
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_, err := userSvc.Update(ctx, req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// save the session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save > %v\n", err.Error())
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	}
}
