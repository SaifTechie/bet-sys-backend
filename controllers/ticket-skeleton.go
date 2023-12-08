package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	model "github.com/SaifTechie/bet-sys-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateTicketSkeleton(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var skeletonT model.BetSkeleton
		wg.Add(1) // Increment the WaitGroup counter

		go func() {
			defer wg.Done() // Decrement the WaitGroup counter when the goroutine completes

			// Decode the request body into the user struct
			err := json.NewDecoder(r.Body).Decode(&skeletonT)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			// Get the user collection from the database
			skeletonCollection := client.Database("bet-sys").Collection("ticket-skeleton")

			// Check if a ticket skeleton with the same ID already exists
			existingSkeleton, err := GetTicketSkeletonByIDFromDB(client, skeletonT.SkeletonID)
			if err != nil && err != mongo.ErrNoDocuments {
				model.SendResponse(w, err.Error(), http.StatusInternalServerError, nil)
				return
			}

			if existingSkeleton.SkeletonID != 0 {
				model.SendResponse(w, "Ticket Skeleton with the same ID already exists", http.StatusBadRequest, nil)
				return
			}

			// Insert the user into the database
			_, err = skeletonCollection.InsertOne(r.Context(), skeletonT)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}
			model.SendResponse(w, "Ticket Skeleton Created Successfully", http.StatusOK, skeletonT)
		}()

		// Wait for the goroutine to finish before responding to the client
		wg.Wait()

	}
}

func GetTicketSkeletonByIDFromDB(client *mongo.Client, sID int) (model.BetSkeleton, error) {
	var skeletonT model.BetSkeleton

	skeletonCollection := client.Database("bet-sys").Collection("ticket-skeleton")

	err := skeletonCollection.FindOne(context.Background(), bson.M{"skeletonID": sID}).Decode(&skeletonT)
	if err != nil {
		return skeletonT, err
	}

	return skeletonT, nil
}

func GetTicketSkeletonByID(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		skeletonID := r.URL.Query().Get("sId")
		if skeletonID == "" {
			model.SendResponse(w, "Skeleton Id is required", http.StatusBadRequest, nil)
			return
		}

		intSkeletonID, err := strconv.Atoi(skeletonID)
		if err != nil {
			model.SendResponse(w, "Wrong Input Format", http.StatusBadRequest, nil)
			return
		}

		skeletonT, err := GetTicketSkeletonByIDFromDB(client, intSkeletonID)
		if err != nil {
			model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
			return
		}

		// Respond to the client
		model.SendResponse(w, "Ticket Skeleton Fetched Successfully", http.StatusOK, skeletonT)
	}
}
