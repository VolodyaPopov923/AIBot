package context

import (
	"fmt"
)

// TokenCounter tracks token usage
type TokenCounter struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	MaxTokens        int
}

// NewTokenCounter creates a new token counter
func NewTokenCounter(maxTokens int) *TokenCounter {
	return &TokenCounter{
		MaxTokens: maxTokens,
	}
}

// CanAddTokens checks if we can add more tokens
func (tc *TokenCounter) CanAddTokens(needed int) bool {
	return tc.TotalTokens+needed <= tc.MaxTokens
}

// RemainingTokens returns available tokens
func (tc *TokenCounter) RemainingTokens() int {
	return tc.MaxTokens - tc.TotalTokens
}

// Add adds tokens to the counter
func (tc *TokenCounter) Add(prompt, completion int) error {
	tc.PromptTokens += prompt
	tc.CompletionTokens += completion
	tc.TotalTokens = tc.PromptTokens + tc.CompletionTokens

	if tc.TotalTokens > tc.MaxTokens {
		return fmt.Errorf("token limit exceeded: %d/%d", tc.TotalTokens, tc.MaxTokens)
	}
	return nil
}

// ContextManager manages conversation history and token limits
type ContextManager struct {
	messages       []Message
	tokenCounter   *TokenCounter
	maxHistorySize int
}

// Message represents a message in context
type Message struct {
	Role    string
	Content string
}

// NewContextManager creates a context manager
func NewContextManager(maxTokens, maxHistorySize int) *ContextManager {
	return &ContextManager{
		messages:       []Message{},
		tokenCounter:   NewTokenCounter(maxTokens),
		maxHistorySize: maxHistorySize,
	}
}

// AddMessage adds a message to history
func (cm *ContextManager) AddMessage(role, content string) {
	cm.messages = append(cm.messages, Message{
		Role:    role,
		Content: content,
	})

	// Keep history size manageable
	if len(cm.messages) > cm.maxHistorySize {
		// Remove oldest user messages, keep system and recent messages
		newMessages := []Message{}
		for i, msg := range cm.messages {
			if i >= len(cm.messages)-cm.maxHistorySize {
				newMessages = append(newMessages, msg)
			}
		}
		cm.messages = newMessages
	}
}

// GetMessages returns the message history
func (cm *ContextManager) GetMessages() []Message {
	return cm.messages
}

// ClearContext resets the context
func (cm *ContextManager) ClearContext() {
	cm.messages = []Message{}
}

// ResetTokenCounter resets the token counter to zero
func (cm *ContextManager) ResetTokenCounter() {
	cm.tokenCounter.PromptTokens = 0
	cm.tokenCounter.CompletionTokens = 0
	cm.tokenCounter.TotalTokens = 0
}

// RemoveOldest removes the oldest "count" messages from history
func (cm *ContextManager) RemoveOldest(count int) {
	if count <= 0 || len(cm.messages) == 0 {
		return
	}
	if count >= len(cm.messages) {
		cm.messages = []Message{}
		return
	}
	cm.messages = cm.messages[count:]
}

// TokenCounter returns the token counter
func (cm *ContextManager) TokenCounter() *TokenCounter {
	return cm.tokenCounter
}

// EstimateTokens estimates tokens for a string (rough approximation)
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token
	return (len(text) + 3) / 4
}
