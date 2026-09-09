package main

import (
	"log"
	"net/http"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/config"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/external/statistics"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/handler"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/usecase"
)

func main() {
	configuration := config.Load()

	statisticsClient, err := statistics.New(configuration.NodeAPIURL, &http.Client{Timeout: configuration.StatisticsTimeout})
	if err != nil {
		log.Fatalf("invalid NODE_API_URL: %v", err)
	}
	service, err := usecase.New(statisticsClient)
	if err != nil {
		log.Fatalf("create QR factorization service: %v", err)
	}

	if err := handler.New(service, configuration.CORSAllowedOrigins).Listen(":" + configuration.Port); err != nil {
		log.Fatal(err)
	}
}
