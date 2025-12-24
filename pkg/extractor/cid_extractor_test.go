package extractor

import (
	"testing"

	"cidtracker/pkg/models"
)

func TestNewCIDExtractor(t *testing.T) {
	e := NewCIDExtractor()
	if e == nil {
		t.Fatal("expected non-nil extractor")
	}
	if e.cidPattern == nil {
		t.Error("cidPattern should not be nil")
	}
	if e.uuidValidator == nil {
		t.Error("uuidValidator should not be nil")
	}
}

func TestExtractCIDs(t *testing.T) {
	tests := []struct {
		name      string
		logLine   string
		wantCount int
		wantCIDs  []string
	}{
		{
			name:      "single CID",
			logLine:   "INFO CID[550e8400-e29b-51d4-a716-446655440000] processing request",
			wantCount: 1,
			wantCIDs:  []string{"550e8400-e29b-51d4-a716-446655440000"},
		},
		{
			name:      "multiple CIDs",
			logLine:   "INFO CID[abc-123] CID[def-456] processing",
			wantCount: 2,
			wantCIDs:  []string{"abc-123", "def-456"},
		},
		{
			name:      "no CID",
			logLine:   "INFO processing request without correlation",
			wantCount: 0,
			wantCIDs:  nil,
		},
		{
			name:      "empty line",
			logLine:   "",
			wantCount: 0,
			wantCIDs:  nil,
		},
		{
			name:      "CID with complex value",
			logLine:   "CID[user:123:session:abc-def-ghi]",
			wantCount: 1,
			wantCIDs:  []string{"user:123:session:abc-def-ghi"},
		},
		{
			name:      "CID at start of line",
			logLine:   "CID[first-cid] some log message",
			wantCount: 1,
			wantCIDs:  []string{"first-cid"},
		},
		{
			name:      "CID at end of line",
			logLine:   "some log message CID[last-cid]",
			wantCount: 1,
			wantCIDs:  []string{"last-cid"},
		},
	}

	e := NewCIDExtractor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := e.ExtractCIDs(tt.logLine)

			if len(entries) != tt.wantCount {
				t.Errorf("ExtractCIDs() returned %d entries, want %d", len(entries), tt.wantCount)
			}

			for i, entry := range entries {
				if i < len(tt.wantCIDs) && entry.CID != tt.wantCIDs[i] {
					t.Errorf("entry[%d].CID = %v, want %v", i, entry.CID, tt.wantCIDs[i])
				}
				if entry.LogLine != tt.logLine {
					t.Errorf("entry[%d].LogLine = %v, want %v", i, entry.LogLine, tt.logLine)
				}
				if entry.Timestamp.IsZero() {
					t.Error("Timestamp should not be zero")
				}
			}
		})
	}
}

func TestExtractCIDs_WithU5UUID(t *testing.T) {
	e := NewCIDExtractor()

	// UUID v5 pattern: version digit is 5, variant is 8, 9, a, or b
	logLine := "CID[550e8400-e29b-51d4-a716-446655440000] request"
	entries := e.ExtractCIDs(logLine)

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.UUIDs) != 1 {
		t.Errorf("expected 1 UUID, got %d", len(entry.UUIDs))
	}

	if len(entry.UUIDs) > 0 {
		uuid := entry.UUIDs[0]
		if uuid.Version != 5 {
			t.Errorf("UUID version = %d, want 5", uuid.Version)
		}
		if uuid.Value != "550e8400-e29b-51d4-a716-446655440000" {
			t.Errorf("UUID value = %s, want 550e8400-e29b-51d4-a716-446655440000", uuid.Value)
		}
	}
}

func TestExtractCIDs_WithNonU5UUID(t *testing.T) {
	e := NewCIDExtractor()

	// UUID v4 - should not be extracted as valid U5
	logLine := "CID[550e8400-e29b-41d4-a716-446655440000] request"
	entries := e.ExtractCIDs(logLine)

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	// Should have no valid UUIDs since it's not v5
	if len(entry.UUIDs) != 0 {
		t.Errorf("expected 0 UUIDs for v4 UUID, got %d", len(entry.UUIDs))
	}
}

