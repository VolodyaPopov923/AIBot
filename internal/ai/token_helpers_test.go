package ai

import (
	"fmt"
	"testing"
)

func TestApproxTokens(t *testing.T) {
	s := "This is a short sentence."
	if approxTokens(s) == 0 {
		t.Fatalf("approxTokens returned 0 for non-empty string")
	}
}

func TestChunkTextByTokens(t *testing.T) {
	long := ""
	for i := 0; i < 1000; i++ {
		long += fmt.Sprintf("Sentence number %d. ", i)
	}
	limit := 60
	chunks := chunkTextByTokens(long, limit)
	if len(chunks) == 0 {
		t.Fatalf("expected chunks for long text")
	}
	// ensure no chunk exceeds the token limit
	for _, c := range chunks {
		if approxTokens(c) > limit+5 {
			t.Fatalf("chunk exceeds token limit (with slack): %d", approxTokens(c))
		}
	}
}
