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

	// use a default mongo url if the MONGOURL environment variable is not set
	MongoURL, ok := os.LookupEnv("MONGOURL")
	if !ok {
		MongoURL = "mongodb://localhost:27017"
	}

	// connect does not do server discovery, use ping
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURL))
	if err != nil {
		return nil, nil, err
	}

	// ping for server discovery
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, nil, err
	}

	log.Println("main.go > INFO > Connected to MongoDB @", MongoURL)
	return client, ctx, nil
}

func initMongoStore(col *mongo.Collection, age int) (*mongostore.MongoStore, error) {
	// generate an authentication key to use if the GORILLA_SESSION_AUTH_KEY environment
	// variable is not set
	GorillaSessionAuthKey, ok := os.LookupEnv("GORILLA_SESSION_AUTH_KEY")
	if !ok {
		GorillaSessionAuthKey = string(securecookie.GenerateRandomKey(32))
	}

	// generate an encryption key to use if the GORILLA_SESSION_ENC_KEY environment
	// variable is not set
	GorillaSessionEncKey, ok := os.LookupEnv("GORILLA_SESSION_ENC_KEY")
	if !ok {
		GorillaSessionEncKey = string(securecookie.GenerateRandomKey(16))
	}

	store := mongostore.NewMongoStore(
		col,
		age,
		[]byte(GorillaSessionAuthKey),
		[]byte(GorillaSessionEncKey),
	)

	return store, nil
}
