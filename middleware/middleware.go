package middleware

import (
	"github.com/go-stuff/mongostore"
)

var (
	store *mongostore.MongoStore
)

// Init gets the store pointer from main.go
func Init(mongostore *mongostore.MongoStore) {
	store = mongostore
}
