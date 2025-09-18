package main

import "time"

type Transaction struct {
	ID              int       `json:"id"`
	UserID          string    `json:"user_id"`
	Amount          float64   `json:"amount"`
	CardNumber      string    `json:"card_number"`
	MerchantDetails string    `json:"merchant_details"`
	Location        string    `json:"location"` // NEW: For geographic rules
	IsFraud         bool      `json:"is_fraud"` // NEW: To flag a transaction's state
	CreatedAt       time.Time `json:"created_at"`
}

type FraudAlert struct {
	ID          string      `json:"id"`
	Transaction Transaction `json:"transaction"`
	RuleName    string      `json:"rule_name"`
	Details     string      `json:"details"`
	Timestamp   time.Time   `json:"timestamp"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
