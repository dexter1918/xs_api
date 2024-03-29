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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserData is ...
type UserData struct {
	Name        string
	Username    string
	Email       string
	UserID      string
	AccessToken string
	AvatarURL   string
}

// ProviderIndex is ...
type ProviderIndex struct {
	Provider    string
	ProviderMap map[string]string
}

func main() {
	goth.UseProviders(
		twitter.New("78KQ0Abr9LVvMRoJhF3952pqS", "VjHpiYvFWS6ntGDGmIHR8aIRRLE4kxHDKhj0oDF9bNBU4rF983", "http://127.0.0.1:3000/auth/twitter/callback"),
	)
	gothic.Store = sessions.NewCookieStore([]byte("VjHpiYvFWS6ntGDGmIHR8aIRRLE4kxHDKhj0oDF9bNBU4rF983"))

	// key := "VjHpiYvFWS6ntGDGmIHR8aIRRLE4kxHDKhj0oDF9bNBU4rF983" // Replace with your SESSION_SECRET or similar
	// maxAge := 86400 * 30                                        // 30 days
	// isProd := false                                             // Set to true when serving over https

	// store := sessions.NewCookieStore([]byte(key))
	// store.MaxAge(maxAge)
	// store.Options.Path = "/"
	// store.Options.HttpOnly = true // HttpOnly should always be enabled
	// store.Options.Secure = isProd

	// gothic.Store = store

	p := pat.New()

	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		var userPresent bool = false

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}

		client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

		if err != nil {
			log.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		err = client.Connect(ctx)

		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)

		//fmt.Println("Connected to MongoDB!")

		tempDB := client.Database("testdb")
		testCollection := tempDB.Collection("myuserdata")

		cursor, err := testCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var userDoc bson.M
			if err = cursor.Decode(&userDoc); err != nil {
				log.Fatal(err)
			}
			// fmt.Println(user.NickName)
			if userDoc["userid"] == user.UserID {
				//fmt.Println("Found the user!")
				userPresent = true
				break
			}
		}
		if userPresent != true {
			myuserdataCollection := client.Database("testdb").Collection("myuserdata")

			// date := time.Now().UTC()
			// log.Println(date.Format("02 Jan 2006"))

			//Insert One user's document
			user1 := UserData{user.Name, user.NickName, user.Email, user.UserID, user.AccessToken, user.AvatarURL}
			// log.Println(user.AvatarURL)
			insertResult, err := myuserdataCollection.InsertOne(context.TODO(), user1)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("User inserted with Object ID: ", insertResult.InsertedID)
		}

		t, _ := template.ParseFiles("profile.html")
		t.Execute(res, user)
	})

	p.Get("/editinfo/{provider}", func(res http.ResponseWriter, req *http.Request) {
		// log.Println("editinfo executed!")
		a := 10
		t, _ := template.ParseFiles("editinfo.html")
		t.Execute(res, a)
	})

	p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		// This tries to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
			//log.Println("If executed")
			t, _ := template.ParseFiles("profile.html")
			t.Execute(res, gothUser)
		} else {
			//log.Println("Else executed")
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
		t, _ := template.ParseFiles("index.html")
		t.Execute(res, providerIndex)
	})

	//log.Println("listening on localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", p))
}


