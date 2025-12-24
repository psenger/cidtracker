package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"cidtracker/pkg/models"
)

// Config holds the application configuration
type Config struct {
	LogSources      []models.LogSource   `json:"log_sources"`
	CIDPatterns     []models.CIDPattern  `json:"cid_patterns"`
	OutputFormat    string               `json:"output_format"`
	OutputPath      string               `json:"output_path"`
	BufferSize      int                  `json:"buffer_size"`
	FlushInterval   time.Duration        `json:"flush_interval"`
	WatchInterval   time.Duration        `json:"watch_interval"`
	EnableU5Only    bool                 `json:"enable_u5_only"`
	CorrelationTTL  time.Duration        `json:"correlation_ttl"`
	LogLevel        string               `json:"log_level"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		LogSources: []models.LogSource{
			{
				Path:        "/var/log/app",
				Name:        "application",
				Patterns:    []string{"*.log"},
				Active:      true,
				Description: "Main application logs",
			},
		},
		CIDPatterns: []models.CIDPattern{
			{
				Name:        "standard_cid",
				RegexString: `CID\s*[=:]\s*([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})`,
				UUIDGroup:   1,
				Enabled:     true,
			},
			{
				Name:        "json_cid",
				RegexString: `"cid"\s*:\s*"([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})"`,
				UUIDGroup:   1,
				Enabled:     true,
			},
		},
		OutputFormat:   "json",
		OutputPath:     "/var/output/cid-tracker.json",
		BufferSize:     1000,
		FlushInterval:  5 * time.Second,
		WatchInterval:  100 * time.Millisecond,
		EnableU5Only:   true,
		CorrelationTTL: 1 * time.Hour,
		LogLevel:       "info",
	}
}

// LoadFromFile loads configuration from JSON file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, config.validate()
}

// validate compiles regex patterns and validates configuration
func (c *Config) validate() error {
	for i := range c.CIDPatterns {
		regex, err := regexp.Compile(c.CIDPatterns[i].RegexString)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", c.CIDPatterns[i].Name, err)
		}
		c.CIDPatterns[i].Regex = regex
	}

	if c.BufferSize <= 0 {
		c.BufferSize = 1000
	}

	if c.FlushInterval <= 0 {
		c.FlushInterval = 5 * time.Second
	}

	return nil
}
