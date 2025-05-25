package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"unicode/utf8"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// JobDetails represents the structured response format for parsed job information
type JobDetails struct {
	Title          string `json:"title"`
	Salary         string `json:"salary"`
	Location       string `json:"location"`
	Experience     string `json:"experience"`
	EmploymentType string `json:"employment-type"`
}

// sanitizeUTF8 removes invalid UTF-8 sequences from a byte slice and returns a valid UTF-8 string
func sanitizeUTF8(data []byte) string {
	// Create a buffer to hold the sanitized string
	var sanitized []byte

	// Process the data in chunks
	for len(data) > 0 {
		// Check if the next rune is valid
		r, size := utf8.DecodeRune(data)

		// If it's a valid UTF-8 sequence, add it to the sanitized buffer
		if r != utf8.RuneError || size > 1 {
			sanitized = append(sanitized, data[:size]...)
		}

		// Move to the next chunk
		data = data[size:]
	}

	return string(sanitized)
}

// parseDocumentWithGemini uses the Gemini AI API to extract job details from a document
func parseDocumentWithGemini(ctx context.Context, client *genai.Client, documentContent string) (JobDetails, error) {
	// Create a system message that instructs the model how to parse the document
	systemMessage := "You are a document parser specialized in extracting job information. " +
		"Extract the following details from the document: job title, salary, location, experience required, and employment type. " +
		"Return the information in a structured JSON format."

	// Create a user message with the document content and question
	userMessage := fmt.Sprintf("Document content: %s\n\nQuestion:\n\nExtract the job details in JSON format.",
		documentContent)

	// Create the chat completion request
	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetTemperature(0.0) // Set to 0 for more deterministic responses

	// Create the prompt with system and user messages
	prompt := []genai.Part{
		genai.Text(systemMessage + "\n\n" + userMessage),
	}

	// Call the Gemini AI API
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return JobDetails{}, fmt.Errorf("error calling Gemini AI API: %w", err)
	}

	// Extract the content from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return JobDetails{}, fmt.Errorf("no response from Gemini AI API")
	}

	// Get the text content from the response
	content, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return JobDetails{}, fmt.Errorf("unexpected response format from Gemini AI API")
	}

	// Parse the JSON response into JobDetails struct
	var jobDetails JobDetails
	if err := json.Unmarshal([]byte(string(content)), &jobDetails); err != nil {
		return JobDetails{}, fmt.Errorf("error parsing Gemini AI response: %w", err)
	}

	return jobDetails, nil
}

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if the .env file doesn't exist
		log.Println("No .env file found. Using system environment variables.")
	} else {
		log.Println("Loaded environment variables from .env file.")
	}

	// Get Gemini API key from environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is not set")
	}

	// Initialize the Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer client.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /parse-document", func(w http.ResponseWriter, r *http.Request) {
		// Get the question parameter from the request
		/*question := r.URL.Query().Get("question")
		if question == "" {
			http.Error(w, "Missing 'question' parameter", http.StatusBadRequest)
			return
		}*/

		// Get the file from the request
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the file content
		fileContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Sanitize and convert file content to valid UTF-8 string
		documentContent := sanitizeUTF8(fileContent)

		// Parse the document using Gemini
		ctx := r.Context()
		jobDetails, err := parseDocumentWithGemini(ctx, client, documentContent)
		if err != nil {
			http.Error(w, "Error parsing document: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return a response with the structured job details
		response := map[string]interface{}{
			"status":      "success",
			"message":     "Document parsed successfully",
			"file_name":   header.Filename,
			"file_size":   len(fileContent),
			"job_details": jobDetails,
		}

		// Return the response as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
