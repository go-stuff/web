package controllers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-stuff/grpc/api"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func userSeed() {
	// TODO IF A ROLE IS REMOVED, Set anyone with that role to read only
}

func userListHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userListHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
		// call api to get a slice of users
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		roleSvc := api.NewRoleServiceClient(apiClient)
		userSvc := api.NewUserServiceClient(apiClient)

		roleReq := new(api.RoleListReq)
		roleRes, err := roleSvc.List(ctx, roleReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userListHandler() > rolesSvc.Slice(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userReq := new(api.UserListReq)
		userRes, err := userSvc.List(ctx, userReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userListHandler() > userSvc.Slice(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get notifications if there are any
		notification, err := getNotification(w, r)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userListHandler() > getNotification(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render(w, r, "userList.html",
			struct {
				Notification string
				Roles        []*api.Role
				Users        []*api.User
			}{
				Notification: notification,
				Roles:        roleRes.Roles,
				Users:        userRes.Users,
			},
		)
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userListHandler() > session.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func userReadHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/userHandler.go > userReadHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
		// get variables from uri
		vars := mux.Vars(r)

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC role and user service
		roleSvc := api.NewRoleServiceClient(apiClient)
		userSvc := api.NewUserServiceClient(apiClient)

		// gRPC get all roles
		roleReq := new(api.RoleListReq)
		roleRes, err := roleSvc.List(ctx, roleReq)
		if err != nil {
			log.Printf("ERROR > controllers/userHandler.go > userReadHandler() > rolsSvc.List(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// gRPC get a user
		userReq := new(api.UserReadReq)
		userReq.ID = vars["id"]
		userRes, err := userSvc.Read(ctx, userReq)
		if err != nil {
			log.Printf("ERROR > controllers/userHandler.go > userReadHandler() > userSvc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// render to page
		render(w, r, "userRead.html",
			struct {
				Roles []*api.Role
				User  *api.User
			}{
				Roles: roleRes.Roles,
				User:  userRes.User,
			},
		)
	}

	// save session
	err = store.Save(r, w, session)
	if err != nil {
		log.Printf("ERROR > controllers/userHandler.go > userReadHandler() > sessions.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
		// get variables from uri
		vars := mux.Vars(r)

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC role and user services
		roleSvc := api.NewRoleServiceClient(apiClient)
		userSvc := api.NewUserServiceClient(apiClient)

		// gRPC find a role
		roleReq := new(api.RoleListReq)
		roleRes, err := roleSvc.List(ctx, roleReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// gRPC find a user
		userReq := new(api.UserReadReq)
		userReq.ID = vars["id"]
		userRes, err := userSvc.Read(ctx, userReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// reder to page
		render(w, r, "userUpsert.html",
			struct {
				CSRF   template.HTML
				Title  string
				Roles  []*api.Role
				User   *api.User
				Action string
			}{
				CSRF:   csrf.TemplateField(r),
				Title:  "Update User",
				Roles:  roleRes.Roles,
				User:   userRes.User,
				Action: "Update",
			},
		)

	case "POST":
		// get variables from uri
		vars := mux.Vars(r)

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPCuser service
		userSvc := api.NewUserServiceClient(apiClient)

		// gRPC update a user
		userReq := new(api.UserUpdateReq)
		userReq.ID = vars["id"]
		userReq.RoleID = r.FormValue("role")
		userReq.ModifiedBy = session.Values["username"].(string)
		_, err := userSvc.Update(ctx, userReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > svc.Update(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// update the session roleid
		session.Values["roleid"] = userReq.RoleID

		// put a notification in the session.Values that a user was updated
		addNotification(w, r, fmt.Sprintf("User '%s' has been updated!", r.FormValue("username")))

		// redirect to user list
		http.Redirect(w, r, "/user/list", http.StatusSeeOther)
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > session.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userDeleteHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
		// get variables from uri
		vars := mux.Vars(r)

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC user service
		svc := api.NewUserServiceClient(apiClient)

		// gRPC get a user
		readReq := new(api.UserReadReq)
		readReq.ID = vars["id"]
		readRes, err := svc.Read(ctx, readReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userDeleteHandler() > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// gRPC delete a user
		deleteReq := new(api.UserDeleteReq)
		deleteReq.ID = vars["id"]
		_, err = svc.Delete(ctx, deleteReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a user was deleted
		addNotification(w, r, fmt.Sprintf("User '%s' was deleted!", readRes.User.Username))

		// redirect to users list
		http.Redirect(w, r, "/user/list", http.StatusTemporaryRedirect)
	}

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userDeleteHandler() > session.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
