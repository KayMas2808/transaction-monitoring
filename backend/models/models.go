package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username" validate:"required,min=3"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Password  string    `json:"-" db:"password_hash"`
	Role      string    `json:"role" db:"role" validate:"required,oneof=admin analyst viewer"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Transaction struct {
	ID           int        `json:"id" db:"id"`
	UserID       int        `json:"user_id" db:"user_id" validate:"required"`
	Amount       float64    `json:"amount" db:"amount" validate:"required,gt=0"`
	Currency     string     `json:"currency" db:"currency" validate:"required,len=3"`
	MerchantID   string     `json:"merchant_id" db:"merchant_id"`
	Location     string     `json:"location" db:"location"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	Status       string     `json:"status" db:"status"`
	FraudScore   float64    `json:"fraud_score" db:"fraud_score"`
	FraudReasons []string   `json:"fraud_reasons" db:"fraud_reasons"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ReviewedBy   *int       `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt   *time.Time `json:"reviewed_at" db:"reviewed_at"`
}

type Alert struct {
	ID            int        `json:"id" db:"id"`
	TransactionID int        `json:"transaction_id" db:"transaction_id"`
	AlertType     string     `json:"alert_type" db:"alert_type"`
	Severity      string     `json:"severity" db:"severity"`
	Message       string     `json:"message" db:"message"`
	Status        string     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at" db:"resolved_at"`
}

type FraudRule struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	RuleType    string    `json:"rule_type" db:"rule_type"`
	Parameters  string    `json:"parameters" db:"parameters"`
	Threshold   float64   `json:"threshold" db:"threshold"`
	Weight      float64   `json:"weight" db:"weight"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type TransactionStats struct {
	TotalTransactions int     `json:"total_transactions"`
	FraudulentCount   int     `json:"fraudulent_count"`
	FraudRate         float64 `json:"fraud_rate"`
	TotalAmount       float64 `json:"total_amount"`
	AvgAmount         float64 `json:"avg_amount"`
}
