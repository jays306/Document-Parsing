# Document Parsing

A Go application that provides an HTTP endpoint for parsing documents using Google's Gemini API.

## Environment Variables

This application requires the following environment variables to be set:

### Required Environment Variables

- `GEMINI_API_KEY`: Your Google Gemini API key for authentication with the Gemini API.
  - If this variable is not set, the application will exit with a fatal error.
- `DB_PASSWORD`: The password for the PostgreSQL database.
  - If this variable is not set, the database connection will fail and the `/finalize-parsed-fields` endpoint will not be available.

### Optional Environment Variables

- `PORT`: The port number on which the server will listen.
  - Default value: `8080`
  - If not specified, the server will listen on port 8080.
- `DB_HOST`: The hostname of the PostgreSQL database.
  - Default value: `localhost`
- `DB_PORT`: The port number of the PostgreSQL database.
  - Default value: `5432`
- `DB_USER`: The username for the PostgreSQL database.
  - Default value: `postgres`
- `DB_NAME`: The name of the PostgreSQL database.
  - Default value: `document_parsing`

## Setting Environment Variables

### Linux/macOS

You can set environment variables in your terminal session:

```bash
# Set the Gemini API key
export GEMINI_API_KEY="your-api-key-here"

# Set the port (optional)
export PORT="3000"

# Run the application
go run main.go
```

Alternatively, you can set them in a single line when running the application:

```bash
GEMINI_API_KEY="your-api-key-here" PORT="3000" go run main.go
```

To make these environment variables persistent across terminal sessions, add them to your shell profile file (e.g., `~/.bash_profile`, `~/.bashrc`, or `~/.zshrc`):

```bash
echo 'export GEMINI_API_KEY="your-api-key-here"' >> ~/.bashrc
echo 'export PORT="3000"' >> ~/.bashrc
source ~/.bashrc
```

## Development

For development, you can use the methods described above to set environment variables locally.

### Using a .env File (Recommended for Development)

