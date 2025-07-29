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

func (h *Handler) GetFraudRules(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT id, name, description, rule_type, parameters, threshold, weight, 
         is_active, created_at, updated_at
         FROM fraud_rules ORDER BY created_at DESC`)

	if err != nil {
		h.logger.WithError(err).Error("Failed to get fraud rules")
		http.Error(w, "Failed to get fraud rules", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rules []models.FraudRule
	for rows.Next() {
		var rule models.FraudRule
		var updatedAt time.Time
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.RuleType,
			&rule.Parameters, &rule.Threshold, &rule.Weight, &rule.IsActive,
			&rule.CreatedAt, &updatedAt,
		)
		if err != nil {
			continue
		}
		rules = append(rules, rule)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"fraud_rules": rules,
	})
}

func (h *Handler) CreateFraudRule(w http.ResponseWriter, r *http.Request) {
	var rule models.FraudRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if rule.Name == "" || rule.RuleType == "" {
		http.Error(w, "Name and rule_type are required", http.StatusBadRequest)
		return
	}

	rule.CreatedAt = time.Now()

	err := h.db.QueryRow(r.Context(),
		`INSERT INTO fraud_rules (name, description, rule_type, parameters, threshold, weight, is_active, created_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
         RETURNING id`,
		rule.Name, rule.Description, rule.RuleType, rule.Parameters,
		rule.Threshold, rule.Weight, rule.IsActive, rule.CreatedAt).Scan(&rule.ID)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create fraud rule")
		http.Error(w, "Failed to create fraud rule", http.StatusInternalServerError)
		return
	}

	// Log audit event
	claims := auth.GetClaimsFromContext(r.Context())
	h.logAuditEvent(r, claims.UserID, "create_fraud_rule", "fraud_rule", rule.ID, map[string]interface{}{
		"name":      rule.Name,
		"rule_type": rule.RuleType,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (h *Handler) UpdateFraudRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid fraud rule ID", http.StatusBadRequest)
		return
	}

	var rule models.FraudRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(r.Context(),
		`UPDATE fraud_rules SET name = $1, description = $2, rule_type = $3, 
         parameters = $4, threshold = $5, weight = $6, is_active = $7, updated_at = $8
         WHERE id = $9`,
		rule.Name, rule.Description, rule.RuleType, rule.Parameters,
		rule.Threshold, rule.Weight, rule.IsActive, time.Now(), id)

	if err != nil {
		h.logger.WithError(err).Error("Failed to update fraud rule")
		http.Error(w, "Failed to update fraud rule", http.StatusInternalServerError)
		return
	}

	// Log audit event
	claims := auth.GetClaimsFromContext(r.Context())
	h.logAuditEvent(r, claims.UserID, "update_fraud_rule", "fraud_rule", id, map[string]interface{}{
		"name":      rule.Name,
		"is_active": rule.IsActive,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	action := query.Get("action")
	userID := query.Get("user_id")

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Build SQL query
	sqlQuery := `SELECT al.id, al.user_id, u.username, al.action, al.resource_type, 
                 al.resource_id, al.details, al.ip_address, al.created_at
                 FROM audit_logs al
                 LEFT JOIN users u ON al.user_id = u.id
                 WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if action != "" {
		argCount++
		sqlQuery += " AND al.action = $" + strconv.Itoa(argCount)
		args = append(args, action)
	}

	if userID != "" {
		argCount++
		sqlQuery += " AND al.user_id = $" + strconv.Itoa(argCount)
		args = append(args, userID)
	}

	sqlQuery += " ORDER BY al.created_at DESC"

	argCount++
	sqlQuery += " LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	argCount++
	sqlQuery += " OFFSET $" + strconv.Itoa(argCount)
	args = append(args, offset)

	rows, err := h.db.Query(r.Context(), sqlQuery, args...)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit logs")
		http.Error(w, "Failed to get audit logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id, userIDCol, resourceID int
		var username, action, resourceType, ipAddress string
		var details map[string]interface{}
		var createdAt time.Time

		err := rows.Scan(&id, &userIDCol, &username, &action, &resourceType,
			&resourceID, &details, &ipAddress, &createdAt)
		if err != nil {
			continue
		}

		logData := map[string]interface{}{
			"id":            id,
			"user_id":       userIDCol,
			"username":      username,
			"action":        action,
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"details":       details,
			"ip_address":    ipAddress,
			"created_at":    createdAt,
		}
		logs = append(logs, logData)
	}

	response := map[string]interface{}{
		"audit_logs": logs,
		"limit":      limit,
		"offset":     offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
