package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// JobDetails represents the structured response format for parsed job information
type JobDetails struct {
	Title          string `json:"title"`
	Salary         string `json:"salary"`
	Location       string `json:"location"`
	Experience     string `json:"experience"`
	EmploymentType string `json:"employment-type"`
}

type Form941 struct {
	EIN       string `json:"EIN"`
	Name      string `json:"Name"`
	TradeName string `json:"Trade Name"`
	Address   string `json:"Address"`
	Box1      string `json:"Box 1"`
	Box2      string `json:"Box 2"`
	Box3      string `json:"Box 3"`
	Box4      bool   `json:"Box 4"`
	Box5e     string `json:"Box 5e"`
	Box5f     string `json:"Box 5f"`
	Box6      string `json:"Box 6"`
	Box7      string `json:"Box 7"`
	Box8      string `json:"Box 8"`
	Box9      string `json:"Box 9"`
	Box10     string `json:"Box 10"`
	Box11     string `json:"Box 11"`
	Box12     string `json:"Box 12"`
	Box13     string `json:"Box 13"`
	Box14     string `json:"Box 14"`
}

func jobDetailsPrompt() string {
	return `You are a document parser specialized in extracting job information.
Extract the following details from the document: job title, salary, location, experience required, and employment type.

Return ONLY a valid JSON object with the following structure:
{
  "title": "Job Title",
  "salary": "Salary Information",
  "location": "Job Location",
  "experience": "Required Experience",
  "employment-type": "Type of Employment (Full-time, Part-time, etc.)"
}

Do not include any explanations, markdown formatting, or additional text outside the JSON object.
If you cannot find a specific field, use an empty string for that field.`
}

func form941Prompt() string {
	return `You are a document parser specialized in extracting job information.
Extract the following details from the document: job title, salary, location, experience required, and employment type.

Return ONLY a valid JSON object with the following structure:
{
	"EIN": "12-3456789",
	"Name": "Company Name",
	"Trade name": "Trade name",
	"Address": "Full address",
	"Box 1": "$11.11",
	"Box 2": "$22.22",
	"Box 3": "$33.33",
	"Box 4": True or False,
	"Box 5e": "$55.55",
	"Box 5f": "$55.55",
	"Box 6": "$66.66",
	"Box 7": "$77.77",
	"Box 8": "$88.88",
	"Box 9": "$99.99"
	"Box 10": "$100.00",
	"Box 11": "$111.11",
	"Box 12": "$121.21"
	"Box 13": "$121.21"
	"Box 14": "$121.21"
}

Do not include any explanations, markdown formatting, or additional text outside the JSON object.
If you cannot find a specific field, use an empty string for that field.`
}

// cleanJSONResponse removes markdown code block markers from a JSON string
// This handles cases where the Gemini API returns JSON wrapped in ```json ... ``` markers
func cleanJSONResponse(jsonStr string) string {
	// Remove leading ```json or ``` if present
	jsonStr = strings.TrimPrefix(strings.TrimSpace(jsonStr), "```json")
	jsonStr = strings.TrimPrefix(strings.TrimSpace(jsonStr), "```")

	// Remove trailing ``` if present
	jsonStr = strings.TrimSuffix(strings.TrimSpace(jsonStr), "```")

	// Trim any remaining whitespace
	return strings.TrimSpace(jsonStr)
}

// parseDocumentWithGeminiMultimodal uses the Gemini AI API to extract job details from a document
// by sending the file directly as binary data instead of as text
func parseDocumentWithGeminiMultimodal(ctx context.Context, client *genai.Client, fileContent []byte, mimeType string) (Form941, error) {

	// Create a detailed schema-based instruction for the model
	schemaInstruction := form941Prompt()

	// Create the chat completion request
	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.0) // Set to 0 for more deterministic responses

	// Create the prompt with schema instructions and the file as binary data
	prompt := []genai.Part{
		genai.Text(schemaInstruction),
		// Send the file directly as binary data with its MIME type
		genai.Blob{
			MIMEType: mimeType,
			Data:     fileContent,
		},
	}

	// Call the Gemini AI API
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return Form941{}, fmt.Errorf("error calling Gemini AI API: %w", err)
	}

	// Extract the content from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return Form941{}, fmt.Errorf("no response from Gemini AI API")
	}

	// Get the text content from the response
	content, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return Form941{}, fmt.Errorf("unexpected response format from Gemini AI API")
	}

	// Clean the response to ensure it's valid JSON
	jsonStr := string(content)

	// Remove markdown code block markers if present
	jsonStr = cleanJSONResponse(jsonStr)

	// Log the cleaned JSON for debugging
	log.Printf("Cleaned JSON response: %s", jsonStr)

	// Parse the JSON response into JobDetails struct
	//var jobDetails JobDetails
	var form941 Form941
	if err := json.Unmarshal([]byte(jsonStr), &form941); err != nil {
		log.Printf("JSON parsing failed: %v. Falling back to text-based approach.", err)
		return Form941{}, fmt.Errorf("error parsing Gemini AI response: %w\nResponse: %s", err, jsonStr)
	}

	return form941, nil
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

		// Determine the MIME type based on the file extension
		mimeType := "application/octet-stream" // Default MIME type
		switch {
		case strings.HasSuffix(strings.ToLower(header.Filename), ".pdf"):
			mimeType = "application/pdf"
		case strings.HasSuffix(strings.ToLower(header.Filename), ".txt"):
			mimeType = "text/plain"
		case strings.HasSuffix(strings.ToLower(header.Filename), ".doc"):
			mimeType = "application/msword"
		case strings.HasSuffix(strings.ToLower(header.Filename), ".docx"):
			mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		}

		// Parse the document using Gemini's multimodal capabilities
		ctx := r.Context()
		parsedResult, err := parseDocumentWithGeminiMultimodal(ctx, client, fileContent, mimeType)
		if err != nil {
			// Fallback to text-based approach if multimodal approach fails
			log.Printf("Multimodal approach failed: %v. Falling back to text-based approach.", err)
			if err != nil {
				http.Error(w, "Error parsing document: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Return a response with the structured job details
		response := map[string]interface{}{
			"status":        "success",
			"message":       "Document parsed successfully",
			"file_name":     header.Filename,
			"file_size":     len(fileContent),
			"parsed_result": parsedResult,
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
