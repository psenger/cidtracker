package models

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// CIDEntry represents a parsed CID log entry
type CIDEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	CID         string    `json:"cid"`
	UUID        uuid.UUID `json:"uuid"`
	RawLine     string    `json:"raw_line"`
	Source      string    `json:"source"`
	Correlation string    `json:"correlation_id,omitempty"`
}

// CIDPattern represents different CID extraction patterns
type CIDPattern struct {
	Name        string         `json:"name"`
	Regex       *regexp.Regexp `json:"-"`
	RegexString string         `json:"regex"`
	UUIDGroup   int            `json:"uuid_group"`
	Enabled     bool           `json:"enabled"`
}

// LogSource represents a monitored log source
type LogSource struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Patterns    []string `json:"patterns"`
	Active      bool     `json:"active"`
	LastRead    int64    `json:"last_read"`
	Description string   `json:"description,omitempty"`
}

// ValidationResult holds UUID validation results
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Version int    `json:"version"`
	Variant string `json:"variant"`
	Error   string `json:"error,omitempty"`
}

// IsU5UUID validates if the UUID is version 5
func (v ValidationResult) IsU5UUID() bool {
	return v.Valid && v.Version == 5
}
