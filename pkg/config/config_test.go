package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check default values
	if len(cfg.LogSources) != 1 {
		t.Errorf("LogSources length = %d, want 1", len(cfg.LogSources))
	}

	if cfg.LogSources[0].Path != "/var/log/app" {
		t.Errorf("LogSources[0].Path = %v, want /var/log/app", cfg.LogSources[0].Path)
	}

	if cfg.LogSources[0].Name != "application" {
		t.Errorf("LogSources[0].Name = %v, want application", cfg.LogSources[0].Name)
	}

	if !cfg.LogSources[0].Active {
		t.Error("LogSources[0].Active should be true")
	}

	if len(cfg.CIDPatterns) != 2 {
		t.Errorf("CIDPatterns length = %d, want 2", len(cfg.CIDPatterns))
	}

	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %v, want json", cfg.OutputFormat)
	}

	if cfg.BufferSize != 1000 {
		t.Errorf("BufferSize = %d, want 1000", cfg.BufferSize)
	}

	if cfg.FlushInterval != 5*time.Second {
		t.Errorf("FlushInterval = %v, want 5s", cfg.FlushInterval)
	}

	if cfg.WatchInterval != 100*time.Millisecond {
		t.Errorf("WatchInterval = %v, want 100ms", cfg.WatchInterval)
	}

	if !cfg.EnableU5Only {
		t.Error("EnableU5Only should be true")
	}

	if cfg.CorrelationTTL != 1*time.Hour {
		t.Errorf("CorrelationTTL = %v, want 1h", cfg.CorrelationTTL)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
}

func TestLoadFromFile_ValidConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"log_sources": [
			{
				"path": "/var/log/test",
				"name": "test",
				"patterns": ["*.log"],
				"active": true,
				"description": "Test logs"
			}
		],
		"cid_patterns": [
			{
				"name": "test_pattern",
				"regex_string": "CID:([a-fA-F0-9-]+)",
				"uuid_group": 1,
				"enabled": true
			}
		],
		"output_format": "structured",
		"output_path": "/var/output/test.json",
		"buffer_size": 500,
		"flush_interval": 10000000000,
		"watch_interval": 200000000,
		"enable_u5_only": false,
		"correlation_ttl": 7200000000000,
		"log_level": "debug"
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if len(cfg.LogSources) != 1 {
		t.Errorf("LogSources length = %d, want 1", len(cfg.LogSources))
	}

	if cfg.LogSources[0].Path != "/var/log/test" {
		t.Errorf("LogSources[0].Path = %v, want /var/log/test", cfg.LogSources[0].Path)
	}

	if cfg.OutputFormat != "structured" {
		t.Errorf("OutputFormat = %v, want structured", cfg.OutputFormat)
	}

	if cfg.BufferSize != 500 {
		t.Errorf("BufferSize = %d, want 500", cfg.BufferSize)
	}

	if cfg.EnableU5Only {
		t.Error("EnableU5Only should be false")
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}

	// Check that regex was compiled
	if len(cfg.CIDPatterns) > 0 && cfg.CIDPatterns[0].Regex == nil {
		t.Error("CIDPatterns[0].Regex should be compiled")
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadFromFile() should return error for non-existent file")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(configPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := LoadFromFile(configPath)
	if err == nil {
		t.Error("LoadFromFile() should return error for invalid JSON")
	}
}

func TestLoadFromFile_InvalidRegex(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"log_sources": [],
		"cid_patterns": [
			{
				"name": "invalid_pattern",
				"regex_string": "[invalid(regex",
				"uuid_group": 1,
				"enabled": true
			}
		],
		"output_format": "json",
		"buffer_size": 1000
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := LoadFromFile(configPath)
	if err == nil {
		t.Error("LoadFromFile() should return error for invalid regex")
	}
}

func TestConfigValidate_DefaultBufferSize(t *testing.T) {
	cfg := &Config{
		BufferSize:    0, // Invalid
		FlushInterval: 5 * time.Second,
	}

	err := cfg.validate()
	if err != nil {
		t.Fatalf("validate() error = %v", err)
	}

	if cfg.BufferSize != 1000 {
		t.Errorf("BufferSize = %d, want 1000 (default)", cfg.BufferSize)
	}
}

