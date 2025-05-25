# Document Parsing

A Go application that provides an HTTP endpoint for parsing documents using Google's Gemini API.

## Prerequisites

This section covers the prerequisites needed to run the application, including environment variables and database setup.

### PostgreSQL Installation

This application requires PostgreSQL for storing parsed document data. Follow these instructions to install PostgreSQL on your system:

#### macOS

##### Using Homebrew (Recommended)

1. Install Homebrew if you haven't already:
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

2. Install PostgreSQL:
   ```bash
   brew install postgresql
   ```

3. Start the PostgreSQL service:
   ```bash
   brew services start postgresql
   ```

##### Using the Installer

1. Download the PostgreSQL installer from the [official website](https://www.postgresql.org/download/macosx/).
2. Run the installer and follow the installation wizard.
3. Complete the installation.

#### Creating a Database and User

After installing PostgreSQL, you need to create a database and user for this application:

1. Switch to the PostgreSQL user:
   ```bash
   sudo -i -u postgres
   ```

2. Access the PostgreSQL command-line interface:
   ```bash
   psql
   ```

3. Create a new user (replace `your_username` and `your_password` with your desired values):
   ```sql
   CREATE USER your_username WITH PASSWORD 'your_password';
   ```

4. Create a new database (default name is `document_parsing`):
   ```sql
   CREATE DATABASE document_parsing;
   ```

5. Grant privileges to the user on the database:
   ```sql
   GRANT ALL PRIVILEGES ON DATABASE document_parsing TO your_username;
   ```

6. Exit the PostgreSQL command-line interface:
   ```sql
   \q
   ```

7. Exit the postgres user shell:
   ```bash
   exit
   ```

8. Update your environment variables or .env file with the database connection details:
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_username
   DB_PASSWORD=your_password
   DB_NAME=document_parsing
   ```

### Environment Variables

This application requires the following environment variables to be set:

#### Required Environment Variables

- `GEMINI_API_KEY`: Your Google Gemini API key for authentication with the Gemini API.
  - If this variable is not set, the application will exit with a fatal error.
  - `DB_PASSWORD`: The password for the PostgreSQL database.
    - If this variable is not set, the database connection will fail and the `/finalize-parsed-fields` endpoint will not be available.

#### Optional Environment Variables

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

### Document Types

The application supports parsing different types of documents. You can specify the document type using the `document_type` parameter in your request:

- `form_941`: For parsing IRS Form 941 documents (default)
- `job_details`: For parsing job-related documents

If no document type is specified, the application defaults to `form_941` for backward compatibility.

### Example curl Command

Here's an example of how to use the `/parse-document` endpoint with curl:

```bash
curl -X GET "http://localhost:8080/parse-document" \
  -F "file=@/path/to/your/document.pdf" \
  -F "document_type=form_941" \
  -H "Content-Type: multipart/form-data"
```

#### Explanation:

- `-X GET`: Specifies that we're making a GET request
- `http://localhost:8080/parse-document`: The endpoint URL
- `-F "file=@/path/to/your/document.pdf"`: Uploads a file using multipart/form-data
  - Replace `/path/to/your/document.pdf` with the actual path to your PDF file
- `-F "document_type=form_941"`: Specifies the document type to parse
  - Options: `form_941` (default) or `job_details`
- `-H "Content-Type: multipart/form-data"`: Sets the appropriate content type header

#### Example Response for Form 941:

```json
{
  "status": "success",
  "message": "Document parsed successfully",
  "file_name": "f941.pdf",
  "file_size": 1234567,
  "document_type": "form_941",
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

#### Example Response for Job Details:

```json
{
  "status": "success",
  "message": "Document parsed successfully",
  "file_name": "job_posting.pdf",
  "file_size": 987654,
  "document_type": "job_details",
  "parsed_result": {
    "title": "Senior Software Engineer",
    "salary": "$120,000 - $150,000 per year",
    "location": "San Francisco, CA (Remote Available)",
    "experience": "5+ years of experience in software development",
    "employment-type": "Full-time"
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

3. **Add the file upload and document type**:
   - Go to the "Body" tab
   - Select "form-data"
   - Add a new key called "file"
   - Click on the dropdown next to the key and select "File"
   - Click "Select Files" and choose your PDF document
   - Add another key called "document_type" (as Text)
   - Enter either "form_941" or "job_details" as the value

4. **Send the request**:
   - Click the "Send" button
   - You should receive a JSON response similar to the example shown in the curl section

#### Tips for Postman:
- Postman automatically sets the correct `Content-Type` header for multipart/form-data
- You can save this request to a collection for future use
- To test with different files, simply select a different file in the form-data section
- If you don't specify a document_type, the system will default to "form_941"

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
    document_name VARCHAR NOT NULL,
    document_type VARCHAR NOT NULL,
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
  "document_name": "example.pdf",
  "document_type": "form_941"
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
    "document_name": "f941.pdf",
    "document_type": "form_941"
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