package browser

import (
	"context"
	"testing"
)

func TestNewManager(t *testing.T) {
	ctx := context.Background()
	mgr, err := NewManager(ctx)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close(ctx)

	if mgr.browser == nil {
		t.Error("Browser not initialized")
	}
}

func TestNavigate(t *testing.T) {
	ctx := context.Background()
	mgr, err := NewManager(ctx)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close(ctx)

	// Test navigation to a basic URL
	err = mgr.Navigate(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Navigation failed: %v", err)
	}

	if mgr.page.URL() == "" {
		t.Error("URL not updated after navigation")
	}
}
