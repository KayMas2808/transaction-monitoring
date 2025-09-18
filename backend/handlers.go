package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func handleTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var t Transaction
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		log.Printf("Error decoding transaction JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.CreatedAt = time.Now()
	t.IsFraud = false

	_, err = CreateTransaction(&t)
	if err != nil {
		log.Printf("Error creating transaction in DB: %v", err)
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received transaction: User %s, Amount %.2f, Location %s\n", t.UserID, t.Amount, t.Location)

	go RunFraudChecks(t)

	hub.Broadcast <- WebSocketMessage{
		Type:    "new_transaction",
		Payload: t,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "transaction_id": fmt.Sprintf("%d", t.ID)})
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan WebSocketMessage, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
