package models

import (
	"encoding/json"
	"time"
)

// JobDetails represents the structured response format for parsed job information
type JobDetails struct {
	Title          string `json:"title"`
	Salary         string `json:"salary"`
	Location       string `json:"location"`
	Experience     string `json:"experience"`
	EmploymentType string `json:"employment-type"`
}

// Form941 represents the structured response format for parsed Form 941 information
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

// ParsedFields represents the data to be stored in the database
type ParsedFields struct {
	ID           int             `json:"id"`
	ParsedFields json.RawMessage `json:"parsed_fields"`
	DocumentName string          `json:"document_name"`
	DocumentType string          `json:"document_type"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// FinalizeRequest represents the request body for the finalize-parsed-fields endpoint
type FinalizeRequest struct {
	ParsedFields json.RawMessage `json:"parsed_fields"`
	DocumentName string          `json:"document_name"`
	DocumentType string          `json:"document_type"`
}

// DocumentType represents the type of document to parse
type DocumentType string

const (
	JobDetailsType DocumentType = "job_details"
	Form941Type    DocumentType = "form_941"
)