package processor

import (
	"context"
	"testing"
	"time"

	"cidtracker/pkg/models"
)

func TestNewProcessor(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	if p == nil {
		t.Fatal("NewProcessor() returned nil")
	}

	if p.extractor == nil {
		t.Error("extractor should not be nil")
	}

	if p.metrics == nil {
		t.Error("metrics should not be nil")
	}

	if p.outputCh == nil {
		t.Error("outputCh should not be nil")
	}
}

func TestProcessor_ProcessLogLine_WithCID(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	logLine := "CID[test-cid-123] processing request"
	err := p.ProcessLogLine(logLine)
	if err != nil {
		t.Fatalf("ProcessLogLine() error = %v", err)
	}

	// Should receive a record
	select {
	case record := <-outputCh:
		if record.CID != "test-cid-123" {
			t.Errorf("CID = %v, want test-cid-123", record.CID)
		}
		if record.RawLogLine != logLine {
			t.Errorf("RawLogLine = %v, want %v", record.RawLogLine, logLine)
		}
		if record.ExtractedAt.IsZero() {
			t.Error("ExtractedAt should not be zero")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for record")
	}
}

func TestProcessor_ProcessLogLine_WithoutCID(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	logLine := "regular log line without correlation id"
	err := p.ProcessLogLine(logLine)
	if err != nil {
		t.Fatalf("ProcessLogLine() error = %v", err)
	}

	// Should not receive a record
	select {
	case record := <-outputCh:
		t.Errorf("unexpected record: %v", record)
	case <-time.After(100 * time.Millisecond):
		// Expected - no record for line without CID
	}
}

func TestProcessor_ProcessLogLine_MultipleCIDs(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	logLine := "CID[first] CID[second] multiple cids"
	err := p.ProcessLogLine(logLine)
	if err != nil {
		t.Fatalf("ProcessLogLine() error = %v", err)
	}

	// Should receive two records
	records := make([]models.CIDRecord, 0)
	timeout := time.After(time.Second)

	for len(records) < 2 {
		select {
		case record := <-outputCh:
			records = append(records, record)
		case <-timeout:
			t.Fatalf("timeout waiting for records, got %d", len(records))
		}
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestProcessor_ProcessLogLine_EmptyLine(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	err := p.ProcessLogLine("")
	if err != nil {
		t.Fatalf("ProcessLogLine() error = %v", err)
	}

	// Should not receive a record
	select {
	case record := <-outputCh:
		t.Errorf("unexpected record: %v", record)
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

func TestProcessor_Metrics(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	// Process some lines
	p.ProcessLogLine("CID[cid1] line 1")
	p.ProcessLogLine("regular line")
	p.ProcessLogLine("CID[cid2] line 2")

	// Drain the output channel
	for i := 0; i < 2; i++ {
		select {
		case <-outputCh:
		case <-time.After(time.Second):
		}
	}

	metrics := p.GetMetrics()
	processed, extracted, _, _, _ := metrics.GetStats()

	if processed != 3 {
		t.Errorf("ProcessedLogs = %d, want 3", processed)
	}

	if extracted != 2 {
		t.Errorf("ExtractedCIDs = %d, want 2", extracted)
	}
}

func TestProcessor_MetricsValidInvalid(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	// Process a line with a valid U5 UUID
	p.ProcessLogLine("CID[550e8400-e29b-51d4-a716-446655440000] valid u5")

	// Process a line without a valid UUID
	p.ProcessLogLine("CID[simple-cid] no uuid")

	// Drain the output channel
	for i := 0; i < 2; i++ {
		select {
		case <-outputCh:
		case <-time.After(time.Second):
		}
	}

	metrics := p.GetMetrics()
	_, _, valid, invalid, _ := metrics.GetStats()

	if valid != 1 {
		t.Errorf("ValidCIDs = %d, want 1", valid)
	}

	if invalid != 1 {
		t.Errorf("InvalidCIDs = %d, want 1", invalid)
	}
}

func TestProcessor_StartStop(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	p.Start()

	// Give the metrics reporter time to start
	time.Sleep(50 * time.Millisecond)

	p.Stop()

	// Should be able to call Stop without hanging
}

func TestMetrics_IncrementProcessed(t *testing.T) {
	m := &Metrics{}

	m.IncrementProcessed()
	m.IncrementProcessed()
	m.IncrementProcessed()

	processed, _, _, _, _ := m.GetStats()
	if processed != 3 {
		t.Errorf("ProcessedLogs = %d, want 3", processed)
	}
}

func TestMetrics_IncrementExtracted(t *testing.T) {
	m := &Metrics{}

	m.IncrementExtracted()
	m.IncrementExtracted()

	_, extracted, _, _, _ := m.GetStats()
	if extracted != 2 {
		t.Errorf("ExtractedCIDs = %d, want 2", extracted)
	}
}

func TestMetrics_IncrementValid(t *testing.T) {
	m := &Metrics{}

	m.IncrementValid()

	_, _, valid, _, _ := m.GetStats()
	if valid != 1 {
		t.Errorf("ValidCIDs = %d, want 1", valid)
	}
}

func TestMetrics_IncrementInvalid(t *testing.T) {
	m := &Metrics{}

	m.IncrementInvalid()
	m.IncrementInvalid()

	_, _, _, invalid, _ := m.GetStats()
	if invalid != 2 {
		t.Errorf("InvalidCIDs = %d, want 2", invalid)
	}
}

func TestMetrics_IncrementErrors(t *testing.T) {
	m := &Metrics{}

	m.IncrementErrors()

	_, _, _, _, errors := m.GetStats()
	if errors != 1 {
		t.Errorf("ProcessingErrors = %d, want 1", errors)
	}
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := &Metrics{}

	// Run concurrent increments
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			m.IncrementProcessed()
			m.IncrementExtracted()
			m.IncrementValid()
			m.IncrementInvalid()
			m.IncrementErrors()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	processed, extracted, valid, invalid, errors := m.GetStats()

	if processed != 100 {
		t.Errorf("ProcessedLogs = %d, want 100", processed)
	}
	if extracted != 100 {
		t.Errorf("ExtractedCIDs = %d, want 100", extracted)
	}
	if valid != 100 {
		t.Errorf("ValidCIDs = %d, want 100", valid)
	}
	if invalid != 100 {
		t.Errorf("InvalidCIDs = %d, want 100", invalid)
	}
	if errors != 100 {
		t.Errorf("ProcessingErrors = %d, want 100", errors)
	}
}

func TestProcessor_ContextCancellation(t *testing.T) {
	outputCh := make(chan models.CIDRecord) // Unbuffered to cause blocking
	p := NewProcessor(outputCh)

	// Cancel the context
	p.cancel()

	// ProcessLogLine should return context error when channel is full
	err := p.ProcessLogLine("CID[test] line")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestProcessor_RecordFields(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	logLine := "CID[my-test-cid] test message"
	before := time.Now()
	p.ProcessLogLine(logLine)
	after := time.Now()

	record := <-outputCh

	if record.CID != "my-test-cid" {
		t.Errorf("CID = %v, want my-test-cid", record.CID)
	}

	if record.RawLogLine != logLine {
		t.Errorf("RawLogLine = %v, want %v", record.RawLogLine, logLine)
	}

	if record.ExtractedAt.Before(before) || record.ExtractedAt.After(after) {
		t.Errorf("ExtractedAt = %v, want between %v and %v", record.ExtractedAt, before, after)
	}
}

func TestProcessor_GetMetrics(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	metrics := p.GetMetrics()
	if metrics == nil {
		t.Error("GetMetrics() should not return nil")
	}

	if metrics != p.metrics {
		t.Error("GetMetrics() should return the same metrics instance")
	}
}

func TestProcessor_ValidU5UUID(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	// U5 UUID should result in IsValid=true
	logLine := "CID[550e8400-e29b-51d4-a716-446655440000] test"
	p.ProcessLogLine(logLine)

	record := <-outputCh
	if !record.IsValid {
		t.Error("IsValid should be true for valid U5 UUID")
	}
}

func TestProcessor_InvalidUUID(t *testing.T) {
	outputCh := make(chan models.CIDRecord, 100)
	p := NewProcessor(outputCh)

	// Simple CID without valid UUID should result in IsValid=false
	logLine := "CID[simple-cid-no-uuid] test"
	p.ProcessLogLine(logLine)

	record := <-outputCh
	if record.IsValid {
		t.Error("IsValid should be false for CID without valid UUID")
	}
}