This application supports loading environment variables from a `.env` file using the [godotenv](https://github.com/joho/godotenv) package. This is the recommended approach for development environments.

To use this feature:

1. Copy the provided `.env.example` file to create your own `.env` file:

```bash
cp .env .env
```

2. Edit the `.env` file to set your environment variables:

```
# Required environment variables
GEMINI_API_KEY=your-api-key-here
DB_PASSWORD=your-database-password

# Optional environment variables
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=document_parsing
```

3. Run the application normally:

```bash
go run main.go
```

The application will automatically load the environment variables from the `.env` file if it exists.


## API Usage

Once the server is running, you can access the document parsing endpoint at:

```
GET /parse-document
```

Include a file in the request using multipart/form-data with the key "file". The application will parse the document and extract structured information using Google's Gemini API.

### Example curl Command

Here's an example of how to use the `/parse-document` endpoint with curl:

```bash
curl -X GET "http://localhost:8080/parse-document" \
  -F "file=@/path/to/your/document.pdf" \
  -H "Content-Type: multipart/form-data"
```

#### Explanation:

- `-X GET`: Specifies that we're making a GET request
- `http://localhost:8080/parse-document`: The endpoint URL
- `-F "file=@/path/to/your/document.pdf"`: Uploads a file using multipart/form-data
  - Replace `/path/to/your/document.pdf` with the actual path to your PDF file
- `-H "Content-Type: multipart/form-data"`: Sets the appropriate content type header

#### Example Response:

```json
{
  "status": "success",
  "message": "Document parsed successfully",
  "file_name": "f941.pdf",
  "file_size": 1234567,
  "parsed_result": {
    "EIN": "12-3456789",
    "Name": "Company Name",
    "Trade Name": "Trade name",
    "Address": "Full address",
    "Box 1": "$11.11",
    "Box 2": "$22.22",
    "Box 3": "$33.33",
    "Box 4": true,
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
    "Box 14": "$121.21"
  }
}
```

Note: The response includes a structured `parsed_result` object with fields extracted from the document. These values are extracted from the document using Google's Gemini API, which analyzes the document content to identify and extract the relevant information.

### Example Postman Request

Here's how to set up a request in Postman to use the `/parse-document` endpoint:

1. **Create a new request**:
   - Open Postman and click on "New" to create a new request
   - Select "GET" as the request method

2. **Set the request URL**:
   - Enter `http://localhost:8080/parse-document`

3. **Add the file upload**:
   - Go to the "Body" tab
   - Select "form-data"
   - Add a new key called "file"
   - Click on the dropdown next to the key and select "File"
   - Click "Select Files" and choose your PDF document

4. **Send the request**:
   - Click the "Send" button
   - You should receive a JSON response similar to the example shown in the curl section

#### Tips for Postman:
- Postman automatically sets the correct `Content-Type` header for multipart/form-data
- You can save this request to a collection for future use
- To test with different files, simply select a different file in the form-data section

## JSON Response Handling

This application includes special handling for JSON responses from the Gemini API. Sometimes, the Gemini API returns JSON wrapped in markdown code block markers (e.g., ```json ... ```), which can cause JSON parsing errors.

### Markdown Code Block Handling

The application includes a `cleanJSONResponse` function that:

1. Removes leading markdown code block markers (```json or ```)
2. Removes trailing markdown code block markers (```)
3. Trims any remaining whitespace

This ensures that even if the Gemini API returns JSON wrapped in markdown code blocks, the application can still parse it correctly.

## CORS Support

This application includes Cross-Origin Resource Sharing (CORS) support, allowing it to be accessed from web applications hosted on different domains. The following CORS headers are set on all responses:

- `Access-Control-Allow-Origin: *` - Allows requests from any origin
- `Access-Control-Allow-Methods: POST, OPTIONS` - Allows POST requests and preflight OPTIONS requests
- `Access-Control-Allow-Headers: Content-Type` - Allows the Content-Type header in requests

This configuration enables the API to be used by web applications regardless of where they are hosted.

## Database Integration

This application includes integration with PostgreSQL for storing parsed document data. The application creates a table called `parsed_fields` with the following schema:

```sql
CREATE TABLE IF NOT EXISTS parsed_fields (
    id SERIAL PRIMARY KEY,
    parsed_fields JSONB NOT NULL,
    document_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Finalize Parsed Fields Endpoint

Once you have parsed a document using the `/parse-document` endpoint, you can store the parsed data in the database using the `/finalize-parsed-fields` endpoint:

```
POST /finalize-parsed-fields
```

#### Request Body

The request body should be a JSON object with the following structure:

```json
{
  "parsed_fields": {
    "field1": "value1",
    "field2": "value2",
    "nested": {
      "field3": "value3"
    }
  },
  "document_name": "example.pdf"
}
```

The `parsed_fields` property can contain any valid JSON object representing the parsed data from the document.

#### Example curl Command

Here's an example of how to use the `/finalize-parsed-fields` endpoint with curl:

```bash
curl -X POST "http://localhost:8080/finalize-parsed-fields" \
  -H "Content-Type: application/json" \
  -d '{
    "parsed_fields": {
      "EIN": "12-3456789",
      "Name": "Company Name",
      "Trade Name": "Trade name",
      "Address": "Full address",
      "Box 1": "$11.11"
    },
    "document_name": "f941.pdf"
  }'
```

#### Example Response

```json
{
  "status": "success",
  "message": "Parsed fields stored successfully",
  "id": 1
}
```

The response includes the ID of the newly created record in the database.

### Database Setup

Before using the `/finalize-parsed-fields` endpoint, you need to set up a PostgreSQL database and configure the application to connect to it using the environment variables described in the "Environment Variables" section.

If the database connection fails, the application will log a warning and the `/finalize-parsed-fields` endpoint will not be available, but the rest of the application will continue to function.
