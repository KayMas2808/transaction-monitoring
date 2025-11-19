package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var hub = newHub()

func main() {
	InitRedis()

	InitDB("postgres://user:mysecretpassword@postgres:5432/fraud_detection?sslmode=disable")

	r := mux.NewRouter()

	go hub.run()

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	r.HandleFunc("/api/transactions", GetTransactions).Methods("GET")
	r.HandleFunc("/api/simulate", SimulateTransaction).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
