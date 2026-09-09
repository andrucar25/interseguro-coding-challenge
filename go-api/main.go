package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/httpapi"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/statistics"
)

const statisticsTimeout = 5 * time.Second

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	statisticsClient, err := statistics.New(os.Getenv("NODE_API_URL"), &http.Client{Timeout: statisticsTimeout})
	if err != nil {
		log.Fatalf("invalid NODE_API_URL: %v", err)
	}

	if err := httpapi.New(statisticsClient, os.Getenv("CORS_ALLOWED_ORIGINS")).Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
