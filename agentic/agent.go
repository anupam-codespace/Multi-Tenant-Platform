package agentic

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AgentPlanStep captures an action taken by the Agent
type AgentPlanStep struct {
	ToolName string `json:"tool_name"`
	Input    string `json:"input"`
	Output   string `json:"output"`
	Status   string `json:"status"`
}

// AgentCartItem represents an item placed into the customer's cart by the agent
type AgentCartItem struct {
	Item           CatalogItem `json:"item"`
	Quantity       int         `json:"quantity"`
	Substituted    bool        `json:"substituted"`
	OriginalSearch string      `json:"original_search,omitempty"`
}

// ShoppingAgentResult is the structured output returned by the Agentic AI
type ShoppingAgentResult struct {
	Query         string          `json:"query"`
	DarkStore     *DarkStore      `json:"dark_store"`
	ExecutionPlan []AgentPlanStep `json:"execution_plan"`
	CartItems     []AgentCartItem `json:"cart_items"`
	Subtotal      float64         `json:"subtotal"`
	ETA           string          `json:"eta"`
	Summary       string          `json:"summary"`
	ExecutionTime string          `json:"execution_time"`
}

// InstamartAIAgent coordinates autonomous catalog searching, inventory validation, and substitutions
type InstamartAIAgent struct {
	catalog AgentTools
}

func NewInstamartAIAgent(catalog AgentTools) *InstamartAIAgent {
	return &InstamartAIAgent{
		catalog: catalog,
	}
}

// ProcessPrompt runs the autonomous Agent loop against a natural language intent
func (a *InstamartAIAgent) ProcessPrompt(ctx context.Context, userPrompt string, userPincode string) (*ShoppingAgentResult, error) {
	startTime := time.Now()
	var plan []AgentPlanStep

	// 1. Tool Call: Locate Nearest Instamart Dark Store
	store, err := a.catalog.FindNearestDarkStore(ctx, userPincode)
	if err != nil {
		return nil, fmt.Errorf("failed locating dark store: %w", err)
	}

	plan = append(plan, AgentPlanStep{
		ToolName: "FindNearestDarkStore",
		Input:    fmt.Sprintf("pincode: %s", userPincode),
		Output:   fmt.Sprintf("Routed to %s (ETA: %s)", store.Name, store.AvgETA),
		Status:   "SUCCESS",
	})

	// 2. Autonomous Intent Extraction (Rule-based or LLM)
	// Example intent: "paneer butter masala with basmati rice"
	neededIngredients := []string{"paneer", "butter", "tomato puree", "garam masala", "basmati rice", "coriander"}

	var cart []AgentCartItem
	var subtotal float64

	// 3. Tool Calling Loop: Query, Check Stock, and Substitute
	for _, ing := range neededIngredients {
		// Only search if user prompt mentions it or related recipe
		if !strings.Contains(strings.ToLower(userPrompt), strings.ToLower(ing)) &&
			!strings.Contains(strings.ToLower(userPrompt), "paneer") &&
			!strings.Contains(strings.ToLower(userPrompt), "dinner") {
			continue
		}

		results, err := a.catalog.SearchCatalog(ctx, store.ID, ing)
		if err != nil || len(results) == 0 {
			plan = append(plan, AgentPlanStep{
				ToolName: "SearchCatalog",
				Input:    fmt.Sprintf("query: '%s'", ing),
				Output:   "No match found in current dark store",
				Status:   "NOT_FOUND",
			})
			continue
		}

		selected := results[0]
		substituted := false

		// Check Stock
		if !selected.InStock {
			subs, subErr := a.catalog.GetSubstitutions(ctx, selected)
			if subErr == nil && len(subs) > 0 {
				plan = append(plan, AgentPlanStep{
					ToolName: "GetSubstitutions",
					Input:    fmt.Sprintf("out of stock: %s (%s)", selected.Name, selected.Brand),
					Output:   fmt.Sprintf("Substituted with in-stock alternative: %s (%s)", subs[0].Name, subs[0].Brand),
					Status:   "SUBSTITUTED",
				})
				selected = subs[0]
				substituted = true
			} else {
				plan = append(plan, AgentPlanStep{
					ToolName: "CheckStock",
					Input:    selected.Name,
					Output:   "Out of stock and no direct substitute available",
					Status:   "OUT_OF_STOCK",
				})
				continue
			}
		}

		cart = append(cart, AgentCartItem{
			Item:           selected,
			Quantity:       1,
			Substituted:    substituted,
			OriginalSearch: ing,
		})

		subtotal += selected.Price

		plan = append(plan, AgentPlanStep{
			ToolName: "AddToCart",
			Input:    fmt.Sprintf("%s (Price: ₹%.2f)", selected.Name, selected.Price),
			Output:   "Item added to reserved cart",
			Status:   "SUCCESS",
		})
	}

	result := &ShoppingAgentResult{
		Query:         userPrompt,
		DarkStore:     store,
		ExecutionPlan: plan,
		CartItems:     cart,
		Subtotal:      subtotal,
		ETA:           store.AvgETA,
		Summary:       fmt.Sprintf("Assembled %d items from %s. Total: ₹%.2f. Delivery expected in %s.", len(cart), store.Name, subtotal, store.AvgETA),
		ExecutionTime: fmt.Sprintf("%dms", time.Since(startTime).Milliseconds()),
	}

	return result, nil
}
