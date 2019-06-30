package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-stuff/grpc/api"
	"github.com/golang/protobuf/ptypes"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Audit any changes to the system
func Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only consider put, post and patch
		switch r.Method {
		case "PUT", "POST", "PATCH":

			// get session
			session, err := store.Get(r, "session")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if session.Values["username"] != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				auditSvc := api.NewAuditServiceClient(apiClient)

				auditReq := new(api.AuditCreateReq)
				auditReq.Audit = &api.Audit{
					ID:        primitive.NewObjectID().Hex(),
					Username:  fmt.Sprintf("%v", session.Values["username"]),
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
			}
		}

		// Send the results of this http request to the next handler.
		next.ServeHTTP(w, r)
		return
	})
}
