package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"transaction-monitoring/auth"
	"transaction-monitoring/models"

	"github.com/gorilla/mux"
)

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(transaction); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set metadata
	transaction.IPAddress = getClientIP(r)
	transaction.UserAgent = r.UserAgent()
	transaction.CreatedAt = time.Now()

	// Run fraud detection
	fraudResult, err := h.fraudDetector.DetectFraud(r.Context(), transaction)
	if err != nil {
		h.logger.WithError(err).Error("Fraud detection failed")
		http.Error(w, "Fraud detection failed", http.StatusInternalServerError)
		return
	}

	// Set transaction status based on fraud result
	transaction.FraudScore = fraudResult.Score
	transaction.FraudReasons = fraudResult.Reasons
	if fraudResult.IsFraud {
		transaction.Status = "under_review"
	} else {
		transaction.Status = "approved"
	}

	// Insert transaction
	err = h.db.QueryRow(r.Context(),
		`INSERT INTO transactions (user_id, amount, currency, merchant_id, location, 
         ip_address, user_agent, status, fraud_score, fraud_reasons, created_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
         RETURNING id`,
		transaction.UserID, transaction.Amount, transaction.Currency, transaction.MerchantID,
		transaction.Location, transaction.IPAddress, transaction.UserAgent, transaction.Status,
		transaction.FraudScore, transaction.FraudReasons, transaction.CreatedAt).Scan(&transaction.ID)

	if err != nil {
		h.logger.WithError(err).Error("Failed to insert transaction")
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	// Create alert if fraud detected
	if fraudResult.IsFraud {
		alert := models.Alert{
			TransactionID: transaction.ID,
			AlertType:     "fraud_detection",
			Severity:      fraudResult.RiskLevel,
			Message:       fmt.Sprintf("Fraudulent transaction detected: %s", fraudResult.Reasons),
			Status:        "active",
			CreatedAt:     time.Now(),
		}

		_, err = h.db.Exec(r.Context(),
			`INSERT INTO alerts (transaction_id, alert_type, severity, message, status, created_at) 
             VALUES ($1, $2, $3, $4, $5, $6)`,
			alert.TransactionID, alert.AlertType, alert.Severity, alert.Message, alert.Status, alert.CreatedAt)

		if err != nil {
			h.logger.WithError(err).Error("Failed to create alert")
		}

		// Send WebSocket notification
		h.broadcastAlert(Alert{
			Type:          "fraud_alert",
			TransactionID: transaction.ID,
			UserID:        transaction.UserID,
			Amount:        transaction.Amount,
			FraudScore:    transaction.FraudScore,
			Reasons:       transaction.FraudReasons,
			Severity:      fraudResult.RiskLevel,
			Timestamp:     time.Now(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	status := query.Get("status")
	userID := query.Get("user_id")

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Build SQL query
	sqlQuery := `SELECT id, user_id, amount, currency, merchant_id, location, 
                 status, fraud_score, fraud_reasons, created_at, reviewed_by, reviewed_at
                 FROM transactions WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if status != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if userID != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, userID)
	}

	sqlQuery += " ORDER BY created_at DESC"

	argCount++
	sqlQuery += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)

	argCount++
	sqlQuery += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := h.db.Query(r.Context(), sqlQuery, args...)
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Amount, &t.Currency, &t.MerchantID, &t.Location,
			&t.Status, &t.FraudScore, &t.FraudReasons, &t.CreatedAt, &t.ReviewedBy, &t.ReviewedAt,
		)
		if err != nil {
			continue
		}
		transactions = append(transactions, t)
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction
	err = h.db.QueryRow(r.Context(),
		`SELECT id, user_id, amount, currency, merchant_id, location, ip_address, 
         user_agent, status, fraud_score, fraud_reasons, created_at, reviewed_by, reviewed_at
         FROM transactions WHERE id = $1`, id).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Currency,
		&transaction.MerchantID, &transaction.Location, &transaction.IPAddress, &transaction.UserAgent,
		&transaction.Status, &transaction.FraudScore, &transaction.FraudReasons, &transaction.CreatedAt,
		&transaction.ReviewedBy, &transaction.ReviewedAt,
	)

	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (h *Handler) ReviewTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=approved rejected"`
		Notes  string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	claims := auth.GetClaimsFromContext(r.Context())
	now := time.Now()

	_, err = h.db.Exec(r.Context(),
		`UPDATE transactions SET status = $1, reviewed_by = $2, reviewed_at = $3 
         WHERE id = $4`, req.Status, claims.UserID, now, id)

	if err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	// Log audit event
	h.logAuditEvent(r, claims.UserID, "review_transaction", "transaction", id, map[string]interface{}{
		"status": req.Status,
		"notes":  req.Notes,
	})

	// Update related alerts
	_, err = h.db.Exec(r.Context(),
		`UPDATE alerts SET status = 'resolved', resolved_at = $1, resolved_by = $2 
         WHERE transaction_id = $3 AND status = 'active'`, now, claims.UserID, id)

	if err != nil {
		h.logger.WithError(err).Error("Failed to update alerts")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetTransactionStats(w http.ResponseWriter, r *http.Request) {
	var stats models.TransactionStats

	// Get basic stats
	err := h.db.QueryRow(r.Context(),
		`SELECT 
            COUNT(*) as total_transactions,
            COUNT(*) FILTER (WHERE status = 'rejected' OR fraud_score >= 0.7) as fraudulent_count,
            COALESCE(SUM(amount), 0) as total_amount,
            COALESCE(AVG(amount), 0) as avg_amount
         FROM transactions 
         WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'`).Scan(
		&stats.TotalTransactions, &stats.FraudulentCount, &stats.TotalAmount, &stats.AvgAmount,
	)

	if err != nil {
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	if stats.TotalTransactions > 0 {
		stats.FraudRate = float64(stats.FraudulentCount) / float64(stats.TotalTransactions)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
