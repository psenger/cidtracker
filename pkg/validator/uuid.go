package validator

import (
	"fmt"

	"cidtracker/pkg/models"
	"github.com/google/uuid"
)

// UUIDValidator handles UUID validation and version checking
type UUIDValidator struct {
	enforceU5Only bool
}

// NewUUIDValidator creates a new UUID validator
func NewUUIDValidator(enforceU5Only bool) *UUIDValidator {
	return &UUIDValidator{
		enforceU5Only: enforceU5Only,
	}
}

// ValidateUUID validates a UUID string and returns detailed results
func (v *UUIDValidator) ValidateUUID(uuidStr string) models.ValidationResult {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return models.ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("invalid UUID format: %v", err),
		}
	}

	version := v.getUUIDVersion(parsedUUID)
	variant := v.getUUIDVariant(parsedUUID)

	result := models.ValidationResult{
		Valid:   true,
		Version: version,
		Variant: variant,
	}

	// If enforcing U5 only, check version
	if v.enforceU5Only && version != 5 {
		result.Valid = false
		result.Error = fmt.Sprintf("expected UUID version 5, got version %d", version)
	}

	return result
}

// IsValidCID checks if a UUID string is valid for CID tracking
func (v *UUIDValidator) IsValidCID(uuidStr string) bool {
	result := v.ValidateUUID(uuidStr)
	return result.Valid && (!v.enforceU5Only || result.IsU5UUID())
}

// getUUIDVersion extracts the version from a UUID
func (v *UUIDValidator) getUUIDVersion(u uuid.UUID) int {
	return int((u[6] & 0xf0) >> 4)
}

// getUUIDVariant extracts the variant from a UUID
func (v *UUIDValidator) getUUIDVariant(u uuid.UUID) string {
	// Check variant bits in byte 8
	// NCS: 0xxxxxxx (high bit is 0)
	// RFC4122: 10xxxxxx (high bits are 10)
	// Microsoft: 110xxxxx (high bits are 110)
	// Reserved: 111xxxxx (high bits are 111)
	b := u[8]
	if (b & 0x80) == 0 {
		return "NCS"
	}
	if (b & 0xc0) == 0x80 {
		return "RFC4122"
	}
	if (b & 0xe0) == 0xc0 {
		return "Microsoft"
	}
	return "Reserved"
}

// IsValidU5UUID checks if a UUID string is a valid version 5 UUID
func (v *UUIDValidator) IsValidU5UUID(uuidStr string) bool {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return false
	}
	return v.getUUIDVersion(parsedUUID) == 5
}
