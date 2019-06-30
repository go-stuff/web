package controllers

import (
	"context"
	"log"
	"net/http"

	"time"

	"github.com/go-stuff/grpc/api"
)

func auditList100Handler(w http.ResponseWriter, r *http.Request) {
	// get audit
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ERROR > controllers/auditHandler.go > auditList100Handler() > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// display audit
	log.Printf("INFO > controllers/auditHandler.go > auditList100Handler() > audit: %v %v\n", session.Values["_id"], session.Values["username"])

	// call api to get a slice of sessions
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	auditSvc := api.NewAuditServiceClient(apiClient)

	auditReq := new(api.AuditList100Req)
	auditRes, err := auditSvc.List100(ctx, auditReq)
	if err != nil {
		log.Printf("ERROR > controllers/auditHandler.go > auditList100Handler() > auditSvc.List(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render(w, r, "auditList.html",
		struct {
			Audit []*api.Audit
		}{
			Audit: auditRes.Audits,
		},
	)
}
