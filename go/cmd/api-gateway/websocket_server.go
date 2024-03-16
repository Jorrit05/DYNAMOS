package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections
		return true
	},
}

func handleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading connection to WebSocket:", err)
			return
		}
		defer conn.Close()

		for {
			// Read message from WebSocket client
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message from WebSocket:", err)
				break
			}

			// Print received message
			log.Printf("Received message: %s", message)

			// Echo back the received message
			if err := conn.WriteMessage(messageType, message); err != nil {
				log.Println("Error echoing message to WebSocket:", err)
				break
			}
		}
	}
}
