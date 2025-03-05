package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var db *pgxpool.Pool

func initDB() {
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://transaction_user:transaction_password@localhost:5432/postgres")
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	log.Println("Connected to PostgreSQL")
}

// Fraud Detection Function
func detectFraud(userID int, amount float64) bool {
	// Rule 1: Flag transactions above $10,000
	if amount > 10000 {
		log.Println("High-Value Transaction Flagged:", amount)
		return true
	}

	// Rule 2: Velocity Check - More than 3 transactions in 1 minute
	var count int
	err := db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM transactions 
		 WHERE user_id = $1 
		 AND created_at >= NOW() - INTERVAL '1 minute'`, userID).Scan(&count)

	if err != nil {
		log.Println("Error checking velocity rule:", err)
		return false
	}

	if count >= 3 {
		log.Println("Velocity Check Flagged for user:", userID)
		return true
	}

	return false
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()
	clients[conn] = true
	log.Println("New WebSocket connection established")

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading message:", err)
			delete(clients, conn)
			break
		}

		// Extract transaction details
		userID := int(msg["user_id"].(float64))
		amount := msg["amount"].(float64)
		currency := msg["currency"].(string)

		// Check for fraud
		isFraud := detectFraud(userID, amount)

		// Save transaction with status
		status := "approved"
		if isFraud {
			for client := range clients {
				alert := map[string]string{
					"alert":   "Suspicious transaction detected!",
					"user_id": fmt.Sprintf("%d", userID),
					"amount":  fmt.Sprintf("%.2f", amount),
					"status":  status,
				}

				log.Println("📤 Sending WebSocket alert:", alert) // ✅ Log the outgoing alert

				err := client.WriteJSON(alert)
				if err != nil {
					log.Println("Error sending WebSocket alert:", err)
					client.Close()
					delete(clients, client)
				}
			}
		}

		_, err = db.Exec(context.Background(),
			"INSERT INTO transactions (user_id, amount, currency, status) VALUES ($1, $2, $3, $4)",
			userID, amount, currency, status)

		if err != nil {
			log.Println("Error inserting transaction:", err)
		} else {
			log.Println("Transaction saved:", msg, "Status:", status)
		}

		// Notify WebSocket clients if fraud is detected
		if isFraud {
			for client := range clients {
				err := client.WriteJSON(map[string]string{
					"alert":   "Suspicious transaction detected!",
					"user_id": fmt.Sprintf("%d", userID),
					"amount":  fmt.Sprintf("%.2f", amount),
					"status":  status,
				})
				if err != nil {
					log.Println("Error sending WebSocket alert:", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

func main() {
	initDB()

	http.HandleFunc("/ws", handleConnections)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
