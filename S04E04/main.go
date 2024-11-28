package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", RobotInstructionsHandler)

	srv := &http.Server{
		Addr:         ":3000",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := srv.ListenAndServe()
	log.Fatal(err)
}
