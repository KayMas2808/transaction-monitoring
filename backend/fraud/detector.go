package fraud

import (
	"context"
	"fmt"
	"math"
	"strings"

	"transaction-monitoring/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type FraudDetector struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

type DetectionResult struct {
	IsFraud    bool     `json:"is_fraud"`
	Score      float64  `json:"score"`
	Reasons    []string `json:"reasons"`
	Confidence float64  `json:"confidence"`
	RiskLevel  string   `json:"risk_level"`
}

func NewFraudDetector(db *pgxpool.Pool, logger *logrus.Logger) *FraudDetector {
	return &FraudDetector{
		db:     db,
		logger: logger,
	}
}

func (fd *FraudDetector) DetectFraud(ctx context.Context, transaction models.Transaction) (*DetectionResult, error) {
	result := &DetectionResult{
		IsFraud:    false,
		Score:      0.0,
		Reasons:    []string{},
		Confidence: 0.0,
	}

	// Get active fraud rules
	rules, err := fd.getActiveFraudRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get fraud rules: %w", err)
	}

	// Apply each fraud detection rule
	for _, rule := range rules {
		ruleResult, err := fd.applyRule(ctx, rule, transaction)
		if err != nil {
			fd.logger.WithError(err).WithField("rule", rule.Name).Warn("Failed to apply fraud rule")
			continue
		}

		if ruleResult.triggered {
			result.Score += ruleResult.score * rule.Weight
			result.Reasons = append(result.Reasons, ruleResult.reason)
		}
	}

	// Advanced ML-like scoring (simulated)
	behaviorScore, err := fd.calculateBehaviorScore(ctx, transaction)
	if err == nil {
		result.Score += behaviorScore
	}

	// Network analysis score
	networkScore, err := fd.calculateNetworkScore(ctx, transaction)
	if err == nil {
		result.Score += networkScore
	}

	// Determine final result
	result.IsFraud = result.Score >= 0.7
	result.Confidence = math.Min(result.Score, 1.0)
	result.RiskLevel = fd.getRiskLevel(result.Score)

	fd.logger.WithFields(logrus.Fields{
		"transaction_id": transaction.ID,
		"user_id":        transaction.UserID,
		"amount":         transaction.Amount,
		"fraud_score":    result.Score,
		"is_fraud":       result.IsFraud,
		"reasons":        result.Reasons,
	}).Info("Fraud detection completed")

	return result, nil
}

type ruleResult struct {
	triggered bool
	score     float64
	reason    string
}

func (fd *FraudDetector) applyRule(ctx context.Context, rule models.FraudRule, transaction models.Transaction) (*ruleResult, error) {
	switch rule.RuleType {
	case "amount_threshold":
		return fd.checkAmountThreshold(rule, transaction), nil
	case "velocity":
		return fd.checkVelocity(ctx, rule, transaction)
	case "location":
		return fd.checkLocationAnomaly(ctx, rule, transaction)
	case "time_pattern":
		return fd.checkTimePattern(rule, transaction), nil
	case "amount_pattern":
		return fd.checkAmountPattern(rule, transaction), nil
	default:
		return &ruleResult{triggered: false}, nil
	}
}

func (fd *FraudDetector) checkAmountThreshold(rule models.FraudRule, transaction models.Transaction) *ruleResult {
	if transaction.Amount > rule.Threshold {
		return &ruleResult{
			triggered: true,
			score:     0.3,
			reason:    fmt.Sprintf("High amount transaction: $%.2f exceeds threshold $%.2f", transaction.Amount, rule.Threshold),
		}
	}
	return &ruleResult{triggered: false}
}

