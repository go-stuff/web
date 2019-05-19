package controllers

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	
	"github.com/go-stuff/mongostore"
)

var (
	store *mongostore.MongoStore
	layout    *template.Template
	templates map[string]*template.Template
)

// Init gets the store pointer from main.go and returns a router
// pointer to main.go
func Init(mongostore *mongostore.MongoStore) *mux.Router {
	store = mongostore	
	initTemplates()
	router := initRouter()
	return router
}

func initTemplates() {
	log.Println("controllers/controllers.go > INFO > initTemplates()")

	// initilize a base html template named layout.html
	layout = template.New("layout.html")

	// parse the files that makeup layout.html
	layout.ParseFiles(
		"./templates/layout/layout.html",
		"./templates/layout/header.html",
		"./templates/layout/footer.html",
		"./templates/layout/sidebar.html",
		"./templates/layout/nav.html",
	)

	// initialize the content files templates map
	templates = make(map[string]*template.Template)

	// recurse content files templates and build separate templates for each of them
	filepath.Walk("./templates/content", walkTemplatesPath)
}

// recurse content folder and build templates
func walkTemplatesPath(path string, fileInfo os.FileInfo, err error) error {

	if fileInfo.IsDir() == false {
		// check that the tmaplate file exists
		file, err := os.Open(path)
		if err != nil {
			log.Fatal("failed to open template '" + fileInfo.Name() + "'")
		}

		// read the contents of the template
		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal("failed to read content from file '" + fileInfo.Name() + "'")
		}
		file.Close()

		// clone the layout templates as base templates
		var content *template.Template
		content = template.Must(layout.Clone())

		// parse the template
		_, err = content.Parse(string(fileContents))
		if err != nil {
			log.Fatal("Failed to parse contents of '" + fileInfo.Name() + "' as template")
		}

		// add the content template to our templates map
		templates[fileInfo.Name()] = content

		log.Printf("controllers/controllers.go > INFO > - %s", fileInfo.Name())
	}

	return nil
}

// render templates with data
func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	log.Printf("controllers/controllers.go > INFO > render() > %s", tmpl)

	// var tpl bytes.Buffer
	// e := templates[tmpl].Execute(&tpl, data)
	// if e != nil {
	// 	log.Debug(tmpl)
	// 	log.Error(e)
	// }
	// log.Debug(tpl.String())

	// Set the content type.
	w.Header().Set("Content-Type", "text/html")

	// Execute the template.
	err := templates[tmpl].Execute(w, data)
	if err != nil {
		log.Printf("controllers.go > ERROR > render() > %v", err)
		//fmt.Println(err)
	}
}

func initRouter() *mux.Router {
	log.Println("controllers/controllers.go > INFO > initRouter()")

	router := mux.NewRouter()

	// Handle URLs
	router.HandleFunc("/", rootHandler).Methods("GET", "POST")
	router.HandleFunc("/home", homeHandler).Methods("GET")

	router.HandleFunc("/test", testHandler).Methods("GET")

	router.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	router.HandleFunc("/logout", loginHandler).Methods("GET")

	// Setup or static files.
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return router
}
