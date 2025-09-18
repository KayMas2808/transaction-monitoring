package main

import (
	"log"
	"time"
)

type FraudRule func(t Transaction) (bool, string)

func RunFraudChecks(t Transaction) {
	rules := []FraudRule{
		CheckVelocity,
		CheckHighValue,
		CheckGeographicInconsistency,
	}

	for _, rule := range rules {
		isFraud, ruleName := rule(t)
		if isFraud {
			log.Printf("FRAUD DETECTED: Rule '%s' failed for transaction %d (User: %s)", ruleName, t.ID, t.UserID)
			err := MarkTransactionAsFraud(t.ID)
			if err != nil {
				log.Printf("Error marking transaction %d as fraud: %v", t.ID, err)
			}

			alertPayload := map[string]interface{}{
				"rule_violated": ruleName,
				"transaction":   t,
			}
			hub.Broadcast <- WebSocketMessage{
				Type:    "fraud_alert",
				Payload: alertPayload,
			}
			return
		}
	}
}

func CheckVelocity(t Transaction) (bool, string) {
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	count, err := CountTransactionsForUser(t.UserID, oneMinuteAgo)
	if err != nil {
		log.Printf("Error checking velocity: %v", err)
		return false, ""
	}
	return count > 3, "Velocity Check"
}

func CheckHighValue(t Transaction) (bool, string) {
	const highValueThreshold = 1500.00
	return t.Amount > highValueThreshold, "High Value Transaction"
}

func CheckGeographicInconsistency(t Transaction) (bool, string) {
	sixtySecondsAgo := time.Now().Add(-60 * time.Second)
	recentLocations, err := GetRecentTransactionLocations(t.UserID, sixtySecondsAgo)
	if err != nil {
		log.Printf("Error checking geo inconsistency: %v", err)
		return false, ""
	}

	for _, loc := range recentLocations {
		if loc != "" && t.Location != loc {
			return true, "Geographic Inconsistency"
		}
	}
	return false, ""
}
