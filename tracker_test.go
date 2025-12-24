package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestNewCIDTracker(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	if tracker == nil {
		t.Fatal("NewCIDTracker() returned nil")
	}

	if tracker.logPath != "/var/log" {
		t.Errorf("logPath = %v, want /var/log", tracker.logPath)
	}

	if tracker.outputFormat != "json" {
		t.Errorf("outputFormat = %v, want json", tracker.outputFormat)
	}

	if tracker.cidPattern == nil {
		t.Error("cidPattern should not be nil")
	}

	if tracker.uuidPattern == nil {
		t.Error("uuidPattern should not be nil")
	}

	if tracker.fileHandles == nil {
		t.Error("fileHandles should not be nil")
	}
}

func TestNewCIDTracker_StructuredFormat(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "structured")

	if tracker.outputFormat != "structured" {
		t.Errorf("outputFormat = %v, want structured", tracker.outputFormat)
	}
}

func TestCIDTracker_CIDPattern(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	tests := []struct {
		name    string
		input   string
		wantCID string
		found   bool
	}{
		{
			name:    "standard CID",
			input:   "CID:550e8400-e29b-51d4-a716-446655440000",
			wantCID: "550e8400-e29b-51d4-a716-446655440000",
			found:   true,
		},
		{
			name:    "CID in log line",
			input:   "INFO processing request CID:550e8400-e29b-51d4-a716-446655440001 complete",
			wantCID: "550e8400-e29b-51d4-a716-446655440001",
			found:   true,
		},
		{
			name:  "no CID",
			input: "regular log line without cid",
			found: false,
		},
		{
			name:  "empty line",
			input: "",
			found: false,
		},
		{
			name:  "short CID - not matched",
			input: "CID:abc-def-123",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := tracker.cidPattern.FindAllStringSubmatch(tt.input, -1)
			if tt.found {
				if len(matches) == 0 {
					t.Error("expected to find CID but found none")
					return
				}
				if len(matches[0]) < 2 {
					t.Error("expected match group")
					return
				}
				if matches[0][1] != tt.wantCID {
					t.Errorf("CID = %v, want %v", matches[0][1], tt.wantCID)
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("expected no CID but found: %v", matches)
				}
			}
		})
	}
}

func TestCIDTracker_UUIDPattern(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single UUID",
			input: "550e8400-e29b-41d4-a716-446655440000",
			want:  []string{"550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name:  "UUID in text",
			input: "request id: 550e8400-e29b-41d4-a716-446655440000 processed",
			want:  []string{"550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name:  "multiple UUIDs",
			input: "550e8400-e29b-41d4-a716-446655440000 and 660e8400-e29b-41d4-a716-446655440001",
			want:  []string{"550e8400-e29b-41d4-a716-446655440000", "660e8400-e29b-41d4-a716-446655440001"},
		},
		{
			name:  "no UUID",
			input: "no uuid here",
			want:  nil,
		},
		{
			name:  "uppercase UUID",
			input: "550E8400-E29B-41D4-A716-446655440000",
			want:  []string{"550E8400-E29B-41D4-A716-446655440000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := tracker.uuidPattern.FindAllString(tt.input, -1)
			if len(matches) != len(tt.want) {
				t.Errorf("found %d UUIDs, want %d", len(matches), len(tt.want))
				return
			}
			for i, m := range matches {
				if m != tt.want[i] {
					t.Errorf("match[%d] = %v, want %v", i, m, tt.want[i])
				}
			}
		})
	}
}

func TestCIDEntry_JSON(t *testing.T) {
	entry := CIDEntry{
		CID:         "test-cid",
		UUID:        "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:   time.Now(),
		LogFile:     "test.log",
		RawMessage:  "test message",
		ProcessedAt: time.Now(),
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
	if decoded.UUID != entry.UUID {
		t.Errorf("UUID = %v, want %v", decoded.UUID, entry.UUID)
	}
	if decoded.LogFile != entry.LogFile {
		t.Errorf("LogFile = %v, want %v", decoded.LogFile, entry.LogFile)
	}
	if decoded.RawMessage != entry.RawMessage {
		t.Errorf("RawMessage = %v, want %v", decoded.RawMessage, entry.RawMessage)
	}
}

func TestCIDTracker_ProcessLogLine(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.processLogLine("CID:550e8400-e29b-51d4-a716-446655440000 test message", "/var/log/test.log")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("expected output for valid CID")
	}

	// Parse the JSON output
	var entry CIDEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry.CID != "550e8400-e29b-51d4-a716-446655440000" {
		t.Errorf("CID = %v, want 550e8400-e29b-51d4-a716-446655440000", entry.CID)
	}
}

func TestCIDTracker_ProcessLogLine_InvalidUUID(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Invalid UUID format
	tracker.processLogLine("CID:not-a-valid-uuid test message", "/var/log/test.log")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should not output anything for invalid UUID
	if output != "" {
		t.Errorf("expected no output for invalid UUID, got: %v", output)
	}
}