func TestConfigValidate_DefaultFlushInterval(t *testing.T) {
	cfg := &Config{
		BufferSize:    1000,
		FlushInterval: 0, // Invalid
	}

	err := cfg.validate()
	if err != nil {
		t.Fatalf("validate() error = %v", err)
	}

	if cfg.FlushInterval != 5*time.Second {
		t.Errorf("FlushInterval = %v, want 5s (default)", cfg.FlushInterval)
	}
}

func TestConfigValidate_NegativeBufferSize(t *testing.T) {
	cfg := &Config{
		BufferSize:    -100,
		FlushInterval: 5 * time.Second,
	}

	err := cfg.validate()
	if err != nil {
		t.Fatalf("validate() error = %v", err)
	}

	if cfg.BufferSize != 1000 {
		t.Errorf("BufferSize = %d, want 1000 (default)", cfg.BufferSize)
	}
}

func TestConfigValidate_NegativeFlushInterval(t *testing.T) {
	cfg := &Config{
		BufferSize:    1000,
		FlushInterval: -5 * time.Second,
	}

	err := cfg.validate()
	if err != nil {
		t.Fatalf("validate() error = %v", err)
	}

	if cfg.FlushInterval != 5*time.Second {
		t.Errorf("FlushInterval = %v, want 5s (default)", cfg.FlushInterval)
	}
}

func TestConfigValidate_MultiplePatterns(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"log_sources": [],
		"cid_patterns": [
			{
				"name": "pattern1",
				"regex_string": "CID:([a-fA-F0-9-]+)",
				"uuid_group": 1,
				"enabled": true
			},
			{
				"name": "pattern2",
				"regex_string": "REQ-([0-9]+)",
				"uuid_group": 1,
				"enabled": true
			}
		],
		"output_format": "json",
		"buffer_size": 1000
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if len(cfg.CIDPatterns) != 2 {
		t.Errorf("CIDPatterns length = %d, want 2", len(cfg.CIDPatterns))
	}

	for i, pattern := range cfg.CIDPatterns {
		if pattern.Regex == nil {
			t.Errorf("CIDPatterns[%d].Regex should be compiled", i)
		}
	}
}

func TestDefaultConfig_CIDPatterns(t *testing.T) {
	cfg := DefaultConfig()

	// Verify default patterns
	if len(cfg.CIDPatterns) != 2 {
		t.Fatalf("expected 2 default CID patterns, got %d", len(cfg.CIDPatterns))
	}

	// Check standard_cid pattern
	standardPattern := cfg.CIDPatterns[0]
	if standardPattern.Name != "standard_cid" {
		t.Errorf("first pattern name = %v, want standard_cid", standardPattern.Name)
	}
	if !standardPattern.Enabled {
		t.Error("standard_cid pattern should be enabled")
	}
	if standardPattern.UUIDGroup != 1 {
		t.Errorf("standard_cid UUIDGroup = %d, want 1", standardPattern.UUIDGroup)
	}

	// Check json_cid pattern
	jsonPattern := cfg.CIDPatterns[1]
	if jsonPattern.Name != "json_cid" {
		t.Errorf("second pattern name = %v, want json_cid", jsonPattern.Name)
	}
	if !jsonPattern.Enabled {
		t.Error("json_cid pattern should be enabled")
	}
}

func TestLoadFromFile_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.json")

	configContent := `{}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Should have default buffer size after validation
	if cfg.BufferSize != 1000 {
		t.Errorf("BufferSize = %d, want 1000 (default)", cfg.BufferSize)
	}
}

func TestLoadFromFile_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.json")

	configContent := `{
		"output_format": "structured",
		"log_level": "warn"
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.OutputFormat != "structured" {
		t.Errorf("OutputFormat = %v, want structured", cfg.OutputFormat)
	}

	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %v, want warn", cfg.LogLevel)
	}
}
