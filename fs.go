package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
)

var templates = template.Must(template.ParseFiles("tmpl/dir.html", "tmpl/file.html"))

// Directory is a directory.
type Directory struct {
	Name  string
	Files []os.FileInfo
}

// File is a file.
type File struct {
	Name string
	Body []byte
}

func internalServerError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/v/"):]
	fi, err := os.Stat("dir/" + filename)
	if err != nil {
		internalServerError(w, err)
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, _ := ioutil.ReadDir("dir/" + filename)
		dir := &Directory{Name: filename, Files: files}
		err := templates.ExecuteTemplate(w, "dir.html", dir)
		if err != nil {
			internalServerError(w, err)
		}
	case mode.IsRegular():
		body, _ := ioutil.ReadFile("dir/" + filename)
		file := &File{Name: filename, Body: body}
		err := templates.ExecuteTemplate(w, "file.html", file)
		if err != nil {
			internalServerError(w, err)
		}
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	dirname := r.URL.Path[len("/u/"):]
	if fi, err := os.Stat("dir/" + dirname); err != nil || !fi.IsDir() {
		internalServerError(w, err)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		internalServerError(w, err)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	writeFile := fmt.Sprintf("dir/%s/%s", dirname, handler.Filename)
	err = ioutil.WriteFile(writeFile, fileBytes, 0600)
	if err != nil {
		internalServerError(w, err)
		return
	}
	http.Redirect(w, r, "/v/"+dirname, http.StatusFound)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v/", http.StatusFound)
	})
	http.HandleFunc("/v/", handleView)
	http.HandleFunc("/u/", handleUpload)
	http.ListenAndServe(":8000", nil)
}
