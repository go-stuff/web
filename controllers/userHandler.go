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

func userListHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userListHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userListHandler() > session.Save(): %s\n", err.Error())
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

func userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get variables from uri
	vars := mux.Vars(r)

	// prepare api
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	roleSvc := api.NewRoleServiceClient(apiClient)
	userSvc := api.NewUserServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a role
		roleReq := new(api.RoleListReq)
		roleRes, err := roleSvc.List(ctx, roleReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// use the api to find a role
		userReq := new(api.UserReadReq)
		userReq.ID = vars["id"]
		userRes, err := userSvc.Read(ctx, userReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// save session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("ERROR > controllers/rolesHandler.go > userUpdateHandler() > session.Save(): %s\n", err.Error())
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
		// use api to update user
		userReq := new(api.UserUpdateReq)
		userReq.User = new(api.User)
		userReq.User.ID = vars["id"]
		userReq.User.Username = r.FormValue("username")
		userReq.User.RoleID = r.FormValue("role")
		userReq.User.ModifiedBy = session.Values["username"].(string)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := userSvc.Update(ctx, userReq)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > svc.Update(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a user was updated
		addNotification(session, fmt.Sprintf("User '%s' has been updated!", r.FormValue("username")))

		// save session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("ERROR > controllers/usersHandler.go > userUpdateHandler() > session.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to user list
		http.Redirect(w, r, "/user/list", http.StatusSeeOther)
	}
}

func userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("controllers/usersHandler.go > ERROR > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get variables from uri
	vars := mux.Vars(r)

	// prepare api
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	svc := api.NewUserServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a user
		readReq := new(api.UserReadReq)
		readReq.ID = vars["id"]
		readRes, err := svc.Read(ctx, readReq)
		if err != nil {
			log.Printf("controllers/usersHandler.go > ERROR > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// use the api to delete a user
		deleteReq := new(api.UserDeleteReq)
		deleteReq.ID = vars["id"]
		_, err = svc.Delete(ctx, deleteReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a user was deleted
		addNotification(session, fmt.Sprintf("User '%s' was deleted!", readRes.User.Username))

		// save session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("controllers/usersHandler.go > ERROR > session.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to users list
		http.Redirect(w, r, "/user/list", http.StatusTemporaryRedirect)
	}
}
