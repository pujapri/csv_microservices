package main

import (
	"csv-microservice/pkg/api"
	"csv-microservice/pkg/db"
	"csv-microservice/pkg/logger"
	"log"
)

func main() {
	// Initialize logger
	logger.InitializeLogger()

	// Connect to the database
	if err := db.ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Start the server
	api.StartServer()
}