func (fd *FraudDetector) checkVelocity(ctx context.Context, rule models.FraudRule, transaction models.Transaction) (*ruleResult, error) {
	var count int
	err := fd.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions 
         WHERE user_id = $1 
         AND created_at >= NOW() - INTERVAL '1 minute'
         AND status != 'rejected'`, transaction.UserID).Scan(&count)

	if err != nil {
		return nil, err
	}

	if float64(count) >= rule.Threshold {
		return &ruleResult{
			triggered: true,
			score:     0.4,
			reason:    fmt.Sprintf("Velocity check triggered: %d transactions in 1 minute", count),
		}, nil
	}

	return &ruleResult{triggered: false}, nil
}

func (fd *FraudDetector) checkLocationAnomaly(ctx context.Context, rule models.FraudRule, transaction models.Transaction) (*ruleResult, error) {
	if transaction.Location == "" {
		return &ruleResult{triggered: false}, nil
	}

	// Get user's historical locations
	var historicalLocations []string
	rows, err := fd.db.Query(ctx,
		`SELECT DISTINCT location FROM transactions 
         WHERE user_id = $1 
         AND location IS NOT NULL 
         AND created_at >= NOW() - INTERVAL '30 days'
         LIMIT 10`, transaction.UserID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var location string
		if err := rows.Scan(&location); err != nil {
			continue
		}
		historicalLocations = append(historicalLocations, location)
	}

	// Simple location check (in real system, would use geolocation)
	isKnownLocation := false
	for _, loc := range historicalLocations {
		if strings.EqualFold(loc, transaction.Location) {
			isKnownLocation = true
			break
		}
	}

	if !isKnownLocation && len(historicalLocations) > 0 {
		return &ruleResult{
			triggered: true,
			score:     0.25,
			reason:    fmt.Sprintf("Transaction from unknown location: %s", transaction.Location),
		}, nil
	}

	return &ruleResult{triggered: false}, nil
}

func (fd *FraudDetector) checkTimePattern(rule models.FraudRule, transaction models.Transaction) *ruleResult {
	hour := transaction.CreatedAt.Hour()

	// Suspicious hours (2 AM - 6 AM)
	if hour >= 2 && hour <= 6 {
		return &ruleResult{
			triggered: true,
			score:     0.15,
			reason:    fmt.Sprintf("Transaction at unusual hour: %d:00", hour),
		}
	}

	return &ruleResult{triggered: false}
}

func (fd *FraudDetector) checkAmountPattern(rule models.FraudRule, transaction models.Transaction) *ruleResult {
	// Check for round amounts (potential money laundering)
	amount := transaction.Amount
	if amount >= 1000 && math.Mod(amount, 1000) == 0 {
		return &ruleResult{
			triggered: true,
			score:     0.1,
			reason:    fmt.Sprintf("Round amount pattern detected: $%.2f", amount),
		}
	}

	return &ruleResult{triggered: false}
}

func (fd *FraudDetector) calculateBehaviorScore(ctx context.Context, transaction models.Transaction) (float64, error) {
	// Simulate ML behavior analysis
	var avgAmount, stdDev float64
	var transactionCount int

	err := fd.db.QueryRow(ctx,
		`SELECT 
            COALESCE(AVG(amount), 0),
            COALESCE(STDDEV(amount), 0),
            COUNT(*)
        FROM transactions 
        WHERE user_id = $1 
        AND created_at >= NOW() - INTERVAL '30 days'
        AND status != 'rejected'`, transaction.UserID).Scan(&avgAmount, &stdDev, &transactionCount)

	if err != nil || transactionCount < 5 {
		return 0.0, err
	}

	// Calculate z-score
	if stdDev > 0 {
		zScore := math.Abs((transaction.Amount - avgAmount) / stdDev)
		if zScore > 3 {
			return 0.2, nil // Significant deviation
		} else if zScore > 2 {
			return 0.1, nil // Moderate deviation
		}
	}

	return 0.0, nil
}

func (fd *FraudDetector) calculateNetworkScore(ctx context.Context, transaction models.Transaction) (float64, error) {
	if transaction.IPAddress == "" {
		return 0.0, nil
	}

	// Check for IP address reuse across multiple users
	var userCount int
	err := fd.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) 
         FROM transactions 
         WHERE ip_address = $1 
         AND created_at >= NOW() - INTERVAL '24 hours'`, transaction.IPAddress).Scan(&userCount)

	if err != nil {
		return 0.0, err
	}

	if userCount > 5 {
		return 0.15, nil // Suspicious IP usage
	}

	return 0.0, nil
}

func (fd *FraudDetector) getRiskLevel(score float64) string {
	switch {
	case score >= 0.8:
		return "critical"
	case score >= 0.6:
		return "high"
	case score >= 0.3:
		return "medium"
	default:
		return "low"
	}
}

func (fd *FraudDetector) getActiveFraudRules(ctx context.Context) ([]models.FraudRule, error) {
	rows, err := fd.db.Query(ctx,
		`SELECT id, name, description, rule_type, parameters, threshold, weight, is_active, created_at
         FROM fraud_rules 
         WHERE is_active = true`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.FraudRule
	for rows.Next() {
		var rule models.FraudRule
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.RuleType,
			&rule.Parameters, &rule.Threshold, &rule.Weight, &rule.IsActive, &rule.CreatedAt,
		)
		if err != nil {
			continue
		}
		rules = append(rules, rule)
	}

	return rules, nil
}
