package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	model "github.com/SaifTechie/bet-sys-backend/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateUserHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var user model.User
		wg.Add(1) // Increment the WaitGroup counter

		go func() {
			defer wg.Done() // Decrement the WaitGroup counter when the goroutine complete

			// Decode the request body into the user struct
			err := json.NewDecoder(r.Body).Decode(&user)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			// Get the user collection from the database
			userCollection := client.Database("bet-sys").Collection("users")

			// Insert the user into the database
			_, err = userCollection.InsertOne(r.Context(), user)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}
		}()

		// Wait for the goroutine to finish before responding to the client
		wg.Wait()

		// Respond to the client (assuming a quick response)
		model.SendResponse(w, "User Created Successfully", http.StatusOK, user)
	}
}

func GetUserHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var user model.User

		wg.Add(1) // Increment the WaitGroup counter

		go func() {
			defer wg.Done() // Decrement the WaitGroup counter when the goroutine completes

			// Get the user ID from the request URL
			vars := mux.Vars(r)
			userID := vars["id"]
			fmt.Println(userID)
			id, _ := primitive.ObjectIDFromHex(userID)
			fmt.Println(id)
			// Get the user collection from the database
			userCollection := client.Database("bet-sys").Collection("users")

			err := userCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&user)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}
		}()

		// Wait for the goroutine to finish before responding to the client
		wg.Wait()

		// Respond to the client (assuming a quick response)
		model.SendResponse(w, "User Fetched Successfully", http.StatusOK, user)
	}
}
