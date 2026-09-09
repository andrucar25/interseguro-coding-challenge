package main

import (
	"log"
	"os"

	"github.com/andrucar25/interseguro-coding-challenge/go-api/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if err := httpapi.New().Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
