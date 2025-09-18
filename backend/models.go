package main

import "time"

type Transaction struct {
	ID          int       `json:"id"`
	UserID      string    `json:"userId"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

type FraudAlert struct {
	AlertID       string    `json:"alertId"`
	TransactionID int       `json:"transactionId"`
	UserID        string    `json:"userId"`
	Timestamp     time.Time `json:"timestamp"`
	Reason        string    `json:"reason"`
	Details       string    `json:"details"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
