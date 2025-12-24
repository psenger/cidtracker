package validator

import (
	"testing"
)

func TestNewUUIDValidator(t *testing.T) {
	tests := []struct {
		name          string
		enforceU5Only bool
	}{
		{"with U5 enforcement", true},
		{"without U5 enforcement", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewUUIDValidator(tt.enforceU5Only)
			if v == nil {
				t.Fatal("expected non-nil validator")
			}
			if v.enforceU5Only != tt.enforceU5Only {
				t.Errorf("enforceU5Only = %v, want %v", v.enforceU5Only, tt.enforceU5Only)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name          string
		uuidStr       string
		enforceU5Only bool
		wantValid     bool
		wantVersion   int
		wantVariant   string
		wantError     bool
	}{
		{
			name:          "valid UUID v4",
			uuidStr:       "550e8400-e29b-41d4-a716-446655440000",
			enforceU5Only: false,
			wantValid:     true,
			wantVersion:   4,
			wantVariant:   "RFC4122",
			wantError:     false,
		},
		{
			name:          "valid UUID v5",
			uuidStr:       "550e8400-e29b-51d4-a716-446655440000",
			enforceU5Only: false,
			wantValid:     true,
			wantVersion:   5,
			wantVariant:   "RFC4122",
			wantError:     false,
		},
		{
			name:          "valid UUID v5 with U5 enforcement",
			uuidStr:       "550e8400-e29b-51d4-a716-446655440000",
			enforceU5Only: true,
			wantValid:     true,
			wantVersion:   5,
			wantVariant:   "RFC4122",
			wantError:     false,
		},
		{
			name:          "valid UUID v4 with U5 enforcement - should fail",
			uuidStr:       "550e8400-e29b-41d4-a716-446655440000",
			enforceU5Only: true,
			wantValid:     false,
			wantVersion:   4,
			wantVariant:   "RFC4122",
			wantError:     true,
		},
		{
			name:          "invalid UUID format",
			uuidStr:       "not-a-uuid",
			enforceU5Only: false,
			wantValid:     false,
			wantVersion:   0,
			wantVariant:   "",
			wantError:     true,
		},
		{
			name:          "empty string",
			uuidStr:       "",
			enforceU5Only: false,
			wantValid:     false,
			wantVersion:   0,
			wantVariant:   "",
			wantError:     true,
		},
		{
			name:          "UUID v1",
			uuidStr:       "550e8400-e29b-11d4-a716-446655440000",
			enforceU5Only: false,
			wantValid:     true,
			wantVersion:   1,
			wantVariant:   "RFC4122",
			wantError:     false,
		},
		{
			name:          "UUID with uppercase",
			uuidStr:       "550E8400-E29B-51D4-A716-446655440000",
			enforceU5Only: false,
			wantValid:     true,
			wantVersion:   5,
			wantVariant:   "RFC4122",
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewUUIDValidator(tt.enforceU5Only)
			result := v.ValidateUUID(tt.uuidStr)

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if tt.wantValid && !tt.wantError {
				if result.Version != tt.wantVersion {
					t.Errorf("Version = %v, want %v", result.Version, tt.wantVersion)
				}
				if result.Variant != tt.wantVariant {
					t.Errorf("Variant = %v, want %v", result.Variant, tt.wantVariant)
				}
			}

			if tt.wantError && result.Error == "" && !result.Valid {
				// For invalid results, we expect an error message
				if tt.uuidStr == "" || tt.uuidStr == "not-a-uuid" {
					if result.Error == "" {
						t.Error("expected error message for invalid UUID")
					}
				}
			}
		})
	}
}

func TestIsValidCID(t *testing.T) {
	tests := []struct {
		name          string
		uuidStr       string
		enforceU5Only bool
		want          bool
	}{
		{"valid v5 UUID with enforcement", "550e8400-e29b-51d4-a716-446655440000", true, true},
		{"valid v4 UUID with enforcement", "550e8400-e29b-41d4-a716-446655440000", true, false},
		{"valid v4 UUID without enforcement", "550e8400-e29b-41d4-a716-446655440000", false, true},
		{"invalid UUID", "invalid", false, false},
		{"empty string", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewUUIDValidator(tt.enforceU5Only)
			got := v.IsValidCID(tt.uuidStr)
			if got != tt.want {
				t.Errorf("IsValidCID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidU5UUID(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		want    bool
	}{
		{"valid v5 UUID", "550e8400-e29b-51d4-a716-446655440000", true},
		{"valid v4 UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid v1 UUID", "550e8400-e29b-11d4-a716-446655440000", false},
		{"invalid UUID", "not-a-uuid", false},
		{"empty string", "", false},
		{"valid v5 uppercase", "550E8400-E29B-51D4-A716-446655440000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewUUIDValidator(false)
			got := v.IsValidU5UUID(tt.uuidStr)
			if got != tt.want {
				t.Errorf("IsValidU5UUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUUIDVersion(t *testing.T) {
	v := NewUUIDValidator(false)

	tests := []struct {
		name        string
		uuidStr     string
		wantVersion int
	}{
		{"version 1", "550e8400-e29b-11d4-a716-446655440000", 1},
		{"version 4", "550e8400-e29b-41d4-a716-446655440000", 4},
		{"version 5", "550e8400-e29b-51d4-a716-446655440000", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateUUID(tt.uuidStr)
			if result.Version != tt.wantVersion {
				t.Errorf("version = %v, want %v", result.Version, tt.wantVersion)
			}
		})
	}
}

func TestGetUUIDVariant(t *testing.T) {
	v := NewUUIDValidator(false)

	tests := []struct {
		name        string
		uuidStr     string
		wantVariant string
	}{
		{"RFC4122 variant", "550e8400-e29b-41d4-a716-446655440000", "RFC4122"},
		{"RFC4122 variant v5", "550e8400-e29b-51d4-b716-446655440000", "RFC4122"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateUUID(tt.uuidStr)
			if result.Variant != tt.wantVariant {
				t.Errorf("variant = %v, want %v", result.Variant, tt.wantVariant)
			}
		})
	}
}

func TestGetUUIDVariant_AllVariants(t *testing.T) {
	v := NewUUIDValidator(false)

	// Test NCS variant (high bit = 0, i.e., 0xxxxxxx)
	// We need to construct a UUID with byte 8 having the pattern 0xxxxxxx
	// For example: 550e8400-e29b-41d4-3716-446655440000 (3 = 0011, high bit is 0)
	ncsResult := v.ValidateUUID("550e8400-e29b-41d4-3716-446655440000")
	if ncsResult.Variant != "NCS" {
		t.Errorf("NCS variant: got %v, want NCS", ncsResult.Variant)
	}

	// Test Microsoft variant (high bits = 110, i.e., 110xxxxx)
	// Byte 8 should be 0xC0-0xDF (e.g., c716)
	msResult := v.ValidateUUID("550e8400-e29b-41d4-c716-446655440000")
	if msResult.Variant != "Microsoft" {
		t.Errorf("Microsoft variant: got %v, want Microsoft", msResult.Variant)
	}

	// Test Reserved variant (high bits = 111, i.e., 111xxxxx)
	// Byte 8 should be 0xE0-0xFF (e.g., e716)
	reservedResult := v.ValidateUUID("550e8400-e29b-41d4-e716-446655440000")
	if reservedResult.Variant != "Reserved" {
		t.Errorf("Reserved variant: got %v, want Reserved", reservedResult.Variant)
	}
}
