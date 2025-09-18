package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	err := InitDB("postgres://user:mysecretpassword@localhost:5432/fraud_detection?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Successfully connected to database!")

	r := mux.NewRouter()

	hub := newHub()
	go hub.run()

	r.HandleFunc("/transaction", handleNewTransaction(hub)).Methods("POST")

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			if r.Method == "OPTIONS" {
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", corsHandler(r)); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
