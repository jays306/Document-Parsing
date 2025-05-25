package main

import (
	"context"
	"log"
	"net/http"

	"DocumentParsingSystem/pkg/api"
	"DocumentParsingSystem/pkg/config"
	"DocumentParsingSystem/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize the Gemini client
	ctx := context.Background()
	client, err := config.InitGeminiClient(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer client.Close()

	// Connect to the database
	db, err := database.ConnectDB()
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Println("The /finalize-parsed-fields endpoint will not be available.")
		db = nil
	} else {
		// Initialize the database
		if err := database.InitDB(db); err != nil {
			log.Printf("Warning: Failed to initialize database: %v", err)
			log.Println("The /finalize-parsed-fields endpoint will not be available.")
			db = nil
		} else {
			log.Println("Database connection established and initialized successfully.")
		}
		defer db.Close()
	}

	// Set up HTTP routes
	mux := http.NewServeMux()
	api.SetupRoutes(mux, client, db)

	// Start the server
	port := cfg.Port
	log.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
