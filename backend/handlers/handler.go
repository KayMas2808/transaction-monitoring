package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"transaction-monitoring/fraud"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	db            *pgxpool.Pool
	logger        *logrus.Logger
	validator     *validator.Validate
	fraudDetector *fraud.FraudDetector
	clients       map[*websocket.Conn]bool
	upgrader      websocket.Upgrader
}

type Alert struct {
	Type          string    `json:"type"`
	TransactionID int       `json:"transaction_id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	FraudScore    float64   `json:"fraud_score"`
	Reasons       []string  `json:"reasons"`
	Severity      string    `json:"severity"`
	Timestamp     time.Time `json:"timestamp"`
}

func NewHandler(db *pgxpool.Pool, logger *logrus.Logger) *Handler {
	return &Handler{
		db:            db,
		logger:        logger,
		validator:     validator.New(),
		fraudDetector: fraud.NewFraudDetector(db, logger),
		clients:       make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

func (h *Handler) broadcastAlert(alert Alert) {
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal alert")
		return
	}

	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, alertJSON)
		if err != nil {
			h.logger.WithError(err).Error("Failed to send WebSocket message")
			client.Close()
			delete(h.clients, client)
		}
	}
}

func (h *Handler) logAuditEvent(r *http.Request, userID int, action, resourceType string, resourceID int, details map[string]interface{}) {
	_, err := h.db.Exec(context.Background(),
		`INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, ip_address, created_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, action, resourceType, resourceID, details, getClientIP(r), time.Now())

	if err != nil {
		h.logger.WithError(err).Error("Failed to log audit event")
	}
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
