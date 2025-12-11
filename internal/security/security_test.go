package security

import (
	"testing"
)

func TestIsDestructive(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		action   string
		expected bool
	}{
		{"delete account", true},
		{"process payment", true},
		{"click button", false},
		{"fill form", false},
		{"remove item", true},
		{"logout", true},
		{"navigate", false},
	}

	for _, tt := range tests {
		result := v.IsDestructive(tt.action)
		if result != tt.expected {
			t.Errorf("IsDestructive(%s) = %v, want %v", tt.action, result, tt.expected)
		}
	}
}