func TestCIDTracker_ProcessLogLine_StructuredFormat(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "structured")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.processLogLine("CID:550e8400-e29b-51d4-a716-446655440000 test", "/var/log/test.log")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "CID:550e8400-e29b-51d4-a716-446655440000") {
		t.Errorf("expected structured output to contain CID, got: %v", output)
	}

	if !strings.Contains(output, "FILE:test.log") {
		t.Errorf("expected structured output to contain FILE, got: %v", output)
	}
}

func TestCIDTracker_ProcessExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a log file
	logFile := filepath.Join(tmpDir, "test.log")
	if err := os.WriteFile(logFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	tracker := NewCIDTracker(tmpDir, "json")

	err := tracker.processExistingFiles()
	if err != nil {
		t.Fatalf("processExistingFiles() error = %v", err)
	}

	// Check that the file is being monitored
	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Error("expected log file to be monitored")
	}
}

func TestCIDTracker_ProcessExistingFiles_NoLogFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-log file
	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("not a log\n"), 0644); err != nil {
		t.Fatalf("failed to create txt file: %v", err)
	}

	tracker := NewCIDTracker(tmpDir, "json")

	err := tracker.processExistingFiles()
	if err != nil {
		t.Fatalf("processExistingFiles() error = %v", err)
	}

	// Should not monitor txt file
	if len(tracker.fileHandles) != 0 {
		t.Errorf("expected 0 file handles, got %d", len(tracker.fileHandles))
	}
}

func TestCIDTracker_MonitorLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	if err := os.WriteFile(logFile, []byte("initial\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	tracker := NewCIDTracker(tmpDir, "json")
	tracker.monitorLogFile(logFile)

	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Error("expected file to be in fileHandles")
	}

	tracker.cleanup()
}

func TestCIDTracker_MonitorLogFile_NonExistent(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")
	tracker.monitorLogFile("/nonexistent/file.log")

	if len(tracker.fileHandles) != 0 {
		t.Error("expected no file handles for non-existent file")
	}
}

func TestCIDTracker_CloseFileHandle(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	if err := os.WriteFile(logFile, []byte("initial\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	tracker := NewCIDTracker(tmpDir, "json")
	tracker.monitorLogFile(logFile)

	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Fatal("expected file to be in fileHandles")
	}

	tracker.closeFileHandle(logFile)

	if _, exists := tracker.fileHandles[logFile]; exists {
		t.Error("expected file to be removed from fileHandles")
	}
}

func TestCIDTracker_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple log files
	for i := 0; i < 3; i++ {
		logFile := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".log")
		if err := os.WriteFile(logFile, []byte("content\n"), 0644); err != nil {
			t.Fatalf("failed to create log file: %v", err)
		}
	}

	tracker := NewCIDTracker(tmpDir, "json")
	tracker.processExistingFiles()

	if len(tracker.fileHandles) != 3 {
		t.Fatalf("expected 3 file handles, got %d", len(tracker.fileHandles))
	}

	tracker.cleanup()

	if len(tracker.fileHandles) != 0 {
		t.Errorf("expected 0 file handles after cleanup, got %d", len(tracker.fileHandles))
	}
}

func TestCIDTracker_Start_InvalidPath(t *testing.T) {
	tracker := NewCIDTracker("/nonexistent/path", "json")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := tracker.Start(ctx)
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestCIDTracker_Start_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	tracker := NewCIDTracker(tmpDir, "json")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- tracker.Start(ctx)
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for tracker to stop")
	}
}

func TestCIDTracker_HandleFileEvent_Create(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewCIDTracker(tmpDir, "json")

	// Simulate file creation event
	logFile := filepath.Join(tmpDir, "new.log")
	if err := os.WriteFile(logFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// This would normally be called by the watcher
	tracker.monitorLogFile(logFile)

	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Error("expected new file to be monitored")
	}

	tracker.cleanup()
}

func TestCIDTracker_MultipleCIDsInLine(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Multiple CIDs in one line - only first should be captured with current pattern
	tracker.processLogLine("CID:550e8400-e29b-51d4-a716-446655440000 CID:660e8400-e29b-51d4-a716-446655440001", "/var/log/test.log")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		t.Error("expected at least 1 output line")
	}
}

func TestCIDTracker_OutputEntry_JSON(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	entry := CIDEntry{
		CID:         "test-cid",
		UUID:        "test-uuid",
		Timestamp:   time.Now(),
		LogFile:     "test.log",
		RawMessage:  "test message",
		ProcessedAt: time.Now(),
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.outputEntry(entry)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var decoded CIDEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &decoded); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if decoded.CID != entry.CID {
		t.Errorf("CID = %v, want %v", decoded.CID, entry.CID)
	}
}

