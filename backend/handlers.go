package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func handleNewTransaction(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Transaction
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		t.Timestamp = time.Now()

		isFraud, reason := CheckVelocity(t.UserID)
		if isFraud {
			alert := FraudAlert{
				AlertID:       "ALERT-" + t.UserID + "-" + t.Timestamp.Format(time.RFC3339),
				TransactionID: 0, // We don't have the ID yet
				UserID:        t.UserID,
				Timestamp:     time.Now(),
				Reason:        reason,
				Details:       "High transaction frequency detected.",
			}
			alertMessage, _ := json.Marshal(WebSocketMessage{Type: "fraud_alert", Payload: alert})
			hub.broadcast <- alertMessage
			log.Printf("FRAUD DETECTED: %s for User: %s", reason, t.UserID)
		}

		id, err := SaveTransaction(t)
		if err != nil {
			http.Error(w, "Failed to save transaction", http.StatusInternalServerError)
			return
		}
		t.ID = id
		message, err := json.Marshal(WebSocketMessage{Type: "new_transaction", Payload: t})
		if err != nil {
			log.Printf("Error marshalling transaction: %v", err)
		} else {
			hub.broadcast <- message
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	}
}
