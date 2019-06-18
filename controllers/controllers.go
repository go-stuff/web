package controllers

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-stuff/mongostore"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

var (
	client      *mongo.Client
	apiClient   *grpc.ClientConn
	store       *mongostore.MongoStore
	router      *mux.Router
	routes      []string
	layout      *template.Template
	templates   map[string]*template.Template
	permissions map[string]string
)

// Init gets the store pointer from main.go and returns a router
// pointer to main.go
func Init(mongoclient *mongo.Client, mongostore *mongostore.MongoStore, apiclient *grpc.ClientConn) *mux.Router {
	client = mongoclient
	store = mongostore
	apiClient = apiclient

	err := initTemplates()
	if err != nil {
		log.Fatal(err)
	}

	router = initRouter()

	CompileRoutes()

	return router
}

func initTemplates() error {
	log.Println("controllers/controllers.go > INFO > initTemplates()")

	// initialize the content files templates map
	templates = make(map[string]*template.Template)

	// build templates with content
	err := initTemplatesWithContent()
	if err != nil {
		return err
	}

	// build templates with nav and content
	err = initTemplatesWithNavAndContent()
	if err != nil {
		return err
	}

	return nil
}

// <html> head, header, content, footer </html
func initTemplatesWithContent() error {
	log.Println("controllers/controllers.go > INFO > initTemplatesWithContent()")
	var err error

	// check the validity of login.html by parsing
	layout, err = template.ParseFiles(
		"./templates/layout/mainContent.html",
		"./templates/layout/head.html",
		"./templates/layout/header.html",
		"./templates/layout/bypass.html",
		"./templates/layout/footer.html",
	)
	if err != nil {
		return err
	}

	layout.Funcs(timestampFM())

	// recurse content files templates and build separate templates for each of them
	filepath.Walk("./templates/mainContent", walkTemplatesPath)

	return nil
}

// <html> head, header, menu, content, footer </html
func initTemplatesWithNavAndContent() error {
	log.Println("controllers/controllers.go > INFO > initTemplatesWithNavAndContent()")
	var err error

	// check the validity of the files that make up layout.html by parsing
	layout, err = template.ParseFiles(
		"./templates/layout/mainNavContent.html",
		"./templates/layout/head.html",
		"./templates/layout/header.html",
		"./templates/layout/bypass.html",
		"./templates/layout/logout.html",
		"./templates/layout/nav.html",
		"./templates/layout/footer.html",
	)
	if err != nil {
		return err
	}

	layout.Funcs(timestampFM())

	// recurse content files templates and build separate templates for each of them
	filepath.Walk("./templates/mainMenuContent", walkTemplatesPath)

	return nil
}

// recurse a directory and build templates
func walkTemplatesPath(path string, fileInfo os.FileInfo, err error) error {

	// if the current fileInfo is not a directory
	if fileInfo.IsDir() == false {

		// check that path exists
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		// read the contents of the file
		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		file.Close()

		// clone the base template
		content := template.Must(layout.Clone())
		content.Funcs(timestampFM())

		// merge the base template and fileContents
		_, err = content.Parse(string(fileContents))
		if err != nil {
			return err
		}

		// add the merged content to the templates map
		templates[fileInfo.Name()] = content

		log.Printf("controllers/controllers.go > INFO > walkTemplatesPath(): - %s", fileInfo.Name())
	}

	return nil
}

// render templates with data
func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	log.Printf("controllers/controllers.go > INFO > render(): %s", tmpl)

	// var tpl bytes.Buffer
	// e := templates[tmpl].Execute(&tpl, data)
	// if e != nil {
	// 	log.Println(tmpl)

	// }
	// log.Println(e)
	// log.Printf("\ntmpl: %v\n", templates[tmpl])
	// log.Printf("\nbytes: %v\n", tpl.String())

	// Set the content type.
	w.Header().Set("Content-Type", "text/html")

	//templates[tmpl].Funcs(timestampFM())

	// Execute the template.
	err := templates[tmpl].Execute(w, data)
	if err != nil {
		log.Printf("controllers.go > ERROR > render(): %v", err)
		//fmt.Println(err)
	}
}

func initRouter() *mux.Router {
	log.Println("controllers/controllers.go > INFO > initRouter()")

	router := mux.NewRouter()

	// System Routes
	router.HandleFunc("/session/list", sessionListHandler).Methods("GET")

	router.HandleFunc("/role/list", roleListHandler).Methods("GET")
	router.HandleFunc("/role/create", roleCreateHandler).Methods("GET", "POST")
	router.HandleFunc("/role/read/{id}", roleReadHandler).Methods("GET")
	router.HandleFunc("/role/update/{id}", roleUpdateHandler).Methods("GET", "POST")
	router.HandleFunc("/role/delete/{id}", roleDeleteHandler).Methods("GET")

	router.HandleFunc("/route/list", routeListHandler).Methods("GET")
	router.HandleFunc("/route/update", routeUpdateHandler).Methods("GET")

	router.HandleFunc("/user/list", userListHandler).Methods("GET")
	router.HandleFunc("/user/update/{id}", userUpdateHandler).Methods("GET", "POST")
	router.HandleFunc("/user/delete/{id}", userDeleteHandler).Methods("GET")

	router.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	router.HandleFunc("/logout", loginHandler).Methods("GET")

	// App Routes
	router.HandleFunc("/", rootHandler).Methods("GET", "POST")
	router.HandleFunc("/home", homeHandler).Methods("GET")

	router.HandleFunc("/server/list", serverListHandler).Methods("GET")
	router.HandleFunc("/server/create", serverCreateHandler).Methods("GET", "POST")


	// Setup or static files.
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return router
}

// func initPermissions() error {
// 	// call api to get a slice of permissions
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()
// 	svc :=  api.NewPermissionServiceClient(apiClient)
// 	req := new(api.PermissionSliceReq)
// 	slice, err := svc.Slice(ctx, req)
// 	if err != nil {
// 		log.Printf("controllers/controllers.go > ERROR > svc.Slice(): %s\n", err.Error())
// 		return err
// 	}

// 	permissions = make(map[string]string)

// for _, permission := range slice.Permissions {
// 	permissions[permission.RoleID] = permission.Route
// }

// return nil
// }

// format timestamps
func timestampFM() template.FuncMap {
	return template.FuncMap{
		"timestamp": func(ts timestamp.Timestamp) string {
			return time.Unix(ts.Seconds, int64(ts.Nanos)).Format("2006-Jan-02 03:04:05 PM MST")
		},
	}
}

// addNotification adds a notification message to session.Values
func addNotification(session *sessions.Session, notification string) {
	session.Values["notification"] = notification
}

// getNotification returns a notification from session.Values if
// one exists, otherwise it returns an empty string
// if a notification was returned, the notification session.Value
// is emptied
func getNotification(w http.ResponseWriter, r *http.Request) (string, error) {

	// get session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("controllers/controllers.go > ERROR > store.Get(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	var notification string

	if session.Values["notification"] == nil {
		notification = ""
	} else {
		notification = session.Values["notification"].(string)
	}

	session.Values["notification"] = ""

	// save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("controllers/controllers.go > ERROR > session.Save(): %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	return notification, nil
}
