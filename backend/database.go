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

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to the database!")
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
