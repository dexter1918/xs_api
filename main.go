package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/twitter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserData is ...
type UserData struct {
	UserID string
	Email  string
	Name   string
}

// ProviderIndex is ...
type ProviderIndex struct {
	Provider    string
	ProviderMap map[string]string
}

func main() {

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected to MongoDB!")

	goth.UseProviders(
		twitter.New("78KQ0Abr9LVvMRoJhF3952pqS", "VjHpiYvFWS6ntGDGmIHR8aIRRLE4kxHDKhj0oDF9bNBU4rF983", "http://127.0.0.1:3000/auth/twitter/callback"),
	)
	gothic.Store = sessions.NewCookieStore([]byte("VjHpiYvFWS6ntGDGmIHR8aIRRLE4kxHDKhj0oDF9bNBU4rF983"))

	p := pat.New()

	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		t, _ := template.ParseFiles("templates/success.html")
		t.Execute(res, user)

		// defer cursor.Close(ctx)
		// for cursor.Next(ctx) {
		// 	var userDoc bson.M
		// 	if err = cursor.Decode(&userDoc); err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	if userDoc["userid"] == "ssssssss" {
		// 		fmt.Println("It's here")
		// 	}
		// }

		myuserdataCollection := client.Database("testdb").Collection("myuserdata")

		//Insert One document
		user1 := UserData{user.UserID, user.Email, user.Name}
		insertResult, err := myuserdataCollection.InsertOne(context.TODO(), user1)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	})

	p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		// This tries to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
			log.Println("If executed")
			t, _ := template.ParseFiles("templates/success.html")
			t.Execute(res, gothUser)
		} else {
			log.Println("Else executed")
			gothic.BeginAuthHandler(res, req)
		}
	})

	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		m := make(map[string]string)
		m["twitter"] = "Twitter"

		providerIndex := &ProviderIndex{
			Provider:    "twitter",
			ProviderMap: m,
		}
		t, _ := template.ParseFiles("templates/index.html")
		t.Execute(res, providerIndex)
	})

	log.Println("listening on localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", p))
}
