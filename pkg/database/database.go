package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectDB establishes a connection to the PostgreSQL database
func ConnectDB() (*sql.DB, error) {
	// Get database connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is not set")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "document_parsing"
	}

	// Create the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

// InitDB initializes the database by creating the necessary tables if they don't exist
func InitDB(db *sql.DB) error {
	// Create the parsed_fields table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS parsed_fields (
		id SERIAL PRIMARY KEY,
		parsed_fields JSONB NOT NULL,
		document_name VARCHAR NOT NULL,
		document_type VARCHAR NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating parsed_fields table: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// StoreParsedFields stores the parsed fields in the database
func StoreParsedFields(db *sql.DB, parsedFields []byte, documentName, documentType string) (int, error) {
	query := `
	INSERT INTO parsed_fields (parsed_fields, document_name, document_type, created_at, updated_at)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	RETURNING id`

	var id int
	err := db.QueryRow(query, parsedFields, documentName, documentType).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error storing parsed fields: %w", err)
	}

	return id, nil
}
