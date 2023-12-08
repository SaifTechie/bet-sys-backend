package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	controller "github.com/SaifTechie/bet-sys-backend/controllers"
	database "github.com/SaifTechie/bet-sys-backend/database"
	router "github.com/SaifTechie/bet-sys-backend/router"
	socket "github.com/SaifTechie/bet-sys-backend/socket"
	socketio "github.com/googollee/go-socket.io"
	"github.com/rs/cors" // Import the cors package
)

func main() {
	// Connect to the database
	client, ctx, cancel, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer cancel()

	// Disconnect from the database when main function exits
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("Error disconnecting from the database:", err)
		}
	}()

	// Initialize wait group
	var wg sync.WaitGroup

	// Create a new Socket.IO server
	server := socketio.NewServer(nil)
	socket.InitSocket(server, client)

	// Schedule a new bet in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		controller.ScheduleNewBet(client)
	}()

	// Create a new router
	router := router.Router(client)
	router.Handle("/socket.io/", server)

	// Create a new CORS handler
	corsHandler := cors.Default()

	// Wrap the router with the CORS handler
	handler := corsHandler.Handler(router)

	// Start the server with the new handler
	port := 8080
	fmt.Printf("Server started on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))

	// Wait for all goroutines to finish
	wg.Wait()
}
