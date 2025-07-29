package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"transaction-monitoring/auth"
	"transaction-monitoring/handlers"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	if os.Getenv("ENVIRONMENT") == "development" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	db, err := initDB()
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	h := handlers.NewHandler(db, logger)
	router := setupRouter(h, logger)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.WithField("port", port).Info("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

func initDB() (*pgxpool.Pool, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbName == "" {
		dbName = "postgres"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	config.MaxConns = 30
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return db, nil
}

func setupRouter(h *handlers.Handler, logger *logrus.Logger) http.Handler {
	r := mux.NewRouter()

	r.Use(loggingMiddleware(logger))
	r.Use(recoveryMiddleware(logger))

	r.HandleFunc("/health", healthCheck).Methods("GET")
	r.HandleFunc("/api/v1/auth/login", h.Login).Methods("POST")
	r.HandleFunc("/ws", auth.JWTMiddleware(h.HandleWebSocket)).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()

	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/profile", auth.JWTMiddleware(h.GetProfile)).Methods("GET")
	authRouter.HandleFunc("/users", auth.RequireRole("admin")(h.CreateUser)).Methods("POST")

	txnRouter := api.PathPrefix("/transactions").Subrouter()
	txnRouter.HandleFunc("", auth.JWTMiddleware(h.CreateTransaction)).Methods("POST")
	txnRouter.HandleFunc("", auth.JWTMiddleware(h.GetTransactions)).Methods("GET")
	txnRouter.HandleFunc("/{id:[0-9]+}", auth.JWTMiddleware(h.GetTransaction)).Methods("GET")
	txnRouter.HandleFunc("/{id:[0-9]+}/review", auth.RequireRole("analyst")(h.ReviewTransaction)).Methods("PUT")
	txnRouter.HandleFunc("/stats", auth.JWTMiddleware(h.GetTransactionStats)).Methods("GET")

	alertRouter := api.PathPrefix("/alerts").Subrouter()
	alertRouter.HandleFunc("", auth.JWTMiddleware(h.GetAlerts)).Methods("GET")
	alertRouter.HandleFunc("/{id:[0-9]+}/resolve", auth.RequireRole("analyst")(h.ResolveAlert)).Methods("PUT")

	adminRouter := api.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/fraud-rules", auth.RequireRole("admin")(h.GetFraudRules)).Methods("GET")
	adminRouter.HandleFunc("/fraud-rules", auth.RequireRole("admin")(h.CreateFraudRule)).Methods("POST")
	adminRouter.HandleFunc("/fraud-rules/{id:[0-9]+}", auth.RequireRole("admin")(h.UpdateFraudRule)).Methods("PUT")
	adminRouter.HandleFunc("/audit-logs", auth.RequireRole("admin")(h.GetAuditLogs)).Methods("GET")
	adminRouter.HandleFunc("/connections", auth.RequireRole("admin")(h.GetActiveConnections)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	return c.Handler(r)
}

func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			logger.WithFields(logrus.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"duration": time.Since(start),
				"ip":       r.RemoteAddr,
			}).Info("HTTP request")
		})
	}
}

func recoveryMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithField("error", err).Error("Panic recovered")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
