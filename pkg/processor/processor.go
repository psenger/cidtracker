package processor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cidtracker/pkg/extractor"
	"github.com/cidtracker/pkg/models"
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

	cids, err := p.extractor.ExtractCIDs(logLine)
	if err != nil {
		p.metrics.IncrementErrors()
		return fmt.Errorf("failed to extract CIDs: %w", err)
	}

	if len(cids) == 0 {
		return nil
	}

	p.metrics.IncrementExtracted()

	for _, cid := range cids {
		record := models.CIDRecord{
			CID:         cid.Value,
			Timestamp:   time.Now(),
			RawLogLine:  logLine,
			IsValid:     cid.IsValid,
			ExtractedAt: time.Now(),
		}

		if cid.IsValid {
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