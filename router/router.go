package router

import (
	"github.com/gorilla/mux"
	controller "github.com/saifwork/bet-sys-backend/controllers"
	"go.mongodb.org/mongo-driver/mongo"
)

func Router(client *mongo.Client) *mux.Router {
	router := mux.NewRouter()

	// User
	router.HandleFunc("/api/user", controller.CreateUserHandler(client)).Methods("POST")
	router.HandleFunc("/api/user/{id}", controller.GetUserHandler(client)).Methods("GET")

	// Ticket
	router.HandleFunc("/api/ticket", controller.CreateTicketHandler(client)).Methods("POST")
	router.HandleFunc("/api/ticket-history", controller.TicketHistoryHandler(client)).Methods("GET")
	router.HandleFunc("/api/ticket-win/{id}/{lno}", controller.GetWinningTicketsHandler(client)).Methods("GET")

	// Sys-Ticket
	router.HandleFunc("/api/sys-ticket", controller.SearchSysTicketsHandler(client)).Methods("GET")

	// Ticket-Skeleton
	router.HandleFunc("/api/skeleton", controller.CreateTicketSkeleton(client)).Methods("POST")
	router.HandleFunc("/api/skeleton", controller.GetTicketSkeletonByID(client)).Methods("GET")

	return router
}
