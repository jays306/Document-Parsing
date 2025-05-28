package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"log"
	"strings"

	"DocumentParsingSystem/pkg/models"
)

// jobDetailsPrompt returns the prompt for extracting job details
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

// form941Prompt returns the prompt for extracting Form 941 details
func form941Prompt() string {
	return `You are a document parser specialized in extracting tax-related information.
Extract the following details from the document based on Form 941: EIN, name, trade name, address, and boxes 1–15.
Note that EIN values are consistently formatted as separate digits that in a separate box, when combined, form a 9-digit number.
All box fields except for Box 1 and Box 4 should follow this format: $11.11 — consisting of a dollar sign, one or more digits, a decimal point, and two digits.

Return only a valid JSON object with the following structure:
{
	"EIN": "123456789",
	"Name": "Company Name",
	"Trade name": "Trade name",
	"Address": "Full address",
	"Box 1": "111",
	"Box 2": "$22.22",
	"Box 3": "$33.33",
	"Box 4": true or false,
	"Box 5e": "$55.55",
	"Box 5f": "$55.55",
	"Box 6": "$66.66",
	"Box 7": "$77.77",
	"Box 8": "$88.88",
	"Box 9": "$99.99",
	"Box 10": "$100.00",
	"Box 11": "$111.11",
	"Box 12": "$121.21",
	"Box 13": "$121.21",
	"Box 14": "$121.21",
	"Box 15": "$121.21"
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

// ParseDocumentWithGeminiMultimodal uses the Gemini AI API to extract structured data from a document
// by sending the file directly as binary data instead of as text
func ParseDocumentWithGeminiMultimodal[T models.JobDetails | models.Form941](ctx context.Context, client *genai.Client, fileContent []byte, mimeType string, docType models.DocumentType) (T, error) {
	// Determine which prompt to use based on document type
	var schemaInstruction string
	switch docType {
	case models.JobDetailsType:
		schemaInstruction = jobDetailsPrompt()
	case models.Form941Type:
		schemaInstruction = form941Prompt()
	default:
		var zero T
		return zero, fmt.Errorf("unsupported document type: %s", docType)
	}

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
		var zero T
		return zero, fmt.Errorf("error calling Gemini AI API: %w", err)
	}

	// Extract the content from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		var zero T
		return zero, fmt.Errorf("no response from Gemini AI API")
	}

	// Get the text content from the response
	content, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		var zero T
		return zero, fmt.Errorf("unexpected response format from Gemini AI API")
	}

	// Clean the response to ensure it's valid JSON
	jsonStr := string(content)

	// Remove markdown code block markers if present
	jsonStr = cleanJSONResponse(jsonStr)

	// Log the cleaned JSON for debugging
	log.Printf("Cleaned JSON response: %s", jsonStr)

	// Parse the JSON response into the appropriate struct based on document type
	var result T
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("JSON parsing failed for %T: %v", result, err)
		var zero T
		return zero, fmt.Errorf("error parsing Gemini AI response for %T: %w\nResponse: %s", result, err, jsonStr)
	}

	return result, nil
}
