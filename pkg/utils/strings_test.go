package utils

import (
	"testing"
)

func TestTruncateText(t *testing.T) {
	text := "This is a long text"
	result := TruncateText(text, 10)

	if len(result) > 13 {
		t.Errorf("Text not truncated properly: %s", result)
	}
}

func TestCleanText(t *testing.T) {
	text := "  Hello   world  \n  test  "
	result := CleanText(text)

	if result != "Hello world test" {
		t.Errorf("Text not cleaned properly: got %q", result)
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "https://example.com"},
		{"https://example.com", "https://example.com"},
		{"http://example.com", "http://example.com"},
	}

	for _, tt := range tests {
		result := NormalizeURL(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeURL(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !StringInSlice("apple", slice) {
		t.Error("StringInSlice failed to find existing element")
	}

	if StringInSlice("grape", slice) {
		t.Error("StringInSlice found non-existing element")
	}
}
