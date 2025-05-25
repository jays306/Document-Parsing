package config

import (
	"context"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Config holds the application configuration
type Config struct {
	Port         string
	GeminiAPIKey string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() *Config {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if the .env file doesn't exist
		log.Println("No .env file found. Using system environment variables.")
	} else {
		log.Println("Loaded environment variables from .env file.")
	}

	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get Gemini API key from environment variable
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is not set")
	}

	return &Config{
		Port:         port,
		GeminiAPIKey: geminiAPIKey,
	}
}

// InitGeminiClient initializes the Gemini client
func InitGeminiClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return client, nil
}
