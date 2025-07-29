package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"transaction-monitoring/auth"
	"transaction-monitoring/models"

	"github.com/gorilla/mux"
)

func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	status := query.Get("status")
	severity := query.Get("severity")

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Build SQL query
	sqlQuery := `SELECT a.id, a.transaction_id, a.alert_type, a.severity, a.message, 
                 a.status, a.created_at, a.resolved_at, a.resolved_by,
                 t.user_id, t.amount, t.currency
                 FROM alerts a
                 JOIN transactions t ON a.transaction_id = t.id
                 WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if status != "" {
		argCount++
		sqlQuery += " AND a.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
	}

	if severity != "" {
		argCount++
		sqlQuery += " AND a.severity = $" + strconv.Itoa(argCount)
		args = append(args, severity)
	}

	sqlQuery += " ORDER BY a.created_at DESC"

	argCount++
	sqlQuery += " LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	argCount++
	sqlQuery += " OFFSET $" + strconv.Itoa(argCount)
	args = append(args, offset)

	rows, err := h.db.Query(r.Context(), sqlQuery, args...)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get alerts")
		http.Error(w, "Failed to get alerts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var alerts []map[string]interface{}
	for rows.Next() {
		var alert models.Alert
		var userID int
		var amount float64
		var currency string

		err := rows.Scan(
			&alert.ID, &alert.TransactionID, &alert.AlertType, &alert.Severity,
			&alert.Message, &alert.Status, &alert.CreatedAt, &alert.ResolvedAt,
			&alert.ResolvedAt, &userID, &amount, &currency,
		)
		if err != nil {
			continue
		}

		alertData := map[string]interface{}{
			"id":             alert.ID,
			"transaction_id": alert.TransactionID,
			"alert_type":     alert.AlertType,
			"severity":       alert.Severity,
			"message":        alert.Message,
			"status":         alert.Status,
			"created_at":     alert.CreatedAt,
			"resolved_at":    alert.ResolvedAt,
			"user_id":        userID,
			"amount":         amount,
			"currency":       currency,
		}
		alerts = append(alerts, alertData)
	}

	response := map[string]interface{}{
		"alerts": alerts,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=resolved false_positive"`
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
		`UPDATE alerts SET status = $1, resolved_at = $2, resolved_by = $3 
         WHERE id = $4`, req.Status, now, claims.UserID, id)

	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve alert")
		http.Error(w, "Failed to resolve alert", http.StatusInternalServerError)
		return
	}

	// Log audit event
	h.logAuditEvent(r, claims.UserID, "resolve_alert", "alert", id, map[string]interface{}{
		"status": req.Status,
		"notes":  req.Notes,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
