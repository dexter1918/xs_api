package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/twitter"
)

// CallbackHandler is the 1st endpoint
func CallbackHandler(res http.ResponseWriter, req *http.Request) {

	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		fmt.Fprintln(res, err)
		return
	}
	t, _ := template.ParseFiles("templates/success.html")
	t.Execute(res, user)
}

// LogoutHandler is the 2nd endpoint
func LogoutHandler(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// AutoAuthenticateHanlder is the 3rd endpoint
func AutoAuthenticateHanlder(res http.ResponseWriter, req *http.Request) {
	// This tries to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
		t, _ := template.ParseFiles("templates/success.html")
		t.Execute(res, gothUser)
	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

// RootHandler is ...
func RootHandler(res http.ResponseWriter, req *http.Request) {
	m := make(map[string]string)
	m["twitter"] = "Twitter"

	providerIndex := &ProviderIndex{
		Providers:    "twitter",
		ProvidersMap: m,
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(res, providerIndex)
}

func main() {
	goth.UseProviders(
		twitter.New("78KQ0Abr9LVvMRoJhF3952pqS", "VjHpiYvFWS6ntGDGmIHR8aIRRLE4kxHDKhj0oDF9bNBU4rF983", "http://127.0.0.1:3000/auth/twitter/callback"),
	)
	gothic.Store = sessions.NewCookieStore([]byte("VjHpiYvFWS6ntGDGmIHR8aIRRLE4kxHDKhj0oDF9bNBU4rF983"))

	p := pat.New()

	p.Get("/auth/{provider}/callback", CallbackHandler)

	p.Get("/logout/{provider}", LogoutHandler)

	p.Get("/auth/{provider}", AutoAuthenticateHanlder)

	p.Get("/", RootHandler)

	log.Println("listening on localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", p))
}

// ProviderIndex is ...
type ProviderIndex struct {
	Providers    string
	ProvidersMap map[string]string
}
