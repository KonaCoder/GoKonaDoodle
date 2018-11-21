package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

type person struct {
	FirstName  string
	LastName   string
	Email      string
	Subscribed bool
}

func main() {
	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.HandleFunc("/about", about)
	http.HandleFunc("/contact", contact)
	http.HandleFunc("/subscribe", subscribe)
	http.HandleFunc("/upload", upload)

	http.HandleFunc("/", index)

	http.Handle("/content/", http.StripPrefix("/content", http.FileServer(http.Dir("content"))))
	http.Handle("/img/", http.StripPrefix("/img", http.FileServer(http.Dir("img"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts", http.FileServer(http.Dir("scripts"))))
	http.Handle("/styles/", http.StripPrefix("/styles", http.FileServer(http.Dir("styles"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(res http.ResponseWriter, req *http.Request) {
	fmt.Println("index requested @ " + req.URL.Path)

	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return
	}

	err := tpl.ExecuteTemplate(res, "index.gohtml", nil)
	if err != nil {
		log.Fatalln("index template didn't execute: ", err)
	}
}

func about(res http.ResponseWriter, req *http.Request) {
	fmt.Println("about requested")
	err := tpl.ExecuteTemplate(res, "about.gohtml", nil)
	if err != nil {
		log.Fatalln("about template didn't execute: ", err)
	}
}

func contact(res http.ResponseWriter, req *http.Request) {
	fmt.Println("contact requested")
	err := tpl.ExecuteTemplate(res, "contact.gohtml", nil)
	if err != nil {
		log.Fatalln("contact template didn't execute: ", err)
	}
}

// Upload a file and save to public/uploaded folder.
func upload(w http.ResponseWriter, req *http.Request) {
	fmt.Println("upload requested")

	var s string
	if req.Method == http.MethodPost {

		// open
		f, h, err := req.FormFile("q")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		// for your information
		fmt.Println("\nfile:", f, "\nheader:", h, "\nerr", err)

		// read
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s = string(bs)

		// store on server
		dst, err := os.Create(filepath.Join("./public/uploaded/", h.Filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = dst.Write(bs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl.ExecuteTemplate(w, "upload.gohtml", s)
}

func subscribe(w http.ResponseWriter, req *http.Request) {
	fmt.Println("subscribe requested")

	// body
	bs := make([]byte, req.ContentLength)
	req.Body.Read(bs)
	body := string(bs)

	err := tpl.ExecuteTemplate(w, "subscribe.gohtml", body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}
