package controllers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/go-stuff/grpc/api"
	"github.com/golang/protobuf/ptypes"
)

func roleListHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	roleSvc := api.NewRoleServiceClient(apiClient)
	
	switch r.Method {
	case "GET":
		roleReq := new(api.RoleListReq)
		roleRes, err := roleSvc.List(ctx, roleReq)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > roleSvc.List(): %s\n", err.Error())
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

		// render to page
		render(w, r, "roleList.html",
			struct {
				Notification string
				Roles        []*api.Role
			}{
				Notification: notification,
				Roles:        roleRes.Roles,
			},
		)
	}
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
	case "GET":
		// save session
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// render to page
		render(w, r, "roleUpsert.html",
			struct {
				CSRF   template.HTML
				Title  string
				Role   *api.Role
				Action string
			}{
				CSRF:   csrf.TemplateField(r),
				Title:  "Create Role",
				Role:   new(api.Role),
				Action: "Create",
			},
		)

	case "POST":
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// use the api to add a role
		roleSvc := api.NewRoleServiceClient(apiClient)

		roleReq := new(api.RoleCreateReq)
		roleReq.Role = &api.Role{
			ID:          primitive.NewObjectID().Hex(),
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			CreatedBy:   session.Values["username"].(string),
			CreatedAt:   ptypes.TimestampNow(),
			ModifiedBy:  session.Values["username"].(string),
			ModifiedAt:  ptypes.TimestampNow(),
		}
	
		_, err := roleSvc.Create(ctx, roleReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was added
		addNotification(session, fmt.Sprintf("Role '%s' has been created!", roleReq.Role.Name))

		// save session
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to roles list
		http.Redirect(w, r, "/role/list", http.StatusSeeOther)
	}
}

func roleReadHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get variables from uri
	vars := mux.Vars(r)

	// prepare api
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	roleSvc := api.NewRoleServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a role
		roleReq := new(api.RoleReadReq)
		roleReq.ID = vars["id"]
		roleRes, err := roleSvc.Read(ctx, roleReq)
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

		// render to page
		render(w, r, "roleRead.html",
			struct {
				Role *api.Role
			}{
				Role: roleRes.Role,
			},
		)
	}
}

func roleUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("controllers/rolesHandler.go > ERROR > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get variables from uri
	vars := mux.Vars(r)

	// prepare api
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	roleSvc := api.NewRoleServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a role
		roleReq := new(api.RoleReadReq)
		roleReq.ID = vars["id"]
		roleRes, err := roleSvc.Read(ctx, roleReq)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// save session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > session.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// reder to page
		render(w, r, "roleUpsert.html",
			struct {
				CSRF   template.HTML
				Title  string
				Role   *api.Role
				Action string
			}{
				CSRF:   csrf.TemplateField(r),
				Title:  "Update Role",
				Role:   roleRes.Role,
				Action: "Update",
			},
		)

	case "POST":
		// use api to update role
		roleReq := new(api.RoleUpdateReq)
		roleReq.Role = new(api.Role)
		roleReq.Role.ID = vars["id"]
		roleReq.Role.Name = r.FormValue("name")
		roleReq.Role.Description = r.FormValue("description")
		roleReq.Role.ModifiedBy = session.Values["username"].(string)
	
		_, err := roleSvc.Update(ctx, roleReq)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > svc.Update(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was updated
		addNotification(session, fmt.Sprintf("Role '%s' has been updated!", r.FormValue("name")))

		// save session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > session.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to roles list
		http.Redirect(w, r, "/role/list", http.StatusSeeOther)
	}
}

func roleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("controllers/rolesHandler.go > ERROR > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get variables from uri
	vars := mux.Vars(r)

	// prepare api
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	roleSvc := api.NewRoleServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a role
		readReq := new(api.RoleReadReq)
		readReq.ID = vars["id"]
		readRes, err := roleSvc.Read(ctx, readReq)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// use the api to delete a role
		deleteReq := new(api.RoleDeleteReq)
		deleteReq.ID = vars["id"]
		_, err = roleSvc.Delete(ctx, deleteReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was deleted
		addNotification(session, fmt.Sprintf("Role '%s' was deleted!", readRes.Role.Name))

		// save session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > session.Save(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to roles list
		http.Redirect(w, r, "/role/list", http.StatusTemporaryRedirect)
	}
}
