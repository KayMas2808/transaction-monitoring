package main

import (
	"fmt"
	"log"
	"math"
	"time"
)

type FraudRule func(t Transaction) (bool, string)

func RunFraudChecks(t Transaction) {
	rules := []FraudRule{
		CheckVelocity,
		CheckHighValue,
		CheckGeographicInconsistency,
		CheckZScore,
	}

	go func() {
		err := AddTransactionAmount(t.UserID, t.Amount)
		if err != nil {
			log.Printf("Error adding transaction to history: %v", err)
		}
	}()

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
	count, err := IncrementVelocity(t.UserID)
	if err != nil {
		log.Printf("Error checking velocity via Redis: %v", err)
		return false, ""
	}
	return count > 5, "Velocity Check (Redis)"
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

func CheckZScore(t Transaction) (bool, string) {
	amounts, err := GetRecentAmounts(t.UserID)
	if err != nil || len(amounts) < 5 {
		return false, ""
	}

	var sum float64
	for _, a := range amounts {
		sum += a
	}
	mean := sum / float64(len(amounts))

	var varianceSum float64
	for _, a := range amounts {
		varianceSum += math.Pow(a-mean, 2)
	}
	stdDev := math.Sqrt(varianceSum / float64(len(amounts)))

	if stdDev == 0 {
		return false, ""
	}

	zScore := (t.Amount - mean) / stdDev

	if math.Abs(zScore) > 3 {
		log.Printf("Z-Score Alert: User %s, Amount %.2f, Mean %.2f, StdDev %.2f, Z %.2f", t.UserID, t.Amount, mean, stdDev, zScore)
		return true, fmt.Sprintf("Statistical Anomaly (Z-Score: %.2f)", zScore)
	}

	return false, ""
}
