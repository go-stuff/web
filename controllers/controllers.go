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
	store     *mongostore.MongoStore
	layout    *template.Template
	templates map[string]*template.Template
)

// Init gets the store pointer from main.go and returns a router
// pointer to main.go
func Init(mongostore *mongostore.MongoStore) *mux.Router {
	store = mongostore

	err := initTemplates()
	if err != nil {
		log.Fatal(err)
	}

	router := initRouter()

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

		// merge the base template and fileContents
		_, err = content.Parse(string(fileContents))
		if err != nil {
			return err
		}

		// add the merged content to the templates map
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
