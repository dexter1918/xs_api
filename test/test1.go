package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserData is ...
type UserData struct {
	UserID string
	Email  string
	Name   string
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

	tempDB := client.Database("testdb")
	testCollection := tempDB.Collection("myuserdata")

	cursor, err := testCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Not found")
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var userDoc bson.M
		if err = cursor.Decode(&userDoc); err != nil {
			log.Fatal(err)
		}
		if userDoc["name"] == "sksalmahaider" {
			fmt.Println("It's here")
		}
	}

	myuserdataCollection := client.Database("testdb").Collection("myuserdata")

	//Insert One document
	user1 := UserData{"ssssssss", "hhhhhh@gmail.com", "rrrrrr"}
	insertResult, err := myuserdataCollection.InsertOne(context.TODO(), user1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

}
