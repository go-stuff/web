package controllers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/go-stuff/grpc/api"
	"github.com/golang/protobuf/ptypes"
)

func rolesHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		// call api to get a slice of users
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		svc := api.NewRoleServiceClient(apiClient)
		req := new(api.RoleSliceReq)
		slice, err := svc.Slice(ctx, req)
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

		// get notifications if there are any
		notification, err := getNotification(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// render to page
		render(w, r, "roles.html",
			struct {
				Notification string
				Roles        []*api.Role
			}{
				Notification: notification,
				Roles:        slice.Roles,
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
		render(w, r, "rolesUpsert.html",
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
		// use the api to add a role
		svc := api.NewRoleServiceClient(apiClient)
		req := new(api.RoleCreateReq)
		req.Role = &api.Role{
			ID:          primitive.NewObjectID().Hex(),
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			CreatedBy:   session.Values["username"].(string),
			CreatedAt:   ptypes.TimestampNow(),
			ModifiedBy:  session.Values["username"].(string),
			ModifiedAt:  ptypes.TimestampNow(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := svc.Create(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was added
		addNotification(session, fmt.Sprintf("Role '%s' has been created!", req.Role.Name))

		// save session
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to roles list
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

	// get variables from uri
	vars := mux.Vars(r)

	// prepare api
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	svc := api.NewRoleServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a role
		req := new(api.RoleReadReq)
		req.ID = vars["id"]
		res, err := svc.Read(ctx, req)
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
		render(w, r, "rolesRead.html",
			struct {
				Role *api.Role
			}{
				Role: res.Role,
			},
		)
	}
}

func roleUpdateHandler(w http.ResponseWriter, r *http.Request) {
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
	svc := api.NewRoleServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a role
		req := new(api.RoleReadReq)
		req.ID = vars["id"]
		res, err := svc.Read(ctx, req)
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

		// reder to page
		render(w, r, "rolesUpsert.html",
			struct {
				CSRF   template.HTML
				Title  string
				Role   *api.Role
				Action string
			}{
				CSRF:   csrf.TemplateField(r),
				Title:  "Update Role",
				Role:   res.Role,
				Action: "Update",
			},
		)

	case "POST":
		// use api to update role
		req := new(api.RoleUpdateReq)
		req.Role = new(api.Role)
		req.Role.ID = vars["id"]
		req.Role.Name = r.FormValue("name")
		req.Role.Description = r.FormValue("description")
		req.Role.ModifiedBy = session.Values["username"].(string)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := svc.Update(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was updated
		addNotification(session, fmt.Sprintf("Role '%s' has been updated!", r.FormValue("name")))

		// save session
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to roles list
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

	// get variables from uri
	vars := mux.Vars(r)

	// prepare api
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	svc := api.NewRoleServiceClient(apiClient)

	switch r.Method {
	case "GET":
		// use the api to find a role
		readReq := new(api.RoleReadReq)
		readReq.ID = vars["id"]
		readRes, err := svc.Read(ctx, readReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// use the api to delete a role
		deleteReq := new(api.RoleDeleteReq)
		deleteReq.ID = vars["id"]
		_, err = svc.Delete(ctx, deleteReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// put a notification in the session.Values that a role was deleted
		addNotification(session, fmt.Sprintf("Role '%s' was deleted!", readRes.Role.Name))

		// save session
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to roles list
		http.Redirect(w, r, "/roles", http.StatusTemporaryRedirect)
	}
}
