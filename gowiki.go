package main

import (
	"fmt"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"github.com/lagframe/justFiles"
)

const templateDir = "tmpl"
const dataDir = "data"

var templatePaths = []string{
	filepath.Join(templateDir, "edit.tmpl"),
	filepath.Join(templateDir, "view.tmpl"),
}

var templates = template.Must(template.ParseFiles(templatePaths...))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var linkRegexp = regexp.MustCompile("\\[([a-zA-Z0-9]+)\\]")

type Page struct {
	Title string
	Body []byte
	DisplayBody template.HTML
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	os.Mkdir(dataDir, 0777)
	return ioutil.WriteFile(filepath.Join(dataDir, filename), p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filepath.Join(dataDir, filename))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // Title is the second subexpression
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl + ".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
    return
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
	}

	escapedBody := []byte(template.HTMLEscapeString(string(p.Body)))
	p.DisplayBody = template.HTML(linkRegexp.ReplaceAllFunc(escapedBody, func(str []byte) []byte {
	    matched := linkRegexp.FindStringSubmatch(string(str))
	    out := []byte("<a href=\"/view/"+matched[1]+"\">"+matched[1]+"</a>")
	    return out
	  }))
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title : title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2]) // Title is the second subexpression
	}
}

func main() {
	fmt.Println("PORT: "+os.Getenv("PORT"));
	fileSystem := justFiles.JustFilesFilesystem{http.Dir("resources/")}
 	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(fileSystem)))

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
