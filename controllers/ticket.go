package controller

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	model "github.com/SaifTechie/bet-sys-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateTicketHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var ticket model.History

		wg.Add(1) // Increment the WaitGroup counter

		go func() {
			defer wg.Done() // Decrement the WaitGroup counter when the goroutine completes

			// Decode the request body into the ticket struct
			err := json.NewDecoder(r.Body).Decode(&ticket)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			// Convert the UserID from string to ObjectID
			userID, err := primitive.ObjectIDFromHex(ticket.UserID.Hex())
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			lID, err := primitive.ObjectIDFromHex(ticket.LotteryID.Hex())
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			// Check if the associated lottery document is visible
			visible, err := checkLotteryVisibility(client, lID)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			fmt.Println("lottery is not expired : ", visible)

			if !visible {
				model.SendResponse(w, "Lottery is Expired", http.StatusBadRequest, nil)
				return
			}

			ticket.UserID = userID
			ticket.LotteryID = lID

			// Set the timestamp for the ticket
			ticket.Timestamp = time.Now()
			ticket.IsWinner = false

			ticketCollection := client.Database("bet-sys").Collection("tickets")

			// Insert the ticket into the database
			_, err = ticketCollection.InsertOne(r.Context(), ticket)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}
			model.SendResponse(w, "Ticket Created Successfully", http.StatusOK, ticket)
		}()

		// Wait for the goroutine to finish before responding to the client
		wg.Wait()
	}
}

func TicketHistoryHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var tickets []model.History

		wg.Add(1)
		go func() {
			defer wg.Done()
			// Get the user ID and lottery number from the request URL
			// vars := mux.Vars(r)
			// userID := vars["id"]
			// lotteryNum := vars["lno"]

			userID := r.URL.Query().Get("uid")
			lotteryNum := r.URL.Query().Get("lid")
			isWinner := r.URL.Query().Get("isWinner")

			ticketCollection := client.Database("bet-sys").Collection("tickets")

			filter := bson.M{}

			if userID != "" {
				uid, _ := primitive.ObjectIDFromHex(userID)
				filter["userID"] = uid
				fmt.Println(uid)
			}

			if lotteryNum != "" {
				lid, _ := primitive.ObjectIDFromHex(lotteryNum)
				fmt.Println(lid)
				filter["lotteryID"] = lid
			}

			if isWinner != "" {
				filter["isWinner"] = true
			}

			cursor, err := ticketCollection.Find(r.Context(), filter)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}
			defer cursor.Close(r.Context())

			for cursor.Next(r.Context()) {
				var ticket model.History
				if err := cursor.Decode(&ticket); err != nil {
					model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
					return
				}
				tickets = append(tickets, ticket)
			}

		}()

		wg.Wait()

		model.SendResponse(w, "Ticket History Fetched", http.StatusOK, tickets)
	}
}

func GetWinningTicketsHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var winningTickets []model.History

		wg.Add(1)
		go func() {
			defer wg.Done()

			userID := r.URL.Query().Get("uid")
			lotteryNum := r.URL.Query().Get("lid")

			ticketCollection := client.Database("bet-sys").Collection("tickets")

			uid, _ := primitive.ObjectIDFromHex(userID)
			filter := bson.M{"userID": uid, "isWinner": true}

			fmt.Println(uid)
			if lotteryNum != "" {
				lid, _ := primitive.ObjectIDFromHex(lotteryNum)
				fmt.Println(lid)
				filter["lotteryID"] = lid
			}

			cursor, err := ticketCollection.Find(r.Context(), filter)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}
			defer cursor.Close(r.Context())

			for cursor.Next(r.Context()) {
				var ticket model.History
				if err := cursor.Decode(&ticket); err != nil {
					model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
					return
				}
				winningTickets = append(winningTickets, ticket)
			}

		}()

		wg.Wait()

		model.SendResponse(w, "Winning Records Fetched", http.StatusOK, winningTickets)
	}
}

func SchedulePrintStatement(client *mongo.Client) {
	fmt.Println("Scheduling print statement every 1 minute")

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	ticker := time.NewTicker(time.Minute)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				fmt.Println("Printing a statement every 1 minute")
			case <-ctx.Done():
				fmt.Println("Context done. Exiting the goroutine.")
				return
			}
		}
	}()

	wg.Wait()

	defer func() {
		ticker.Stop()
		cancel()
	}()
}

func ScheduleNewBet(client *mongo.Client) {
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	ticker := time.NewTicker(time.Minute)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker.C:

				err := AddNewBetToDatabase(client, ctx, 2)
				if err != nil {
					fmt.Println("Error adding NewBet:", err)
				}
			case <-ctx.Done():
				fmt.Println("Context done. Exiting the goroutine.")
				return
			}
		}
	}()

	wg.Wait()

	defer func() {
		ticker.Stop()
		cancel()
	}()
}

