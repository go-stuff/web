package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/securecookie"

	"github.com/go-stuff/mongostore"
	"github.com/go-stuff/web/controllers"
	"github.com/go-stuff/web/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var MONGOURL string

func main() {

	// init database
	client, ctx, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// init store
	store, err := initMongoStore(client.Database("test").Collection("sessions"), 240)
	if err != nil {
		log.Fatal(err)
	}

	// init controllers
	router := controllers.Init(store)

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
	log.Printf("main.go > INFO > Listening and Serving on %s ...", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func initMongoClient() (*mongo.Client, context.Context, error) {
	// A Context carries a deadline, cancelation signal, and request-scoped values
	// across API boundaries. Its methods are safe for simultaneous use by multiple
	// goroutines.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	MONGOURL, ok := os.LookupEnv("MONGOURL")
	if !ok {
		MONGOURL = "mongodb://localhost:27017"
	}

	// Connect does not do server discovery, use Ping method.
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MONGOURL))
	if err != nil {
		return nil, nil, err
	}

	// Ping for server discovery.
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, nil, err
	}

	log.Println("main.go > INFO > Connected to MongoDB. @ %s", MONGOURL)
	return client, ctx, nil
}

func initMongoStore(col *mongo.Collection, age int) (*mongostore.MongoStore, error) {
	// set authentication key environment variables if it is not set
	if os.Getenv("GORILLA_SESSION_AUTH_KEY") == "" {
		os.Setenv("GORILLA_SESSION_AUTH_KEY", string(securecookie.GenerateRandomKey(32)))
	}

	// set encryption key environment variable if it is not set
	if os.Getenv("GORILLA_SESSION_ENC_KEY") == "" {
		os.Setenv("GORILLA_SESSION_ENC_KEY", string(securecookie.GenerateRandomKey(16)))
	}

	store := mongostore.NewMongoStore(
		col,
		age,
		[]byte(os.Getenv("GORILLA_SESSION_AUTH_KEY")), // Auth Key
		[]byte(os.Getenv("GORILLA_SESSION_ENC_KEY")),  // Enc Key
	)

	return store, nil
}
