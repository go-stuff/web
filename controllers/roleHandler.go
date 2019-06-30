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

	"github.com/go-stuff/grpc/api"
)

// roleSeed adds the admin and read only built-in roles
func roleSeed() error {
	// create a context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// gRPC role service
	roleSvc := api.NewRoleServiceClient(apiClient)

	// gRPC get a role named admin
	readReq := new(api.RoleReadByNameReq)
	readReq.Name = "Admin"
	readRes, err := roleSvc.ReadByName(ctx, readReq)
	if err != nil {
		return err
	}

	// if the admin role does not exist create it
	if readRes.Role.ID == "" {
		// gRPC create a role
		createReq := new(api.RoleCreateReq)
		createReq.Name = "Admin"
		createReq.Description = "Administrative Role (Built-In)"
		createReq.CreatedBy = "System"
		_, err := roleSvc.Create(ctx, createReq)
		if err != nil {
			return err
		}
	}

	// gRPC get a role named read only
	readReq = new(api.RoleReadByNameReq)
	readReq.Name = "Read Only"
	readRes, err = roleSvc.ReadByName(ctx, readReq)
	if err != nil {
		return err
	}

	// if the read only role does not exist create it
	if readRes.Role.ID == "" {
		// gRPC create a role
		createReq := new(api.RoleCreateReq)
		createReq.Name = "Read Only"
		createReq.Description = "Read Only Role (Built-In)"
		createReq.CreatedBy = "System"
		_, err = roleSvc.Create(ctx, createReq)
		if err != nil {

			return err
		}
	}

	return nil
}

func roleListHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleListHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC role service
		roleSvc := api.NewRoleServiceClient(apiClient)

		// gRPC get all roles
		roleReq := new(api.RoleListReq)
		roleRes, err := roleSvc.List(ctx, roleReq)
		if err != nil {
			log.Printf("ERROR > controllers/rolesHandler.go > roleSvc.List(): %s\n", err.Error())
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
				CSRF         template.HTML
				Notification string
				Roles        []*api.Role
			}{
				CSRF:         csrf.TemplateField(r),
				Notification: notification,
				Roles:        roleRes.Roles,
			},
		)
	}

	// save session
	err = store.Save(r, w, session)
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleListHandler() > sessions.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func roleCreateHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleCreateHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "GET":
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
		// parse form fields
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC role service
		roleSvc := api.NewRoleServiceClient(apiClient)

		// gRPC create a role
		roleReq := new(api.RoleCreateReq)
		roleReq.Name = r.FormValue("name")
		roleReq.Description = r.FormValue("description")
		roleReq.CreatedBy = session.Values["username"].(string)
		_, err = roleSvc.Create(ctx, roleReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was added
		addNotification(w, r, fmt.Sprintf("Role '%s' has been created!", roleReq.Name))

		// reseed routes
		routeSeed()

		// redirect to roles list
		http.Redirect(w, r, "/role/list", http.StatusSeeOther)
	}

	// save session
	err = store.Save(r, w, session)
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleCreateHandler() > sessions.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func roleReadHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleReadHandler() > store.Get(): %s\n", err.Error())
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

		// gRPC role service
		roleSvc := api.NewRoleServiceClient(apiClient)

		// gRPC get a role
		roleReq := new(api.RoleReadReq)
		roleReq.ID = vars["id"]
		roleRes, err := roleSvc.Read(ctx, roleReq)
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

	// save session
	err = store.Save(r, w, session)
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleReadHandler() > sessions.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func roleUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleUpdateHandler() > store.Get(): %s\n", err.Error())
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

		// gRPC role service
		roleSvc := api.NewRoleServiceClient(apiClient)

		// gRPC get a role
		roleReq := new(api.RoleReadReq)
		roleReq.ID = vars["id"]
		roleRes, err := roleSvc.Read(ctx, roleReq)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > svc.Read(): %s\n", err.Error())
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
		// get variables from uri
		vars := mux.Vars(r)

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC role service
		roleSvc := api.NewRoleServiceClient(apiClient)

		// gRPC update a role
		roleReq := new(api.RoleUpdateReq)
		roleReq.ID = vars["id"]
		roleReq.Name = r.FormValue("name")
		roleReq.Description = r.FormValue("description")
		roleReq.ModifiedBy = session.Values["username"].(string)
		_, err := roleSvc.Update(ctx, roleReq)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > svc.Update(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was updated
		addNotification(w, r, fmt.Sprintf("Role '%s' has been updated!", r.FormValue("name")))

		// reseed routes
		routeSeed()

		// redirect to roles list
		http.Redirect(w, r, "/role/list", http.StatusSeeOther)
	}

	// save session
	err = store.Save(r, w, session)
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleUpdateHandler() > sessions.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func roleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleDeleteHandler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// handle each method
	switch r.Method {
	case "POST":
		// get variables from uri
		vars := mux.Vars(r)

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// gRPC role service
		roleSvc := api.NewRoleServiceClient(apiClient)

		// gRPC get a role
		readReq := new(api.RoleReadReq)
		readReq.ID = vars["id"]
		readRes, err := roleSvc.Read(ctx, readReq)
		if err != nil {
			log.Printf("controllers/rolesHandler.go > ERROR > svc.Read(): %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// gRPC delete a role
		deleteReq := new(api.RoleDeleteReq)
		deleteReq.ID = vars["id"]
		_, err = roleSvc.Delete(ctx, deleteReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was deleted
		addNotification(w, r, fmt.Sprintf("Role '%s' was deleted!", readRes.Role.Name))

		// reseed the routes
		routeSeed()

		// redirect to roles list
		http.Redirect(w, r, "/role/list", http.StatusSeeOther)
	}

	// save session
	err = store.Save(r, w, session)
	if err != nil {
		log.Printf("ERROR > controllers/roleHandler.go > roleDeleteHandler() > sessions.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
