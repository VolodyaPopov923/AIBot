package agent

import (
	"testing"

	ctxmgr "github.com/VolodyaPopov923/AIBot/internal/context"
)

// TestContextResetBetweenTasks verifies that context is properly reset between tasks
func TestContextResetBetweenTasks(t *testing.T) {
	// Create a context manager
	ctx := ctxmgr.NewContextManager(8000, 20)

	// Simulate first task: add messages and consume tokens
	ctx.AddMessage("system", "System prompt for task 1")
	ctx.AddMessage("user", "User request for task 1")
	ctx.TokenCounter().Add(100, 50)

	// Check that tokens were added
	if ctx.TokenCounter().TotalTokens != 150 {
		t.Fatalf("expected 150 tokens after first task, got %d", ctx.TokenCounter().TotalTokens)
	}

	// Check that messages were added
	if len(ctx.GetMessages()) != 2 {
		t.Fatalf("expected 2 messages after first task, got %d", len(ctx.GetMessages()))
	}

	// Reset context for second task (simulating ExecuteTask call)
	ctx.ClearContext()
	ctx.ResetTokenCounter()

	// Check that context was cleared
	if len(ctx.GetMessages()) != 0 {
		t.Fatalf("expected 0 messages after reset, got %d", len(ctx.GetMessages()))
	}

	// Check that tokens were reset
	if ctx.TokenCounter().TotalTokens != 0 {
		t.Fatalf("expected 0 tokens after reset, got %d", ctx.TokenCounter().TotalTokens)
	}
	if ctx.TokenCounter().PromptTokens != 0 {
		t.Fatalf("expected 0 prompt tokens after reset, got %d", ctx.TokenCounter().PromptTokens)
	}
	if ctx.TokenCounter().CompletionTokens != 0 {
		t.Fatalf("expected 0 completion tokens after reset, got %d", ctx.TokenCounter().CompletionTokens)
	}

	// Simulate second task: add new messages and tokens
	ctx.AddMessage("system", "System prompt for task 2")
	ctx.AddMessage("user", "User request for task 2")
	ctx.TokenCounter().Add(200, 75)

	// Check that second task tokens are independent
	if ctx.TokenCounter().TotalTokens != 275 {
		t.Fatalf("expected 275 tokens after second task, got %d", ctx.TokenCounter().TotalTokens)
	}

	// Check that second task messages are independent
	if len(ctx.GetMessages()) != 2 {
		t.Fatalf("expected 2 messages for second task, got %d", len(ctx.GetMessages()))
	}
}
