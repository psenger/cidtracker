package models

import (
	"regexp"
	"time"
)

// CID represents a correlation identifier with validation status
type CID struct {
	Value   string `json:"value"`
	IsValid bool   `json:"is_valid"`
	UUID    string `json:"uuid,omitempty"`
}

// CIDRecord represents a processed log entry with extracted CID information
type CIDRecord struct {
	CID         string            `json:"cid"`
	UUID        string            `json:"uuid,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	RawLogLine  string            `json:"raw_log_line"`
	IsValid     bool              `json:"is_valid"`
	ExtractedAt time.Time         `json:"extracted_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// LogEntry represents a structured log entry for processing
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level,omitempty"`
	Message   string            `json:"message"`
	Source    string            `json:"source,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// ProcessingResult contains the results of CID extraction and validation
type ProcessingResult struct {
	CIDs      []CID     `json:"cids"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// LogSource represents a log file source configuration
type LogSource struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Patterns    []string `json:"patterns"`
	Active      bool     `json:"active"`
	Description string   `json:"description"`
}

// CIDPattern represents a pattern for extracting CIDs
type CIDPattern struct {
	Name        string         `json:"name"`
	RegexString string         `json:"regex_string"`
	Regex       *regexp.Regexp `json:"-"`
	UUIDGroup   int            `json:"uuid_group"`
	Enabled     bool           `json:"enabled"`
}

// CIDEntry represents an extracted CID entry with associated UUIDs
type CIDEntry struct {
	CID       string    `json:"cid"`
	Timestamp time.Time `json:"timestamp"`
	LogLine   string    `json:"log_line"`
	UUIDs     []UUID    `json:"uuids"`
}

// UUID represents an extracted UUID with metadata
type UUID struct {
	Value       string    `json:"value"`
	Version     int       `json:"version"`
	ExtractedAt time.Time `json:"extracted_at"`
}

// CorrelatedEntry represents a CID entry with correlation information
type CorrelatedEntry struct {
	CIDEntry      CIDEntry  `json:"cid_entry"`
	CorrelationID string    `json:"correlation_id"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// ValidationResult contains the result of UUID validation
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Version int    `json:"version"`
	Variant string `json:"variant"`
	Error   string `json:"error,omitempty"`
}

// IsU5UUID returns true if the UUID is version 5
func (v ValidationResult) IsU5UUID() bool {
	return v.Version == 5
}