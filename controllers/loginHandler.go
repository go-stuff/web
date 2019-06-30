package controllers

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-stuff/grpc/api"
	"github.com/go-stuff/ldap"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/csrf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
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
		// parse form fields
		err := r.ParseForm()
		if err != nil {
			log.Printf("ERROR > controllers/loginHandler.go > r.ParseForm(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// start a new session
		session, err := store.New(r, "session")
		if err != nil {
			log.Printf("ERROR > controllers/loginHandler.go > store.New(): %s\n", err.Error())
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

				// audit a login failure
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				auditSvc := api.NewAuditServiceClient(apiClient)

				auditReq := new(api.AuditCreateReq)
				auditReq.Audit = &api.Audit{
					ID:        primitive.NewObjectID().Hex(),
					Username:  fmt.Sprintf("%v", r.FormValue("username")),
					Action:    fmt.Sprintf("%v: %v", r.Method, r.URL),
					Session:   fmt.Sprintf("%v", err),
					CreatedBy: "System",
					CreatedAt: ptypes.TimestampNow(),
				}
				_, err = auditSvc.Create(ctx, auditReq)
				if err != nil {
					log.Printf("ERROR > controllers/loginHandler.go > auditSvc.Create(): %s\n", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

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

			// audit a login failure
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			auditSvc := api.NewAuditServiceClient(apiClient)

			auditReq := new(api.AuditCreateReq)
			auditReq.Audit = &api.Audit{
				ID:        primitive.NewObjectID().Hex(),
				Username:  fmt.Sprintf("%v", r.FormValue("username")),
				Action:    fmt.Sprintf("%v: %v", r.Method, r.URL),
				Session:   fmt.Sprintf("%v", errors.New("username not found")),
				CreatedBy: "System",
				CreatedAt: ptypes.TimestampNow(),
			}
			_, err = auditSvc.Create(ctx, auditReq)
			if err != nil {
				log.Printf("ERROR > controllers/loginHandler.go > auditSvc.Create(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

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

		roleSvc := api.NewRoleServiceClient(apiClient)
		userSvc := api.NewUserServiceClient(apiClient)

		userReq := new(api.UserReadByUsernameReq)
		userReq.Username = user.Username

		foundRes, err := userSvc.ReadByUsername(ctx, userReq)
		if err != nil {
			log.Printf("ERROR > controllers/loginHandler.go > userSvc.ReadByUsername(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if foundRes.User.ID != "" {
			// if they do exist, update their groups
			userReq := new(api.UserUpdateReq)
			//userReq.User = new(api.User)
			userReq.ID = foundRes.User.ID
			//userReq.Username = user.Username
			userReq.Groups = user.Groups
			userReq.ModifiedBy = "System"

			// update the sessions roleid
			userReq.RoleID = foundRes.User.RoleID
			session.Values["roleid"] = foundRes.User.RoleID

			// if user is in the admin ad group, give them admin permissions
			for _, group := range user.Groups {
				if group == os.Getenv("ADMIN_AD_GROUP") {
					readReq := new(api.RoleReadByNameReq)
					readReq.Name = "Admin"
					readRes, err := roleSvc.ReadByName(ctx, readReq)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					// update sessions roleid
					userReq.RoleID = readRes.Role.ID
					session.Values["roleid"] = readRes.Role.ID
				}

				_, err := userSvc.Update(ctx, userReq)
				if err != nil {
					log.Printf("controllers/loginHandler.go > ERROR > userSvc.Update(): %s\n", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			// if they don't exist add them
			userReq := new(api.UserCreateReq)

			userReq.Username = user.Username
			userReq.Groups = user.Groups
			userReq.RoleID = "No Role"
			userReq.CreatedBy = "System"

			// if user is in the admin ad group, give the user the admin roleid
			for _, group := range user.Groups {
				if group == os.Getenv("ADMIN_AD_GROUP") {
					log.Printf("%s", group)
					readReq := new(api.RoleReadByNameReq)
					readReq.Name = "Admin"
					readRes, err := roleSvc.ReadByName(ctx, readReq)
					if err != nil {
						log.Printf("ERROR > controllers/loginHandler.go > roleSvc.ReadByName(): %s\n", err.Error())
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					userReq.RoleID = readRes.Role.ID
					session.Values["roleid"] = readRes.Role.ID
				}
			}

			// if no role is given, give the default read only roleid
			if userReq.RoleID == "No Role" {
				readReq := new(api.RoleReadByNameReq)
				readReq.Name = "Read Only"
				readRes, err := roleSvc.ReadByName(ctx, readReq)
				if err != nil {
					log.Printf("ERROR > controllers/loginHandler.go > roleSvc.ReadByName(): %s\n", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				userReq.RoleID = readRes.Role.ID
				session.Values["roleid"] = readRes.Role.ID
			}

			_, err = userSvc.Create(ctx, userReq)
			if err != nil {
				log.Printf("ERROR > controllers/loginHandler.go > userSvc.Create(): %s\n", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// save the session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("ERROR > middleware/loginHandler.go > loginHandler() > sessions.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// audit a successful login
		auditSvc := api.NewAuditServiceClient(apiClient)

		auditReq := new(api.AuditCreateReq)
		auditReq.Audit = &api.Audit{
			ID:        primitive.NewObjectID().Hex(),
			Username:  fmt.Sprintf("%v", r.FormValue("username")),
			Action:    fmt.Sprintf("%v: %v", r.Method, r.URL),
			Session:   fmt.Sprintf("%v", session.Values),
			CreatedBy: "System",
			CreatedAt: ptypes.TimestampNow(),
		}
		_, err = auditSvc.Create(ctx, auditReq)
		if err != nil {
			log.Printf("ERROR > controllers/loginHandler.go > auditSvc.Create(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	}
}
