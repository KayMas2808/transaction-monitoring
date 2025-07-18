package handlers

import (
	"encoding/json"
	"net/http"

	"transaction-monitoring/auth"

	"github.com/gorilla/websocket"
)

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate WebSocket connection
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade to WebSocket")
		return
	}
	defer conn.Close()

	h.clients[conn] = true
	h.logger.WithField("user_id", claims.UserID).Info("New WebSocket connection established")

	// Send initial connection confirmation
	confirmMsg := map[string]interface{}{
		"type":    "connection",
		"status":  "connected",
		"user_id": claims.UserID,
	}

	if err := conn.WriteJSON(confirmMsg); err != nil {
		h.logger.WithError(err).Error("Failed to send connection confirmation")
		delete(h.clients, conn)
		return
	}

	// Handle incoming messages (optional - for bidirectional communication)
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			h.logger.WithError(err).Debug("WebSocket connection closed")
			delete(h.clients, conn)
			break
		}

		// Handle different message types
		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "ping":
			pongMsg := map[string]interface{}{
				"type": "pong",
			}
			if err := conn.WriteJSON(pongMsg); err != nil {
				h.logger.WithError(err).Error("Failed to send pong")
				delete(h.clients, conn)
				return
			}
		case "subscribe":
			// Handle subscription to specific alert types
			h.handleSubscription(conn, msg, claims)
		}
	}
}

func (h *Handler) handleSubscription(conn *websocket.Conn, msg map[string]interface{}, claims *auth.Claims) {
	alertTypes, ok := msg["alert_types"].([]interface{})
	if !ok {
		return
	}

	// Store subscription preferences (in a real system, this would be in database)
	subscriptionMsg := map[string]interface{}{
		"type":    "subscription_confirmed",
		"alerts":  alertTypes,
		"user_id": claims.UserID,
	}

	if err := conn.WriteJSON(subscriptionMsg); err != nil {
		h.logger.WithError(err).Error("Failed to send subscription confirmation")
	}
}

func (h *Handler) BroadcastAlert(alert Alert) {
	h.broadcastAlert(alert)
}

func (h *Handler) GetActiveConnections(w http.ResponseWriter, r *http.Request) {
	connectionCount := len(h.clients)

	response := map[string]interface{}{
		"active_connections": connectionCount,
		"timestamp":          "now",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