func AddNewBetToDatabase(client *mongo.Client, ctx context.Context, skeletonId int) error {

	skeletonT, err := GetTicketSkeletonByIDFromDB(client, skeletonId)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	deadlineTime := skeletonT.DeadlineTime
	duration := time.Duration(deadlineTime * float64(time.Minute))
	globalDeadlineTime := duration
	currentBet := model.NewBet{
		LotteryID: primitive.NewObjectID(),
		Numbers:   generateRandomNumbers(2),
		Price:     skeletonT.Price,
		Date:      currentTime.Format("02:01:2006"),
		Time:      currentTime.Format("03:04 PM"),
		MinTicket: skeletonT.MinTicket,
		Visible:   true,
	}

	currentBet.SetDeadline(duration)

	currentBetCollection := client.Database("bet-sys").Collection("sys-tickets")

	_, errr := currentBetCollection.InsertOne(ctx, currentBet)
	if errr != nil {
		fmt.Println("Error inserting NewBet into the database:", errr)
		return errr
	}

	fmt.Println("A new bet gets created")
	fmt.Println(currentBet)

	resultCh := make(chan []model.History)
	errCh := make(chan error)

	go func() {
		for {
			select {
			case <-time.After(currentBet.Deadline):
				fmt.Println(currentBet.LotteryID)
				fmt.Println(currentBet.Numbers)
				fmt.Println("Deadline reached. Execute statement.")

				if countTicketsByLotteryID(client, ctx, currentBet.LotteryID, skeletonT.MinTicket) {

					fmt.Println("inside if for announcing winner")
					err := updateSysTicketVisibility(client, ctx, currentBet.LotteryID)
					if err != nil {
						fmt.Println("Error getting tickets:", err)
						errCh <- err
						return
					}

					result, err := GetTicketsByLotteryIDNew(client, ctx, currentBet.LotteryID, currentBet.Numbers)()
					if err != nil {
						fmt.Println("Error getting tickets:", err)
						errCh <- err
						return
					}

					resultCh <- result
					return // Exit the goroutine after processing
				} else {

					fmt.Println("inside else for increasing the time")
					addOnTimer := 1 * time.Minute

					globalDeadlineTime += addOnTimer
					newDeadline := globalDeadlineTime
					// currentBet.SetDeadline(newDeadline)
					// currentBet.SetDeadline(currentBet.Deadline)

					err := updateSysTicketDeadlineTime(client, ctx, currentBet.LotteryID, newDeadline)
					if err != nil {
						fmt.Println("Error getting tickets:", err)
						errCh <- err
						return
					}
				}

				// err := updateSysTicket(client, ctx, currentBet.LotteryID)
				// if err != nil {
				// 	fmt.Println("Error getting tickets:", err)
				// 	errCh <- err
				// 	return
				// }

				// result, err := GetTicketsByLotteryIDNew(client, ctx, currentBet.LotteryID, currentBet.Numbers)()

				// if err != nil {
				// 	fmt.Println("Error getting tickets:", err)
				// 	errCh <- err
				// 	return
				// }

				// resultCh <- result

			case <-ctx.Done():
				fmt.Println("Context done. Exiting the monitoring goroutine.")
				return
			}
		}
	}()

	// Wait for either the result or an error
	select {
	case result := <-resultCh:
		fmt.Println("Got result:", result)
		if len(result) > 0 {
			findWinner(result, currentBet.Numbers, client, ctx)
		} else {
			fmt.Println("User doesn't buy any tickets")
		}

	case err := <-errCh:
		fmt.Println("Got error:", err)
		return err
	}

	return nil
}

func generateRandomNumbers(n int) []int {
	numbers := make([]int, n)

	for i := 0; i < n; i++ {
		var num uint32
		err := binary.Read(rand.Reader, binary.BigEndian, &num)
		if err != nil {
			fmt.Println("Error generating random number:", err)
			return nil
		}
		numbers[i] = int(num % 10)
	}

	return numbers
}

func findWinner(results []model.History, systemNum []int, client *mongo.Client, ctx context.Context) {
	fmt.Println("Processing results:")
	for _, result := range results {
		fmt.Printf("ID: %s, UserID: %s, LotteryID: %s, SelectedNumbers: %v, IsWinner: %t, Timestamp: %s\n",
			result.ID, result.UserID, result.LotteryID, result.SelectedNumbers, result.IsWinner, result.Timestamp)

		if slicesMatch(result.SelectedNumbers, systemNum) {
			fmt.Println(result.UserID)
			fmt.Println("Winner found!")
			err := updateWinnerStatus(client, ctx, result.ID)
			if err != nil {
				fmt.Println("Error updating winner status:", err)
			}
		}
	}
}

func updateSysTicketVisibility(client *mongo.Client, ctx context.Context, dID primitive.ObjectID) error {

	fmt.Println("update sys called")

	ticketsCollection := client.Database("bet-sys").Collection("sys-tickets")

	filter := bson.M{"lotteryID": dID}

	update := bson.M{"$set": bson.M{"visible": false}}

	_, err := ticketsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Error updating Sys-Ticket status:", err)
		return err
	}

	return nil
}

