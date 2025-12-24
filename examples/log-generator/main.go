package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
)

var (
	actions = []string{
		"Processing user request",
		"Fetching data from database",
		"Calling external API",
		"Validating input",
		"Generating response",
		"Updating cache",
		"Sending notification",
		"Completing transaction",
	}

	levels = []string{"INFO", "DEBUG", "WARN"}

	services = []string{
		"auth-service",
		"user-service",
		"order-service",
		"payment-service",
		"notification-service",
	}
)

func main() {
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/var/log/app"
	}

	interval := os.Getenv("LOG_INTERVAL")
	if interval == "" {
		interval = "2s"
	}

	duration, err := time.ParseDuration(interval)
	if err != nil {
		duration = 2 * time.Second
	}

	logFile := logDir + "/application.log"

	// Ensure directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	logger := log.New(file, "", 0)

	fmt.Printf("Log Generator Started\n")
	fmt.Printf("  Writing to: %s\n", logFile)
	fmt.Printf("  Interval: %s\n", duration)
	fmt.Printf("\nGenerating sample logs with CIDs...\n\n")

	// Also print to stdout for visibility
	stdLogger := log.New(os.Stdout, "", 0)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logLine := generateLogLine()
			logger.Println(logLine)
			stdLogger.Println(logLine)
		}
	}
}

func generateLogLine() string {
	// Generate a UUID v5 (namespace-based) for the CID
	// Using URL namespace with a random path for variety
	namespace := uuid.NameSpaceURL
	name := fmt.Sprintf("https://example.com/request/%d", rand.Int())
	cid := uuid.NewSHA1(namespace, []byte(name))

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	level := levels[rand.Intn(len(levels))]
	action := actions[rand.Intn(len(actions))]
	service := services[rand.Intn(len(services))]
	requestID := rand.Intn(100000)

	return fmt.Sprintf("%s %s [%s] CID:%s request_id=%d %s",
		timestamp,
		level,
		service,
		cid.String(),
		requestID,
		action,
	)
}
