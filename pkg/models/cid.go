package models

import (
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