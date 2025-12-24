package extractor

import (
	"regexp"
	"strings"
	"time"

	"cidtracker/pkg/models"
	"cidtracker/pkg/validator"
)

type CIDExtractor struct {
	cidPattern *regexp.Regexp
	uuidValidator *validator.UUIDValidator
}

func NewCIDExtractor() *CIDExtractor {
	return &CIDExtractor{
		cidPattern:    regexp.MustCompile(`CID\[(\S+)\]`),
		uuidValidator: validator.NewUUIDValidator(true),
	}
}

func (e *CIDExtractor) ExtractCIDs(logLine string) []models.CIDEntry {
	matches := e.cidPattern.FindAllStringSubmatch(logLine, -1)
	if matches == nil {
		return nil
	}

	var entries []models.CIDEntry
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		cidValue := match[1]
		uuids := e.extractUUIDs(cidValue)

		entry := models.CIDEntry{
			CID: cidValue,
			Timestamp: time.Now(),
			LogLine: strings.TrimSpace(logLine),
			UUIDs: uuids,
		}

		entries = append(entries, entry)
	}

	return entries
}

func (e *CIDExtractor) extractUUIDs(cidValue string) []models.UUID {
	var uuids []models.UUID

	// Extract U5 UUIDs from CID value
	u5Pattern := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-5[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`)
	matches := u5Pattern.FindAllString(cidValue, -1)

	for _, match := range matches {
		if e.uuidValidator.IsValidU5UUID(match) {
			uuid := models.UUID{
				Value: match,
				Version: 5,
				ExtractedAt: time.Now(),
			}
			uuids = append(uuids, uuid)
		}
	}

	return uuids
}

func (e *CIDExtractor) CorrelateEntries(entries []models.CIDEntry) []models.CorrelatedEntry {
	var correlated []models.CorrelatedEntry

	for _, entry := range entries {
		corr := models.CorrelatedEntry{
			CIDEntry: entry,
			CorrelationID: e.generateCorrelationID(entry),
			ProcessedAt: time.Now(),
		}
		correlated = append(correlated, corr)
	}

	return correlated
}

func (e *CIDExtractor) generateCorrelationID(entry models.CIDEntry) string {
	// Simple correlation ID based on CID and timestamp
	return entry.CID + "_" + entry.Timestamp.Format("20060102150405")
}