package validator

import (
	"fmt"

	"github.com/cidtracker/pkg/models"
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
	variantBits := (u[8] & 0xc0) >> 6
	switch variantBits {
	case 0:
		return "NCS"
	case 1:
		return "RFC4122"
	case 2:
		return "Microsoft"
	case 3:
		return "Reserved"
	default:
		return "Unknown"
	}
}