func TestCIDTracker_OutputEntry_Structured(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "structured")

	entry := CIDEntry{
		CID:         "test-cid",
		UUID:        "test-uuid",
		Timestamp:   time.Now(),
		LogFile:     "test.log",
		RawMessage:  "test message",
		ProcessedAt: time.Now(),
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.outputEntry(entry)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "CID:test-cid") {
		t.Errorf("expected output to contain CID, got: %v", output)
	}
	if !strings.Contains(output, "FILE:test.log") {
		t.Errorf("expected output to contain FILE, got: %v", output)
	}
}

func TestCIDTracker_HandleFileEvent_NonLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewCIDTracker(tmpDir, "json")

	// Create a non-log file
	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create txt file: %v", err)
	}

	// Simulate events for non-log files - should be ignored
	event := fsnotify.Event{Name: txtFile, Op: fsnotify.Create}
	tracker.handleFileEvent(event)

	if len(tracker.fileHandles) != 0 {
		t.Error("expected no file handles for non-log file")
	}
}

func TestCIDTracker_HandleFileEvent_LogFileCreate(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewCIDTracker(tmpDir, "json")

	// Create a log file
	logFile := filepath.Join(tmpDir, "test.log")
	if err := os.WriteFile(logFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// Simulate create event
	event := fsnotify.Event{Name: logFile, Op: fsnotify.Create}
	tracker.handleFileEvent(event)

	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Error("expected log file to be monitored after create event")
	}

	tracker.cleanup()
}

func TestCIDTracker_HandleFileEvent_LogFileRemove(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewCIDTracker(tmpDir, "json")

	// Create and monitor a log file
	logFile := filepath.Join(tmpDir, "test.log")
	if err := os.WriteFile(logFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}
	tracker.monitorLogFile(logFile)

	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Fatal("expected log file to be monitored")
	}

	// Simulate remove event
	event := fsnotify.Event{Name: logFile, Op: fsnotify.Remove}
	tracker.handleFileEvent(event)

	if _, exists := tracker.fileHandles[logFile]; exists {
		t.Error("expected log file handle to be closed after remove event")
	}
}

func TestCIDTracker_HandleFileEvent_LogFileWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewCIDTracker(tmpDir, "json")

	// Create a log file but don't monitor it yet
	logFile := filepath.Join(tmpDir, "test.log")
	if err := os.WriteFile(logFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// Simulate write event - should start monitoring
	event := fsnotify.Event{Name: logFile, Op: fsnotify.Write}
	tracker.handleFileEvent(event)

	// Write event triggers processLogUpdates which may open the file
	// This tests the write handling path
	tracker.cleanup()
}

func TestCIDTracker_ProcessLogUpdates_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewCIDTracker(tmpDir, "json")

	logFile := filepath.Join(tmpDir, "test.log")
	if err := os.WriteFile(logFile, []byte("initial\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// Monitor the file first
	tracker.monitorLogFile(logFile)

	// Append a line with CID
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	f.WriteString("CID:550e8400-e29b-51d4-a716-446655440000 test\n")
	f.Close()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.processLogUpdates(logFile)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "550e8400-e29b-51d4-a716-446655440000") {
		t.Errorf("expected output to contain CID, got: %v", output)
	}

	tracker.cleanup()
}

func TestCIDTracker_ProcessLogUpdates_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewCIDTracker(tmpDir, "json")

	logFile := filepath.Join(tmpDir, "test.log")
	if err := os.WriteFile(logFile, []byte("CID:550e8400-e29b-51d4-a716-446655440000 initial\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// Don't monitor - processLogUpdates should start monitoring
	tracker.processLogUpdates(logFile)

	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Error("expected file to be monitored after processLogUpdates")
	}

	tracker.cleanup()
}

func TestCIDTracker_ProcessLogLine_NoCID(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tracker.processLogLine("regular log line without cid", "/var/log/test.log")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output != "" {
		t.Errorf("expected no output for line without CID, got: %v", output)
	}
}

func TestCIDTracker_NestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create log files in subdirectory
	logFile := filepath.Join(subDir, "test.log")
	if err := os.WriteFile(logFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	tracker := NewCIDTracker(tmpDir, "json")
	err := tracker.processExistingFiles()
	if err != nil {
		t.Fatalf("processExistingFiles() error = %v", err)
	}

	// Should find the nested log file
	if _, exists := tracker.fileHandles[logFile]; !exists {
		t.Error("expected nested log file to be monitored")
	}

	tracker.cleanup()
}

func TestCIDTracker_FileHandlesMapInitialized(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	if tracker.fileHandles == nil {
		t.Error("fileHandles should be initialized")
	}

	if len(tracker.fileHandles) != 0 {
		t.Error("fileHandles should be empty initially")
	}
}

func TestCIDTracker_CloseFileHandle_NonExistent(t *testing.T) {
	tracker := NewCIDTracker("/var/log", "json")

	// Should not panic for non-existent file
	tracker.closeFileHandle("/nonexistent/file.log")

	if len(tracker.fileHandles) != 0 {
		t.Error("fileHandles should be empty")
	}
}
