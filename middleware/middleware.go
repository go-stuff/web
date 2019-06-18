package middleware

import (
	"github.com/go-stuff/mongostore"
	"google.golang.org/grpc"
)

var (
	store     *mongostore.MongoStore
	apiClient *grpc.ClientConn
)

// Init gets the store pointer from main.go
func Init(mongostore *mongostore.MongoStore, apiclient *grpc.ClientConn) {
	store = mongostore
	apiClient = apiclient
}
