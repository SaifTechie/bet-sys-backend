// socket/socket.go

package socket

import (
	"log"

	socketio "github.com/googollee/go-socket.io"
	controller "github.com/saifwork/bet-sys-backend/controllers"
	"go.mongodb.org/mongo-driver/mongo"
)

// InitSocket initializes the Socket.IO server
func InitSocket(server *socketio.Server, client *mongo.Client) {

	// Handle Socket.IO connection event
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("Socket.IO connected:", s.ID())
		return nil
	})

	// Handle Socket.IO disconnection event
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("Socket.IO disconnected:", s.ID(), reason)
	})

	// Handle "get-sys-current-ticket" event
	server.OnEvent("/", "get-sys-current-ticket", func(s socketio.Conn, params map[string]interface{}) {
		log.Println("Received 'get-sys-current-ticket' event from:", s.ID())

		// Extract parameters from the map
		lotteryID, _ := params["lotteryID"].(string)
		price, _ := params["price"].(float64)
		date, _ := params["date"].(string)
		time, _ := params["time"].(string)
		visible, _ := params["visible"].(string)

		// Call the controller method and send the response back
		result, _ := controller.SearchTickets(client, lotteryID, price, date, time, visible)
		s.Emit("get-sys-current-ticket-response", result)
	})
}
