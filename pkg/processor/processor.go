package processor

import (
	"sync"
	"time"

	"github.com/cidtracker/pkg/extractor"
	"github.com/cidtracker/pkg/models"
)

type LogProcessor struct {
	extractor *extractor.CIDExtractor
	entries   []models.CorrelatedEntry
	mu        sync.RWMutex
}

func NewLogProcessor() *LogProcessor {
	return &LogProcessor{
		extractor: extractor.NewCIDExtractor(),
		entries:   make([]models.CorrelatedEntry, 0),
	}
}

func (p *LogProcessor) ProcessLogLine(logLine string) error {
	cidEntries := p.extractor.ExtractCIDs(logLine)
	if len(cidEntries) == 0 {
		return nil
	}

	correlated := p.extractor.CorrelateEntries(cidEntries)

	p.mu.Lock()
	p.entries = append(p.entries, correlated...)
	p.mu.Unlock()

	return nil
}

func (p *LogProcessor) GetEntries() []models.CorrelatedEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]models.CorrelatedEntry(nil), p.entries...)
}

func (p *LogProcessor) GetEntriesSince(since time.Time) []models.CorrelatedEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var filtered []models.CorrelatedEntry
	for _, entry := range p.entries {
		if entry.ProcessedAt.After(since) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (p *LogProcessor) ClearOldEntries(olderThan time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	var kept []models.CorrelatedEntry

	for _, entry := range p.entries {
		if entry.ProcessedAt.After(cutoff) {
			kept = append(kept, entry)
		}
	}

	p.entries = kept
}