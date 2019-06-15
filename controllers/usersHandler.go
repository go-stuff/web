package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-stuff/grpc/api"
)

func usersHandler(w http.ResponseWriter, r *http.Request) {
	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// call api to get a slice of users
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	userSC := api.NewUserServiceClient(apiClient)
	req := new(api.UserSliceReq)
	slice, err := userSC.Slice(ctx, req)
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

	render(w, r, "users.html",
		struct {
			Notification string
			Users        []*api.User
		}{
			Notification: notification,
			Users:        slice.Users,
		},
	)
}
