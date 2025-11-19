package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB(dataSourceName string) {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			fmt.Println("Successfully connected to the database!")

			createTableQuery := `
			CREATE TABLE IF NOT EXISTS transactions (
				id SERIAL PRIMARY KEY,
				user_id VARCHAR(50) NOT NULL,
				amount DECIMAL(10, 2) NOT NULL,
				card_number VARCHAR(20) NOT NULL,
				merchant_details VARCHAR(100) NOT NULL,
				location VARCHAR(100) NOT NULL,
				is_fraud BOOLEAN DEFAULT FALSE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`

			if _, err := db.Exec(createTableQuery); err != nil {
				log.Fatalf("Failed to create transactions table: %v", err)
			}
			fmt.Println("Transactions table initialized.")
			return
		}
		log.Printf("Failed to connect to database (attempt %d/30): %v", i+1, err)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("Could not connect to database after 30 seconds: %v", err)
}

func CreateTransaction(t *Transaction) (*Transaction, error) {
	query := `INSERT INTO transactions (user_id, amount, card_number, merchant_details, location, is_fraud, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := db.QueryRow(query, t.UserID, t.Amount, t.CardNumber, t.MerchantDetails, t.Location, t.IsFraud, t.CreatedAt).Scan(&t.ID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func MarkTransactionAsFraud(transactionID int) error {
	query := `UPDATE transactions SET is_fraud = TRUE WHERE id = $1`
	_, err := db.Exec(query, transactionID)
	return err
}

func CountTransactionsForUser(userID string, since time.Time) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND created_at >= $2`
	err := db.QueryRow(query, userID, since).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetRecentTransactionLocations(userID string, since time.Time) ([]string, error) {
	var locations []string
	query := `SELECT location FROM transactions WHERE user_id = $1 AND created_at >= $2`
	rows, err := db.Query(query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var location string
		if err := rows.Scan(&location); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}
	return locations, nil
}

func GetRecentTransactions(limit int) ([]Transaction, error) {
	var transactions []Transaction
	query := `SELECT id, user_id, amount, card_number, merchant_details, location, is_fraud, created_at FROM transactions ORDER BY created_at DESC LIMIT $1`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.CardNumber, &t.MerchantDetails, &t.Location, &t.IsFraud, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
