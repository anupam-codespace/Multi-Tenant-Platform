package agentic

import (
	"context"
	"testing"
)

func TestInstamartAIAgent_ProcessPrompt(t *testing.T) {
	catalog := NewInMemoryStoreCatalog()
	agent := NewInstamartAIAgent(catalog)

	prompt := "Prepare groceries for Paneer Butter Masala dinner in Whitefield"
	pincode := "560066"

	result, err := agent.ProcessPrompt(context.Background(), prompt, pincode)
	if err != nil {
		t.Fatalf("agent failed processing prompt: %v", err)
	}

	if result.DarkStore == nil || result.DarkStore.Pincode != "560066" {
		t.Errorf("expected routed to Whitefield 560066, got %+v", result.DarkStore)
	}

	if len(result.CartItems) == 0 {
		t.Fatalf("expected non-empty cart items, got %d", len(result.CartItems))
	}

	if result.Subtotal <= 0 {
		t.Errorf("expected positive subtotal, got %.2f", result.Subtotal)
	}

	// Verify that execution plan logs tool calls
	foundDarkStoreTool := false
	for _, step := range result.ExecutionPlan {
		if step.ToolName == "FindNearestDarkStore" {
			foundDarkStoreTool = true
			break
		}
	}

	if !foundDarkStoreTool {
		t.Errorf("expected FindNearestDarkStore tool step in execution plan")
	}
}
