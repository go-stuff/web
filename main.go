package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/securecookie"

	"github.com/go-stuff/mongostore"
	"github.com/go-stuff/web/controllers"
	"github.com/go-stuff/web/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// init database
	client, ctx, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// set a default session ttl to 20 minutes
	if os.Getenv("MONGOSTORE_SESSION_TTL") == "" {
		os.Setenv("MONGOSTORE_SESSION_TTL", strconv.Itoa(20*60))
	}

	// get the ttl from an environment variable
	ttl, err := strconv.Atoi(os.Getenv("MONGOSTORE_SESSION_TTL"))
	if err != nil {
		log.Fatal(err)
	}

	// init store
	store, err := initMongoStore(client.Database("test").Collection("sessions"), ttl)
	if err != nil {
		log.Fatal(err)
	}

	// init controllers
	router := controllers.Init(client, store)

	// init middlware
	middleware.Init(store)

	// apply middleware
	router.Use(middleware.Auth)

	// init server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// start server
	log.Println("main.go > INFO > Listening and Serving @", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func initMongoClient() (*mongo.Client, context.Context, error) {
	// a Context carries a deadline, cancelation signal, and request-scoped values
	// across API boundaries. Its methods are safe for simultaneous use by multiple
	// goroutines
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// use a default mongo uri if the MONGOURL environment variable is not set
	if os.Getenv("MONGOURL") == "" {
		os.Setenv("MONGOURL", "mongodb://localhost:27017")
	}

	// connect does not do server discovery, use ping
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGOURL")))
	if err != nil {
		return nil, nil, err
	}

	// ping for server discovery
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, nil, err
	}

	log.Println("main.go > INFO > Connected to MongoDB @", os.Getenv("MONGOURL"))
	return client, ctx, nil
}

func initMongoStore(col *mongo.Collection, age int) (*mongostore.MongoStore, error) {
	// generate an authentication key to use if the GORILLA_SESSION_AUTH_KEY environment
	// variable is not set
	if os.Getenv("GORILLA_SESSION_AUTH_KEY") == "" {
		os.Setenv("GORILLA_SESSION_AUTH_KEY", base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)))
	}

	// generate an encryption key to use if the GORILLA_SESSION_ENC_KEY environment
	// variable is not set
	if os.Getenv("GORILLA_SESSION_ENC_KEY") == "" {
		os.Setenv("GORILLA_SESSION_ENC_KEY", base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16)))
	}

	store := mongostore.NewMongoStore(
		col,
		age,
		[]byte(os.Getenv("GORILLA_SESSION_AUTH_KEY")),
		[]byte(os.Getenv("GORILLA_SESSION_ENC_KEY")),
	)

	return store, nil
}
