package handlers

import (
	"encoding/json"
	"net/http"

	"transaction-monitoring/auth"
	"transaction-monitoring/models"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get user from database
	var user models.User
	var hashedPassword string
	err := h.db.QueryRow(r.Context(),
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
         FROM users WHERE username = $1`, req.Username).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Log audit event
	h.logAuditEvent(r, user.ID, "login", "user", user.ID, map[string]interface{}{
		"username": user.Username,
	})

	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := h.db.QueryRow(r.Context(),
		`SELECT id, username, email, role, created_at, updated_at 
         FROM users WHERE id = $1`, claims.UserID).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(user); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Insert user
	err = h.db.QueryRow(r.Context(),
		`INSERT INTO users (username, email, password_hash, role) 
         VALUES ($1, $2, $3, $4) 
         RETURNING id, created_at, updated_at`,
		user.Username, user.Email, string(hashedPassword), user.Role).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			http.Error(w, "Username or email already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	// Log audit event
	claims := auth.GetClaimsFromContext(r.Context())
	if claims != nil {
		h.logAuditEvent(r, claims.UserID, "create_user", "user", user.ID, map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
		})
	}

	user.Password = "" // Don't return password
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
