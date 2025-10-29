package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Command line flags
	logPath := flag.String("log-path", "/var/log/app", "Path to mounted docker logs directory")
	outputFormat := flag.String("output", "json", "Output format: json or structured")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Configure logging
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})

	log.WithFields(log.Fields{
		"version": "0.1.0",
		"log_path": *logPath,
		"output_format": *outputFormat,
	}).Info("Starting CID Tracker")

	// Validate log path exists
	if _, err := os.Stat(*logPath); os.IsNotExist(err) {
		log.WithField("path", *logPath).Fatal("Log path does not exist")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize tracker
	tracker := NewCIDTracker(*logPath, *outputFormat)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal, stopping tracker...")
		cancel()
	}()

	// Start monitoring
	if err := tracker.Start(ctx); err != nil {
		log.WithError(err).Fatal("Failed to start CID tracker")
	}

	log.Info("CID Tracker stopped")
}