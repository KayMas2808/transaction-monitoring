package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}
	return db.Ping()
}

func SaveTransaction(t Transaction) (int, error) {
	var id int
	err := db.QueryRow(
		"INSERT INTO transactions (user_id, amount, description, timestamp) VALUES ($1, $2, $3, $4) RETURNING id",
		t.UserID, t.Amount, t.Description, t.Timestamp,
	).Scan(&id)

	if err != nil {
		log.Printf("Error saving transaction: %v", err)
		return 0, err
	}
	return id, nil
}

func GetTransactionCountByUserInWindow(userID string, window time.Duration) (int, error) {
	var count int
	since := time.Now().Add(-window)

	err := db.QueryRow(
		"SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND timestamp >= $2",
		userID, since,
	).Scan(&count)

	if err != nil {
		log.Printf("Error getting transaction count for user %s: %v", userID, err)
		return 0, err
	}
	return count, nil
}
