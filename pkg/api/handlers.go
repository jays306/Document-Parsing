package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/generative-ai-go/genai"

	"DocumentParsingSystem/pkg/database"
	"DocumentParsingSystem/pkg/models"
	"DocumentParsingSystem/pkg/parsers"
)

// SetupRoutes configures the HTTP routes for the application
func SetupRoutes(mux *http.ServeMux, client *genai.Client, db *sql.DB) {
	mux.HandleFunc("POST /parse-document", func(w http.ResponseWriter, r *http.Request) {
		handleParseDocument(w, r, client, db)
	})

	mux.HandleFunc("POST /finalize-parsed-fields", func(w http.ResponseWriter, r *http.Request) {
		handleFinalizeParsedFields(w, r, db)
	})
}

// handleParseDocument handles the /parse-document endpoint
func handleParseDocument(w http.ResponseWriter, r *http.Request, client *genai.Client, db *sql.DB) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	case strings.HasSuffix(strings.ToLower(header.Filename), ".csv"):
		mimeType = "text/csv"
	case strings.HasSuffix(strings.ToLower(header.Filename), ".png"):
		mimeType = "image/png"
	case strings.HasSuffix(strings.ToLower(header.Filename), ".txt"):
		mimeType = "text/plain"
	case strings.HasSuffix(strings.ToLower(header.Filename), ".doc"):
		mimeType = "application/msword"
	case strings.HasSuffix(strings.ToLower(header.Filename), ".docx"):
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	// Determine document type from request parameter, default to Form941Type if not specified
	docTypeStr := r.FormValue("document_type")
	var docType models.DocumentType
	switch docTypeStr {
	case "job_details":
		docType = models.JobDetailsType
	case "form_941":
		docType = models.Form941Type
	default:
		// Default to Form941Type for backward compatibility
		docType = models.Form941Type
	}

	// Parse the document using Gemini's multimodal capabilities
	ctx := r.Context()

	// Use the appropriate type parameter based on document type
	var parsedResult interface{}
	var parseErr error

	switch docType {
	case models.JobDetailsType:
		var result models.JobDetails
		result, parseErr = parsers.ParseDocumentWithGeminiMultimodal[models.JobDetails](ctx, client, fileContent, mimeType, docType)
		parsedResult = result
	case models.Form941Type:
		var result models.Form941
		result, parseErr = parsers.ParseDocumentWithGeminiMultimodal[models.Form941](ctx, client, fileContent, mimeType, docType)
		parsedResult = result
	default:
		http.Error(w, "Unsupported document type: "+string(docType), http.StatusBadRequest)
		return
	}

	if parseErr != nil {
		// Fallback to text-based approach if multimodal approach fails
		log.Printf("Multimodal approach failed: %v. Falling back to text-based approach.", parseErr)
		http.Error(w, "Error parsing document: "+parseErr.Error(), http.StatusInternalServerError)
		return
	}

	// Return a response with the structured details
	response := map[string]interface{}{
		"status":        "success",
		"message":       "Document parsed successfully",
		"file_name":     header.Filename,
		"file_size":     len(fileContent),
		"document_type": docType,
		"parsed_result": parsedResult,
	}

	// Return the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleFinalizeParsedFields handles the /finalize-parsed-fields endpoint
func handleFinalizeParsedFields(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if database is available
	if db == nil {
		http.Error(w, "Database connection is not available", http.StatusServiceUnavailable)
		return
	}

	// Parse the request body
	var req models.FinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.DocumentName == "" {
		http.Error(w, "document_name is required", http.StatusBadRequest)
		return
	}

	if len(req.ParsedFields) == 0 {
		http.Error(w, "parsed_fields is required", http.StatusBadRequest)
		return
	}

	// Store the parsed fields in the database
	id, err := database.StoreParsedFields(db, req.ParsedFields, req.DocumentName, req.DocumentType)
	if err != nil {
		http.Error(w, "Error storing parsed fields: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	response := map[string]interface{}{
		"status":  "success",
		"message": "Parsed fields stored successfully",
		"id":      id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
