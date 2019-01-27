package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	
	"google.golang.org/appengine"
	"google.golang.org/appengine/mail"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

type subscriber struct {
	Name    string
	Phone   string
	Email   string
	Message string
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

	// This should NOT be needed since this is using the new go111 runtime but... to create a valid
	// context object to use for sending mail via appengine... this is necessary. {facepalm}
	// HOWEVER, if you run locally w/o commenting out below it was throw up at runtime.
	appengine.Main()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, req *http.Request) {
	fmt.Println("index requested @ " + req.URL.Path)

	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	err := tpl.ExecuteTemplate(w, "index.gohtml", nil)
	if err != nil {
		log.Fatalln("index template didn't execute: ", err)
	}
}

func about(w http.ResponseWriter, req *http.Request) {
	fmt.Println("about requested")
	err := tpl.ExecuteTemplate(w, "about.gohtml", nil)
	if err != nil {
		log.Fatalln("about template didn't execute: ", err)
	}
}

func contact(w http.ResponseWriter, req *http.Request) {
	fmt.Println("contact requested")

	n := req.FormValue("name")
	p := req.FormValue("phone")
	e := req.FormValue("email")
	m := req.FormValue("message")

	s := subscriber{
		Name:    n,
		Phone:   p,
		Email:   e,
		Message: m,
	}

	// Added check for > 50 in case the maxlength of the input is overridden by malicious user.
	if e != "" || len(e) > 50 {
		log.Println("Contact Email: ", e)
		// Send email.
		// GCP Doc: https://cloud.google.com/appengine/docs/standard/go/mail/sending-receiving-with-mail-api
		ctx := appengine.NewContext(req)
		req = req.WithContext(ctx)

		msg := &mail.Message{
			Sender: "carriekroutil@gmail.com", // FYI: Registered as an authorized sender in GCP so no password needed in code.
			To: []string{"Carrie <carrie@shebytes.io>"},
			Subject: "New Subscriber - Kona Doodle: " + e,
			Body: "New Subscriber: \n\n Name: " + n + "\n Phone: " + p + "n Email: " + e + "\n\n Message: " + m,
		}
		if err := mail.Send(req.Context(), msg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err := tpl.ExecuteTemplate(w, "contact.gohtml", s)
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

	e := req.FormValue("email")

	if e != "" || len(e) > 50 {
		log.Println("Email: ", e)
	}

	err := tpl.ExecuteTemplate(w, "subscribe.gohtml", e)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}