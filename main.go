package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"google.golang.org/grpc"

	"github.com/go-stuff/mongostore"
	"github.com/go-stuff/web/controllers"
	"github.com/go-stuff/web/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// init environment
	err := initEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	// init database
	client, ctx, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	apiClient, err := initAPI()
	defer apiClient.Close()

	// get database name from an environment variable
	if os.Getenv("MONGOSTORE_HTTPS_ONLY") == "" {
		os.Setenv("MONGOSTORE_HTTPS_ONLY", "false")
	}

	// set a default session ttl to 20 minutes
	if os.Getenv("MONGOSTORE_SESSION_TTL") == "" {
		os.Setenv("MONGOSTORE_SESSION_TTL", strconv.Itoa(20*60))
	}

	// get the ttl from an environment variable
	ttl, err := strconv.Atoi(os.Getenv("MONGOSTORE_SESSION_TTL"))
	if err != nil {
		log.Fatal(err)
	}

	// get database name from an environment variable
	if os.Getenv("MONGO_DB_NAME") == "" {
		os.Setenv("MONGO_DB_NAME", "test")
	}

	// users in this ad group are admins by default
	if os.Getenv("ADMIN_AD_GROUP") == "" {
		os.Setenv("ADMIN_AD_GROUP", "SomeADGroup")
	}

	// init store
	store, err := initMongoStore(client.Database(os.Getenv("MONGO_DB_NAME")).Collection("sessions"), ttl)
	if err != nil {
		log.Fatal(err)
	}

	// init controllers
	router := controllers.Init(client, store, apiClient)

	// init middlware
	middleware.Init(store, apiClient)

	// generate an csrf key to use if the GORILLA_CSRF_KEY environment
	// variable is not set
	if os.Getenv("GORILLA_CSRF_KEY") == "" {
		os.Setenv("GORILLA_CSRF_KEY", base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)))
	}

	// Generate Keys
	// fmt.Println(base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)))

	// All POST requests without a valid token will return HTTP 403 Forbidden.
	// We should also ensure that our mutating (non-idempotent) handler only
	// matches on POST requests. We can check that here, at the router level, or
	// within the handler itself via r.Method.
	middlewareCSRF := csrf.Protect(
		[]byte(os.Getenv("GORILLA_CSRF_KEY")),
		// PS: Don't forget to pass csrf.Secure(false) if you're developing locally
		// over plain HTTP (just don't leave it on in production).
		csrf.Secure(false),
	)

	// if os.Getenv("ENVIRONMENT") != "production" {
	// 	middlewareCSRF = csrf.Protect(
	// 		[]byte(os.Getenv("GORILLA_CSRF_KEY")),
	// 		// PS: Don't forget to pass csrf.Secure(false) if you're developing locally
	// 		// over plain HTTP (just don't leave it on in production).
	// 		csrf.Secure(false),
	// 	)
	// }

	// apply middleware
	router.Use(middlewareCSRF)
	router.Use(middleware.Headers)
	router.Use(middleware.Auth) // Auth should be before Permissions
	router.Use(middleware.Permissions)
	router.Use(middleware.Audit)

	// init server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// start server
	log.Println("INFO > main.go > main(): Listening and Serving @", server.Addr)
	//err = server.ListenAndServeTLS("./cert.pem", "./key.pem")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func initEnvironment() error {
	_, err := os.Stat(".env")
	if os.IsNotExist(err) {
		log.Println("INFO > main.go > initEnvironment(): .env does not exist")
	} else {
		log.Println("INFO > main.go > initEnvironment(): .env loaded")
		// open .env
		file, err := os.Open(".env")
		if err != nil {
			return err
		}
		defer file.Close()

		// read each line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if scanner.Text() != "" {
				regex := regexp.MustCompile(`([a-zA-Z0-9-_]*)\s*=\s*"(.*)"`)
				matches := regex.FindStringSubmatch(scanner.Text())
				if len(matches) != 3 {
					return errors.New("error in .env")
				}

				// []matches = [0]line, [1](group1), [2](group2)
				err := os.Setenv(matches[1], matches[2])
				if err != nil {
					return err
				}

				// DO NOT PRINT ENVIRONMENT VARIABLES OUT
				// fmt.Printf("env:%s = %s\n", matches[1], os.Getenv(matches[1]))
			}
		}

		// catch scanner errors
		err = scanner.Err()
		if err != nil {
			return err
		}
	}
	return nil
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

	// Register custom codecs for protobuf Timestamp and wrapper types
	//reg := bsoncodec.Registry()
	//reg := bsoncodec.NewRegistryBuilder().Build()

	// connect does not do server discovery, use ping
	client, err := mongo.Connect(ctx,
		options.Client().
			ApplyURI(os.Getenv("MONGOURL")), //.
		//	SetRegistry(reg),
	)
	if err != nil {
		return nil, nil, err
	}

	// ping for server discovery
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, nil, err
	}

	log.Println("INFO > main.go > initMongoClient(): Connected to MongoDB @", os.Getenv("MONGOURL"))
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

	// DO NOT PRINT OUT SESSION KEYS
	// fmt.Println(os.Getenv("GORILLA_SESSION_AUTH_KEY"))
	// fmt.Println(os.Getenv("GORILLA_SESSION_ENC_KEY"))

	// Generate Keys
	// fmt.Println(base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)))
	// fmt.Println(base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16)))

	store := mongostore.NewMongoStore(
		col,
		age,
		[]byte(os.Getenv("GORILLA_SESSION_AUTH_KEY")),
		[]byte(os.Getenv("GORILLA_SESSION_ENC_KEY")),
	)

	return store, nil
}

func initAPI() (*grpc.ClientConn, error) {
	// creds, err := credentials.NewClientTLSFromFile("./certs/cert.pem", "")
	// if err != nil {
	// 	return nil, err
	// }
	// with cert
	//conn, err := grpc.Dial("127.0.0.1:6000", grpc.WithTransportCredentials(creds))
	// without cert
	conn, err := grpc.Dial("127.0.0.1:6000", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
