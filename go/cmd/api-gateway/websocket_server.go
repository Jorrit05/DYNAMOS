package main

import (
	"fmt"
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

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

func StartWebSocketServer() {
	// Serve WebSocket endpoint at /ws
	http.HandleFunc("/ws", handleWebSocket)

	// Start HTTP server
	fmt.Println("WebSocket server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
