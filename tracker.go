package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// CIDEntry represents a correlation ID entry with metadata
type CIDEntry struct {
	CID         string    `json:"cid"`
	UUID        string    `json:"uuid,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	LogFile     string    `json:"log_file"`
	RawMessage  string    `json:"raw_message"`
	ProcessedAt time.Time `json:"processed_at"`
}

// CIDTracker monitors log files for correlation IDs
type CIDTracker struct {
	logPath      string
	outputFormat string
	cidPattern   *regexp.Regexp
	uuidPattern  *regexp.Regexp
	watcher      *fsnotify.Watcher
	fileHandles  map[string]*os.File
}

// NewCIDTracker creates a new CID tracker instance
func NewCIDTracker(logPath, outputFormat string) *CIDTracker {
	return &CIDTracker{
		logPath:      logPath,
		outputFormat: outputFormat,
		cidPattern:   regexp.MustCompile(`CID:([a-fA-F0-9-]{36})`),
		uuidPattern:  regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`),
		fileHandles:  make(map[string]*os.File),
	}
}

// Start begins monitoring log files
func (ct *CIDTracker) Start(ctx context.Context) error {
	var err error
	ct.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer ct.watcher.Close()

	// Watch log directory
	if err := ct.watcher.Add(ct.logPath); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", ct.logPath, err)
	}

	// Process existing log files
	if err := ct.processExistingFiles(); err != nil {
		log.WithError(err).Warn("Error processing existing files")
	}

	log.WithField("path", ct.logPath).Info("Started monitoring log directory")

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			ct.cleanup()
			return nil
		case event, ok := <-ct.watcher.Events:
			if !ok {
				return nil
			}
			ct.handleFileEvent(event)
		case err, ok := <-ct.watcher.Errors:
			if !ok {
				return nil
			}
			log.WithError(err).Warn("File watcher error")
		}
	}
}

// processExistingFiles processes log files that already exist
func (ct *CIDTracker) processExistingFiles() error {
	return filepath.Walk(ct.logPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".log") {
			ct.monitorLogFile(path)
		}
		return nil
	})
}

// handleFileEvent processes file system events
func (ct *CIDTracker) handleFileEvent(event fsnotify.Event) {
	if !strings.HasSuffix(event.Name, ".log") {
		return
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		log.WithField("file", event.Name).Debug("New log file detected")
		ct.monitorLogFile(event.Name)
	case event.Op&fsnotify.Write == fsnotify.Write:
		ct.processLogUpdates(event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		log.WithField("file", event.Name).Debug("Log file removed")
		ct.closeFileHandle(event.Name)
	}
}

// monitorLogFile starts monitoring a specific log file
func (ct *CIDTracker) monitorLogFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.WithError(err).WithField("file", filePath).Warn("Failed to open log file")
		return
	}

	// Seek to end of file to monitor new entries only
	file.Seek(0, 2)
	ct.fileHandles[filePath] = file

	log.WithField("file", filePath).Debug("Started monitoring log file")
}

// processLogUpdates processes new log entries
func (ct *CIDTracker) processLogUpdates(filePath string) {
	file, exists := ct.fileHandles[filePath]
	if !exists {
		ct.monitorLogFile(filePath)
		file = ct.fileHandles[filePath]
		if file == nil {
			return
		}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		ct.processLogLine(line, filePath)
	}
}

// processLogLine extracts CIDs from a log line
func (ct *CIDTracker) processLogLine(line, filePath string) {
	matches := ct.cidPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 1 {
			cidValue := match[1]
			
			// Validate UUID format
			if _, err := uuid.Parse(cidValue); err != nil {
				log.WithFields(log.Fields{
					"cid": cidValue,
					"error": err,
				}).Debug("Invalid UUID format in CID")
				continue
			}

			entry := CIDEntry{
				CID:         cidValue,
				UUID:        cidValue,
				Timestamp:   time.Now(),
				LogFile:     filepath.Base(filePath),
				RawMessage:  line,
				ProcessedAt: time.Now(),
			}

			ct.outputEntry(entry)
		}
	}
}

// outputEntry writes the CID entry to stdout
func (ct *CIDTracker) outputEntry(entry CIDEntry) {
	switch ct.outputFormat {
	case "json":
		if data, err := json.Marshal(entry); err == nil {
			fmt.Println(string(data))
		}
	default:
		fmt.Printf("[%s] CID:%s FILE:%s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.CID,
			entry.LogFile)
	}
}

// closeFileHandle closes a file handle
func (ct *CIDTracker) closeFileHandle(filePath string) {
	if file, exists := ct.fileHandles[filePath]; exists {
		file.Close()
		delete(ct.fileHandles, filePath)
	}
}

// cleanup closes all file handles
func (ct *CIDTracker) cleanup() {
	for filePath := range ct.fileHandles {
		ct.closeFileHandle(filePath)
	}
}