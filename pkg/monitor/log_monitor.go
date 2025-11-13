package monitor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Line      string    `json:"line"`
	Source    string    `json:"source"`
}

type LogMonitor struct {
	watcher   *fsnotify.Watcher
	logPaths  []string
	outputCh  chan LogEntry
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	fileTails map[string]*os.File
}

func NewLogMonitor(logPaths []string) (*LogMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	lm := &LogMonitor{
		watcher:   watcher,
		logPaths:  logPaths,
		outputCh:  make(chan LogEntry, 1000),
		ctx:       ctx,
		cancel:    cancel,
		fileTails: make(map[string]*os.File),
	}

	for _, path := range logPaths {
		if err := lm.addWatchPath(path); err != nil {
			return nil, fmt.Errorf("failed to add watch path %s: %w", path, err)
		}
	}

	return lm, nil
}

func (lm *LogMonitor) addWatchPath(path string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Watch directory for file creation
		dir := filepath.Dir(path)
		if err := lm.watcher.Add(dir); err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", dir, err)
		}
		return nil
	}

	// File exists, start tailing it
	if err := lm.startTailing(path); err != nil {
		return fmt.Errorf("failed to start tailing %s: %w", path, err)
	}

	return lm.watcher.Add(path)
}

func (lm *LogMonitor) startTailing(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}

	// Seek to end of file to only read new entries
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek to end of file %s: %w", path, err)
	}

	lm.fileTails[path] = file

	// Start reading goroutine
	go lm.tailFile(path, file)

	return nil
}

func (lm *LogMonitor) tailFile(path string, file *os.File) {
	scanner := bufio.NewScanner(file)
	for {
		select {
		case <-lm.ctx.Done():
			return
		default:
			if scanner.Scan() {
				entry := LogEntry{
					Timestamp: time.Now(),
					Line:      scanner.Text(),
					Source:    path,
				}
				select {
				case lm.outputCh <- entry:
				case <-lm.ctx.Done():
					return
				}
			} else {
				// No new lines, wait a bit
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (lm *LogMonitor) Start() <-chan LogEntry {
	go lm.watchEvents()
	return lm.outputCh
}

func (lm *LogMonitor) watchEvents() {
	for {
		select {
		case <-lm.ctx.Done():
			return
		case event, ok := <-lm.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				// New file created, check if it's one we should monitor
				lm.handleFileCreation(event.Name)
			} else if event.Op&fsnotify.Write == fsnotify.Write {
				// File modified, the tailFile goroutine will handle this
				continue
			}

		case err, ok := <-lm.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue monitoring
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

func (lm *LogMonitor) handleFileCreation(fileName string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if this file matches any of our log paths
	for _, logPath := range lm.logPaths {
		if fileName == logPath {
			if err := lm.startTailing(fileName); err != nil {
				fmt.Printf("Failed to start tailing new file %s: %v\n", fileName, err)
			}
			break
		}
	}
}

func (lm *LogMonitor) Stop() error {
	lm.cancel()

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Close all file handles
	for path, file := range lm.fileTails {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file %s: %v\n", path, err)
		}
	}

	close(lm.outputCh)
	return lm.watcher.Close()
}