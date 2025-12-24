package models

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"
)

func TestCIDJSON(t *testing.T) {
	cid := CID{
		Value:   "550e8400-e29b-51d4-a716-446655440000",
		IsValid: true,
		UUID:    "550e8400-e29b-51d4-a716-446655440000",
	}

	data, err := json.Marshal(cid)
	if err != nil {
		t.Fatalf("failed to marshal CID: %v", err)
	}

	var decoded CID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CID: %v", err)
	}

	if decoded.Value != cid.Value {
		t.Errorf("Value = %v, want %v", decoded.Value, cid.Value)
	}
	if decoded.IsValid != cid.IsValid {
		t.Errorf("IsValid = %v, want %v", decoded.IsValid, cid.IsValid)
	}
	if decoded.UUID != cid.UUID {
		t.Errorf("UUID = %v, want %v", decoded.UUID, cid.UUID)
	}
}

func TestCIDRecordJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	record := CIDRecord{
		CID:         "test-cid",
		UUID:        "550e8400-e29b-51d4-a716-446655440000",
		Timestamp:   now,
		RawLogLine:  "test log line",
		IsValid:     true,
		ExtractedAt: now,
		Metadata:    map[string]string{"key": "value"},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("failed to marshal CIDRecord: %v", err)
	}

	var decoded CIDRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CIDRecord: %v", err)
	}

	if decoded.CID != record.CID {
		t.Errorf("CID = %v, want %v", decoded.CID, record.CID)
	}
	if decoded.UUID != record.UUID {
		t.Errorf("UUID = %v, want %v", decoded.UUID, record.UUID)
	}
	if decoded.RawLogLine != record.RawLogLine {
		t.Errorf("RawLogLine = %v, want %v", decoded.RawLogLine, record.RawLogLine)
	}
	if decoded.IsValid != record.IsValid {
		t.Errorf("IsValid = %v, want %v", decoded.IsValid, record.IsValid)
	}
	if decoded.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want %v", decoded.Metadata["key"], "value")
	}
}

func TestLogEntryJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	entry := LogEntry{
		Timestamp: now,
		Level:     "INFO",
		Message:   "test message",
		Source:    "/var/log/app.log",
		Fields:    map[string]string{"field1": "value1"},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal LogEntry: %v", err)
	}

	var decoded LogEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal LogEntry: %v", err)
	}

	if decoded.Level != entry.Level {
		t.Errorf("Level = %v, want %v", decoded.Level, entry.Level)
	}
	if decoded.Message != entry.Message {
		t.Errorf("Message = %v, want %v", decoded.Message, entry.Message)
	}
	if decoded.Source != entry.Source {
		t.Errorf("Source = %v, want %v", decoded.Source, entry.Source)
	}
}

func TestProcessingResultJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	result := ProcessingResult{
		CIDs: []CID{
			{Value: "cid1", IsValid: true},
			{Value: "cid2", IsValid: false},
		},
		Timestamp: now,
		Success:   true,
		Error:     "",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal ProcessingResult: %v", err)
	}

	var decoded ProcessingResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ProcessingResult: %v", err)
	}

	if len(decoded.CIDs) != 2 {
		t.Errorf("len(CIDs) = %v, want %v", len(decoded.CIDs), 2)
	}
	if decoded.Success != result.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, result.Success)
	}
}

func TestLogSourceJSON(t *testing.T) {
	source := LogSource{
		Path:        "/var/log/app",
		Name:        "application",
		Patterns:    []string{"*.log", "*.txt"},
		Active:      true,
		Description: "Application logs",
	}

	data, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("failed to marshal LogSource: %v", err)
	}

	var decoded LogSource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal LogSource: %v", err)
	}

	if decoded.Path != source.Path {
		t.Errorf("Path = %v, want %v", decoded.Path, source.Path)
	}
	if decoded.Name != source.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, source.Name)
	}
	if len(decoded.Patterns) != 2 {
		t.Errorf("len(Patterns) = %v, want %v", len(decoded.Patterns), 2)
	}
	if decoded.Active != source.Active {
		t.Errorf("Active = %v, want %v", decoded.Active, source.Active)
	}
}

func TestCIDPatternJSON(t *testing.T) {
	pattern := CIDPattern{
		Name:        "standard_cid",
		RegexString: `CID\s*[=:]\s*([a-fA-F0-9-]+)`,
		Regex:       regexp.MustCompile(`CID\s*[=:]\s*([a-fA-F0-9-]+)`),
		UUIDGroup:   1,
		Enabled:     true,
	}

	data, err := json.Marshal(pattern)
	if err != nil {
		t.Fatalf("failed to marshal CIDPattern: %v", err)
	}

	var decoded CIDPattern
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CIDPattern: %v", err)
	}

	if decoded.Name != pattern.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, pattern.Name)
	}
	if decoded.RegexString != pattern.RegexString {
		t.Errorf("RegexString = %v, want %v", decoded.RegexString, pattern.RegexString)
	}
	if decoded.UUIDGroup != pattern.UUIDGroup {
		t.Errorf("UUIDGroup = %v, want %v", decoded.UUIDGroup, pattern.UUIDGroup)
	}
	// Regex should be nil after JSON decode since it's tagged with json:"-"
	if decoded.Regex != nil {
		t.Error("Regex should be nil after JSON decode")
	}
}

func TestCIDEntryJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	entry := CIDEntry{
		CID:       "test-cid",
		Timestamp: now,
		LogLine:   "test log line with CID",
		UUIDs: []UUID{
			{Value: "550e8400-e29b-51d4-a716-446655440000", Version: 5, ExtractedAt: now},
		},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal CIDEntry: %v", err)
	}

	var decoded CIDEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CIDEntry: %v", err)
	}

	if decoded.CID != entry.CID {
		t.Errorf("CID = %v, want %v", decoded.CID, entry.CID)
	}
	if len(decoded.UUIDs) != 1 {
		t.Errorf("len(UUIDs) = %v, want %v", len(decoded.UUIDs), 1)
	}
}

func TestUUIDJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	uuid := UUID{
		Value:       "550e8400-e29b-51d4-a716-446655440000",
		Version:     5,
		ExtractedAt: now,
	}

	data, err := json.Marshal(uuid)
	if err != nil {
		t.Fatalf("failed to marshal UUID: %v", err)
	}

	var decoded UUID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal UUID: %v", err)
	}

	if decoded.Value != uuid.Value {
		t.Errorf("Value = %v, want %v", decoded.Value, uuid.Value)
	}
	if decoded.Version != uuid.Version {
		t.Errorf("Version = %v, want %v", decoded.Version, uuid.Version)
	}
}

func TestCorrelatedEntryJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	correlated := CorrelatedEntry{
		CIDEntry: CIDEntry{
			CID:       "test-cid",
			Timestamp: now,
			LogLine:   "test log",
			UUIDs:     nil,
		},
		CorrelationID: "corr-123",
		ProcessedAt:   now,
	}

	data, err := json.Marshal(correlated)
	if err != nil {
		t.Fatalf("failed to marshal CorrelatedEntry: %v", err)
	}

	var decoded CorrelatedEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CorrelatedEntry: %v", err)
	}

	if decoded.CorrelationID != correlated.CorrelationID {
		t.Errorf("CorrelationID = %v, want %v", decoded.CorrelationID, correlated.CorrelationID)
	}
	if decoded.CIDEntry.CID != correlated.CIDEntry.CID {
		t.Errorf("CIDEntry.CID = %v, want %v", decoded.CIDEntry.CID, correlated.CIDEntry.CID)
	}
}

func TestValidationResultJSON(t *testing.T) {
	result := ValidationResult{
		Valid:   true,
		Version: 5,
		Variant: "RFC4122",
		Error:   "",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult: %v", err)
	}

	var decoded ValidationResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult: %v", err)
	}

	if decoded.Valid != result.Valid {
		t.Errorf("Valid = %v, want %v", decoded.Valid, result.Valid)
	}
	if decoded.Version != result.Version {
		t.Errorf("Version = %v, want %v", decoded.Version, result.Version)
	}
	if decoded.Variant != result.Variant {
		t.Errorf("Variant = %v, want %v", decoded.Variant, result.Variant)
	}
}

func TestValidationResult_IsU5UUID(t *testing.T) {
	tests := []struct {
		name    string
		version int
		want    bool
	}{
		{"version 5", 5, true},
		{"version 4", 4, false},
		{"version 1", 1, false},
		{"version 0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidationResult{Version: tt.version}
			if got := v.IsU5UUID(); got != tt.want {
				t.Errorf("IsU5UUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCIDRecordWithEmptyMetadata(t *testing.T) {
	record := CIDRecord{
		CID:      "test-cid",
		Metadata: nil,
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("failed to marshal CIDRecord with nil metadata: %v", err)
	}

	var decoded CIDRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CIDRecord: %v", err)
	}

	if decoded.CID != record.CID {
		t.Errorf("CID = %v, want %v", decoded.CID, record.CID)
	}
}

func TestProcessingResultWithError(t *testing.T) {
	result := ProcessingResult{
		CIDs:      nil,
		Timestamp: time.Now(),
		Success:   false,
		Error:     "processing failed",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal ProcessingResult: %v", err)
	}

	var decoded ProcessingResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ProcessingResult: %v", err)
	}

	if decoded.Success != false {
		t.Error("Success should be false")
	}
	if decoded.Error != "processing failed" {
		t.Errorf("Error = %v, want %v", decoded.Error, "processing failed")
	}
}