func TestExtractUUIDs(t *testing.T) {
	e := NewCIDExtractor()

	tests := []struct {
		name      string
		cidValue  string
		wantCount int
	}{
		{
			name:      "valid U5 UUID",
			cidValue:  "550e8400-e29b-51d4-a716-446655440000",
			wantCount: 1,
		},
		{
			name:      "valid U4 UUID - not extracted",
			cidValue:  "550e8400-e29b-41d4-a716-446655440000",
			wantCount: 0,
		},
		{
			name:      "non-UUID string",
			cidValue:  "simple-cid-value",
			wantCount: 0,
		},
		{
			name:      "multiple UUIDs",
			cidValue:  "550e8400-e29b-51d4-a716-446655440000:550e8400-e29b-51d4-b716-446655440001",
			wantCount: 2,
		},
		{
			name:      "mixed UUID versions",
			cidValue:  "550e8400-e29b-51d4-a716-446655440000:550e8400-e29b-41d4-a716-446655440001",
			wantCount: 1, // Only v5 should be extracted
		},
		{
			name:      "empty string",
			cidValue:  "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuids := e.extractUUIDs(tt.cidValue)
			if len(uuids) != tt.wantCount {
				t.Errorf("extractUUIDs() returned %d UUIDs, want %d", len(uuids), tt.wantCount)
			}

			for _, uuid := range uuids {
				if uuid.Version != 5 {
					t.Errorf("UUID version = %d, want 5", uuid.Version)
				}
				if uuid.ExtractedAt.IsZero() {
					t.Error("ExtractedAt should not be zero")
				}
			}
		})
	}
}

func TestCorrelateEntries(t *testing.T) {
	e := NewCIDExtractor()

	logLines := []string{
		"CID[cid-001] first message",
		"CID[cid-002] second message",
		"CID[cid-003] third message",
	}

	var entries []models.CIDEntry
	for _, line := range logLines {
		entries = append(entries, e.ExtractCIDs(line)...)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	correlated := e.CorrelateEntries(entries)

	if len(correlated) != 3 {
		t.Fatalf("expected 3 correlated entries, got %d", len(correlated))
	}

	for i, corr := range correlated {
		if corr.CorrelationID == "" {
			t.Errorf("correlated[%d].CorrelationID should not be empty", i)
		}
		if corr.ProcessedAt.IsZero() {
			t.Errorf("correlated[%d].ProcessedAt should not be zero", i)
		}
		if corr.CIDEntry.CID != entries[i].CID {
			t.Errorf("correlated[%d].CIDEntry.CID = %v, want %v", i, corr.CIDEntry.CID, entries[i].CID)
		}
	}
}

func TestGenerateCorrelationID(t *testing.T) {
	e := NewCIDExtractor()

	entries := e.ExtractCIDs("CID[test-cid] message")
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}

	correlationID := e.generateCorrelationID(entries[0])

	if correlationID == "" {
		t.Error("correlationID should not be empty")
	}

	// Check that it contains the CID
	if len(correlationID) < len("test-cid") {
		t.Error("correlationID should contain the CID")
	}
}

func TestExtractCIDs_SpecialCharacters(t *testing.T) {
	e := NewCIDExtractor()

	tests := []struct {
		name    string
		logLine string
		wantCID string
	}{
		{
			name:    "CID with dashes",
			logLine: "CID[abc-def-ghi] message",
			wantCID: "abc-def-ghi",
		},
		{
			name:    "CID with underscores",
			logLine: "CID[abc_def_ghi] message",
			wantCID: "abc_def_ghi",
		},
		{
			name:    "CID with colons",
			logLine: "CID[service:instance:request] message",
			wantCID: "service:instance:request",
		},
		{
			name:    "CID with numbers",
			logLine: "CID[request12345] message",
			wantCID: "request12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := e.ExtractCIDs(tt.logLine)
			if len(entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(entries))
			}
			if entries[0].CID != tt.wantCID {
				t.Errorf("CID = %v, want %v", entries[0].CID, tt.wantCID)
			}
		})
	}
}

var _ = models.CIDEntry{}
