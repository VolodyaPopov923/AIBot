package context

import (
	"testing"
)

func TestTokenCounter(t *testing.T) {
	tc := NewTokenCounter(1000)

	if !tc.CanAddTokens(500) {
		t.Error("Should be able to add 500 tokens")
	}

	err := tc.Add(300, 200)
	if err != nil {
		t.Fatalf("Failed to add tokens: %v", err)
	}

	if tc.TotalTokens != 500 {
		t.Errorf("Expected 500 tokens, got %d", tc.TotalTokens)
	}

	if tc.RemainingTokens() != 500 {
		t.Errorf("Expected 500 remaining tokens, got %d", tc.RemainingTokens())
	}
}

func TestContextManager(t *testing.T) {
	cm := NewContextManager(8000, 20)

	cm.AddMessage("user", "Hello")
	cm.AddMessage("assistant", "Hi there")

	messages := cm.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "user" || messages[0].Content != "Hello" {
		t.Error("First message not stored correctly")
	}
}

func TestEstimateTokens(t *testing.T) {
	text := "Hello, this is a test message"
	tokens := EstimateTokens(text)

	// Rough estimate: ~4 chars per token
	expected := (len(text) + 3) / 4
	if tokens != expected {
		t.Errorf("Expected %d tokens, got %d", expected, tokens)
	}
}
