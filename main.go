package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var embedPageTemplate *template.Template

type fileSystem struct {
	dir http.Dir
}

func (f fileSystem) Open(name string) (http.File, error) {
	file, err := f.dir.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, os.ErrPermission
	}

	return file, err
}

func main() {
	tmpl, err := template.ParseFiles("embed_page.html")
	if err != nil {
		panic(err)
	}
	embedPageTemplate = tmpl

	http.HandleFunc("/newpage", createEmbedPage)

	fs := http.FileServer(fileSystem{http.Dir("./embeds")})
	fs = http.StripPrefix("/embeds/", fs)

	http.Handle("/embeds/", fs)

	log.Fatal( http.ListenAndServe(":4043", nil) )
}

type newEmbedPage struct {
	Meta map[string]string
	Title string
	Color string
	Redirect string
	Name string
}

func requireAuth(token string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
func createEmbedPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var embedPage newEmbedPage
	data, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(data, &embedPage)

	embedPage.Name = "embeds/" + embedPage.Name

	dir := filepath.Dir(embedPage.Name)
	os.MkdirAll(dir, 0755)
	file, err := os.Create(embedPage.Name+".html")
	defer file.Close()
	if err != nil {
		panic(err)
	}



	_ = embedPageTemplate.Execute(file, embedPage)
	w.WriteHeader(http.StatusOK)
}
