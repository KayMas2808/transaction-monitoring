package main

import (
	"log"
	"time"
)

const (
	maxTransactionsPerMinute = 3
	velocityCheckWindow      = 1 * time.Minute
)

func CheckVelocity(userID string) (isFraudulent bool, reason string) {
	count, err := GetTransactionCountByUserInWindow(userID, velocityCheckWindow)
	if err != nil {
		log.Printf("Error during velocity check for user %s: %v", userID, err)
		return false, ""
	}

	if (count + 1) > maxTransactionsPerMinute {
		return true, "VELOCITY_CHECK_FAILED"
	}

	return false, ""
}
