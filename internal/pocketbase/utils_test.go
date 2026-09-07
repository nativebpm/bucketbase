package pocketbase_test

import (
	"testing"

	"github.com/nativebpm/pocketstream/internal/pocketbase"
)

func TestValidateEncryptionKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "valid 32-char hex lowercase",
			key:      "0123456789abcdef0123456789abcdef",
			expected: true,
		},
		{
			name:     "valid 32-char hex uppercase",
			key:      "0123456789ABCDEF0123456789ABCDEF",
			expected: true,
		},
		{
			name:     "invalid too short (16 chars)",
			key:      "0123456789abcdef",
			expected: false,
		},
		{
			name:     "invalid too long (33 chars)",
			key:      "0123456789abcdef0123456789abcdef0",
			expected: false,
		},
		{
			name:     "invalid non-hex character (g)",
			key:      "0123456789abcdef0123456789abcdeg",
			expected: false,
		},
		{
			name:     "empty string",
			key:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pocketbase.ValidateEncryptionKey(tt.key)
			if got != tt.expected {
				t.Errorf("ValidateEncryptionKey(%q) = %v, expected %v", tt.key, got, tt.expected)
			}
		})
	}
}
