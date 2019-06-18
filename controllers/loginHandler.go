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
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/csrf"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// parse form fields
	err := r.ParseForm()
	if err != nil {
		log.Printf("controllers/loginHandler.go > ERROR > r.ParseForm(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
			log.Printf("controllers/loginHandler.go > ERROR > store.New(): %s\n", err.Error())
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

		user := api.User{}

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

		userReq := new(api.UserReadByUsernameReq)
		userReq.Username = user.Username

		foundRes, err := userSvc.ReadByUsername(ctx, userReq)
		if err != nil {
			if strings.Contains(err.Error(), mongo.ErrNoDocuments.Error()) {
				// If they dont exist, add them
				userReq := new(api.UserCreateReq)
				userReq.User = &api.User{
					ID:         primitive.NewObjectID().Hex(),
					Username:   user.Username,
					Groups:     user.Groups,
					RoleID:     user.RoleID,
					CreatedBy:  "System",
					CreatedAt:  ptypes.TimestampNow(),
					ModifiedBy: "System",
					ModifiedAt: ptypes.TimestampNow(),
				}

				_, err := userSvc.Create(ctx, userReq)
				if err != nil {
					log.Printf("controllers/loginHandler.go > ERROR > userSvc.Create(): %s\n", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				log.Printf("controllers/loginHandler.go > ERROR > userSvc.ByUsername(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// If they do exist, update their groups
			userReq := new(api.UserUpdateReq)
			userReq.User = new(api.User)
			userReq.User.ID = foundRes.User.ID
			userReq.User.Username = user.Username
			userReq.User.Groups = user.Groups
			userReq.User.RoleID = foundRes.User.RoleID
			userReq.User.ModifiedBy = "System"

			_, err := userSvc.Update(ctx, userReq)
			if err != nil {
				log.Printf("controllers/loginHandler.go > ERROR > userSvc.Update(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			session.Values["roleid"] = foundRes.User.RoleID
		}

		// save the session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("middleware/auth.go > ERROR > Auth() > sessions.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	}
}