func updateSysTicketDeadlineTime(client *mongo.Client, ctx context.Context, dID primitive.ObjectID, duration time.Duration) error {

	fmt.Println("update sys deadline time")

	ticketsCollection := client.Database("bet-sys").Collection("sys-tickets")

	filter := bson.M{"lotteryID": dID}

	update := bson.M{"$set": bson.M{"deadline": duration}}

	_, err := ticketsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Error updating Sys-Ticket Deadline time:", err)
		return err
	}

	return nil
}

func updateWinnerStatus(client *mongo.Client, ctx context.Context, dID primitive.ObjectID) error {
	ticketsCollection := client.Database("bet-sys").Collection("tickets")

	filter := bson.M{"_id": dID}

	update := bson.M{"$set": bson.M{"isWinner": true}}

	_, err := ticketsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Error updating winner status:", err)
		return err
	}

	return nil
}

func slicesMatch(slice1, slice2 []int) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func GetTicketsByLotteryIDNew(client *mongo.Client, ctx context.Context, lotteryID primitive.ObjectID, numbers []int) func() ([]model.History, error) {
	return func() ([]model.History, error) {
		var wg sync.WaitGroup
		var tickets []model.History

		wg.Add(1)
		go func() {
			defer wg.Done()
			ticketCollection := client.Database("bet-sys").Collection("tickets")

			filter := bson.M{"lotteryID": lotteryID}
			cursor, err := ticketCollection.Find(ctx, filter)
			if err != nil {
				return
			}

			defer cursor.Close(ctx)

			for cursor.Next(ctx) {
				var ticket model.History
				if err := cursor.Decode(&ticket); err != nil {
					return
				}
				tickets = append(tickets, ticket)
			}

		}()

		wg.Wait()
		return tickets, nil
	}
}

func SearchTickets(client *mongo.Client, lotteryID string, price float64, date, time string, visible string) ([]model.NewBet, error) {
	var tickets []model.NewBet

	// Create a filter for the query
	filter := bson.M{}

	if lotteryID != "" {
		oid, err := primitive.ObjectIDFromHex(lotteryID)
		if err != nil {
			return nil, err
		}
		filter["lotteryID"] = oid
	}

	if price != 0 {
		filter["price"] = price
	}

	if date != "" {
		filter["date"] = date
	}

	if time != "" {
		filter["time"] = time
	}

	fmt.Println(visible)

	// Only include "visible" in the filter if it's true
	if visible != "" {
		v, err := strconv.ParseBool(visible)
		if err != nil {
			return nil, err
		}

		fmt.Println(v)
		filter["visible"] = v
	}

	fmt.Println(filter)

	// Execute the query
	cursor, err := client.Database("bet-sys").Collection("sys-tickets").Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode the results
	for cursor.Next(context.Background()) {
		var ticket model.NewBet
		if err := cursor.Decode(&ticket); err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func SearchSysTicketsHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get parameters from the query string
		lotteryID := r.URL.Query().Get("lotteryID")
		priceStr := r.URL.Query().Get("price")
		date := r.URL.Query().Get("date")
		time := r.URL.Query().Get("time")
		visibleStr := r.URL.Query().Get("visible")

		var wg sync.WaitGroup
		var tickets []model.NewBet

		// Convert price from string to float64, default to 0 if not provided
		var price float64
		if priceStr != "" {
			p, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				model.SendResponse(w, "Invalid price format", http.StatusBadRequest, nil)
				return
			}
			price = p
		}

		// Convert visible from string to bool if provided
		// var visible bool
		// if visibleStr != "" {
		// 	v, err := strconv.ParseBool(visibleStr)
		// 	if err != nil {
		// 		model.SendResponse(w, "Invalid visible format", http.StatusBadRequest, nil)
		// 		return
		// 	}
		// 	visible = v
		// }

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Call the standalone function
			t, err := SearchTickets(client, lotteryID, price, date, time, visibleStr)
			if err != nil {
				model.SendResponse(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			// Append the results to the outer scope variable
			tickets = append(tickets, t...)
		}()

		// Wait for the goroutine to finish
		wg.Wait()

		model.SendResponse(w, "Tickets Fetched", http.StatusOK, tickets)
	}
}

func checkLotteryVisibility(client *mongo.Client, lotteryID primitive.ObjectID) (bool, error) {

	lotteryCollection := client.Database("bet-sys").Collection("sys-tickets")

	filter := bson.M{"lotteryID": lotteryID, "visible": true}

	var result bson.M
	err := lotteryCollection.FindOne(context.Background(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {

		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func countTicketsByLotteryID(client *mongo.Client, ctx context.Context, lotteryID primitive.ObjectID, minTicketSize int) bool {

	ticketsCollection := client.Database("bet-sys").Collection("tickets")

	// Create a filter to match documents with the specified lottery ID
	filter := bson.M{"lotteryID": lotteryID}

	// Count the number of documents that match the filter
	count, err := ticketsCollection.CountDocuments(ctx, filter)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("total ticket required is", minTicketSize)
	fmt.Println("total ticket count", count)
	fmt.Println("return type", minTicketSize >= int(count))

	return int(count) >= minTicketSize
}
