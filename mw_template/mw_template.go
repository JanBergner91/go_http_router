package mw_template

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var templates = template.Must(template.ParseFiles("./html-templates/index.html"))

func WriteTemplate(title, description, templatedir string, status int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			initTemplates(templatedir)
			w.WriteHeader(status)
			renderTemplate(w, "index.html", nil)
		})
	}
}

func ProcessTemplate(w http.ResponseWriter, templatename, templatedir string, status int, data interface{}) {
	initTemplates(templatedir)
	w.WriteHeader(status)
	renderTemplate(w, templatename, data)
}

func initTemplates(directory string) {

	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	var filePaths []string
	for _, file := range files {
		if !file.IsDir() {
			filePaths = append(filePaths, filepath.Join(directory, file.Name()))
			fmt.Println("Added " + file.Name() + " to list")
		}
	}

	tmpl, err := template.ParseFiles(filePaths...)
	if err != nil {
		log.Fatal(err)
	}

	templates = template.Must(tmpl, err)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
