package processor

import (
	"context"
	"log"
	"sync"
	"time"

	"cidtracker/pkg/extractor"
	"cidtracker/pkg/models"
)

type Metrics struct {
	ProcessedLogs    int64
	ExtractedCIDs    int64
	ValidCIDs        int64
	InvalidCIDs      int64
	ProcessingErrors int64
	mu               sync.RWMutex
}

func (m *Metrics) IncrementProcessed() {
	m.mu.Lock()
	m.ProcessedLogs++
	m.mu.Unlock()
}

func (m *Metrics) IncrementExtracted() {
	m.mu.Lock()
	m.ExtractedCIDs++
	m.mu.Unlock()
}

func (m *Metrics) IncrementValid() {
	m.mu.Lock()
	m.ValidCIDs++
	m.mu.Unlock()
}

func (m *Metrics) IncrementInvalid() {
	m.mu.Lock()
	m.InvalidCIDs++
	m.mu.Unlock()
}

func (m *Metrics) IncrementErrors() {
	m.mu.Lock()
	m.ProcessingErrors++
	m.mu.Unlock()
}

func (m *Metrics) GetStats() (int64, int64, int64, int64, int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ProcessedLogs, m.ExtractedCIDs, m.ValidCIDs, m.InvalidCIDs, m.ProcessingErrors
}

type Processor struct {
	extractor *extractor.CIDExtractor
	metrics   *Metrics
	outputCh  chan<- models.CIDRecord
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewProcessor(outputCh chan<- models.CIDRecord) *Processor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Processor{
		extractor: extractor.NewCIDExtractor(),
		metrics:   &Metrics{},
		outputCh:  outputCh,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (p *Processor) ProcessLogLine(logLine string) error {
	p.metrics.IncrementProcessed()

	entries := p.extractor.ExtractCIDs(logLine)

	if len(entries) == 0 {
		return nil
	}

	p.metrics.IncrementExtracted()

	for _, entry := range entries {
		// Check if any valid UUIDs were extracted
		isValid := len(entry.UUIDs) > 0

		record := models.CIDRecord{
			CID:         entry.CID,
			Timestamp:   entry.Timestamp,
			RawLogLine:  entry.LogLine,
			IsValid:     isValid,
			ExtractedAt: time.Now(),
		}

		if isValid {
			p.metrics.IncrementValid()
		} else {
			p.metrics.IncrementInvalid()
		}

		select {
		case p.outputCh <- record:
		case <-p.ctx.Done():
			return p.ctx.Err()
		}
	}

	return nil
}

func (p *Processor) Start() {
	p.wg.Add(1)
	go p.metricsReporter()
}

func (p *Processor) Stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *Processor) metricsReporter() {
	defer p.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			processed, extracted, valid, invalid, errors := p.metrics.GetStats()
			log.Printf("Processor metrics - Processed: %d, Extracted: %d, Valid: %d, Invalid: %d, Errors: %d",
				processed, extracted, valid, invalid, errors)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Processor) GetMetrics() *Metrics {
	return p.metrics
}