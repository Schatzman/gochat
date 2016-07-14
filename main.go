package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"trace"
	"text/template"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func main() {
	port := ":8080"
	var addr = flag.String("addr", port, "The addr of the application.")
	flag.Parse()

	var ipAddr = "http://localhost" + port
	authCb := "/auth/callback/"
	key := "258566265514-iapn65754umceaob20auplhrms99lecm.apps.googleusercontent.com"
	secret := "7p8n_N5Isaer3YrClBJMKelm"
	gomniauth.SetSecurityKey("datta bayo")
	gomniauth.WithProviders(
		facebook.New(key, secret,
			ipAddr + authCb + "facebook"),
		github.New(key, secret,
			ipAddr + authCb + "github"),
		google.New(key, secret,
			ipAddr + authCb + "google"))

	r := newRoom()
	r.tracer = trace.New(os.Stdout)
	
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)

	http.Handle("/room", r)
	
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))


	go r.run()
	log.Println("Starting gochat web server on", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
