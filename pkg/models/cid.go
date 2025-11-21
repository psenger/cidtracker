package models

import "time"

// CID represents a correlation identifier found in logs
type CID struct {
	ID        string    `json:"id"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// CIDEntry represents a single CID extraction from a log line
type CIDEntry struct {
	CID       string    `json:"cid"`
	Timestamp time.Time `json:"timestamp"`
	LogLine   string    `json:"log_line"`
	UUIDs     []UUID    `json:"uuids"`
}

// UUID represents an extracted UUID from CID content
type UUID struct {
	Value       string    `json:"value"`
	Version     int       `json:"version"`
	ExtractedAt time.Time `json:"extracted_at"`
}

// CorrelatedEntry represents a processed CID entry with correlation metadata
type CorrelatedEntry struct {
	CIDEntry      CIDEntry  `json:"cid_entry"`
	CorrelationID string    `json:"correlation_id"`
	ProcessedAt   time.Time `json:"processed_at"`
}