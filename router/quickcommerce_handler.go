package router

import (
	"ecommerce-platform/agentic"
	"ecommerce-platform/models"
	"ecommerce-platform/quickcommerce"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	inventoryMgr     = quickcommerce.NewInventoryManager(5 * time.Minute)
	idempotencyStore = quickcommerce.NewIdempotencyStore(10 * time.Minute)
	agentCatalog     = agentic.NewInMemoryStoreCatalog()
	aiAgent          = agentic.NewInstamartAIAgent(agentCatalog)
)

func init() {
	// Seed some dark store inventory for demonstration
	inventoryMgr.SeedStock(101, 1, 100)
	inventoryMgr.SeedStock(101, 2, 50)
	inventoryMgr.SeedStock(101, 3, 75)
}

// ReserveStockRequest payload
type ReserveStockRequest struct {
	ReservationID string `json:"reservation_id" binding:"required"`
	StoreID       int64  `json:"store_id" binding:"required"`
	ItemID        int64  `json:"item_id" binding:"required"`
	UserID        int64  `json:"user_id" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required,gt=0"`
}

// AIAgentRequest payload for natural language quick commerce
type AIAgentRequest struct {
	Prompt  string `json:"prompt" binding:"required"`
	Pincode string `json:"pincode" binding:"required"`
}

// ReserveStockHandler handles atomic inventory reservations with TTL
func (r *Router) ReserveStockHandler(c *gin.Context) {
	var req ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.Error{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid request payload",
				Detail:     err.Error(),
			},
		})
		return
	}

	reservation, err := inventoryMgr.ReserveStock(c.Request.Context(), req.ReservationID, req.StoreID, req.ItemID, req.UserID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.Error{
				StatusCode: http.StatusConflict,
				Message:    "Failed to reserve stock",
				Detail:     err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "Stock reserved successfully",
		Data:       reservation,
	})
}

// CommitStockHandler confirms stock deduction on successful payment
func (r *Router) CommitStockHandler(c *gin.Context) {
	reservationID := c.Param("reservation_id")
	if err := inventoryMgr.CommitReservation(reservationID); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.Error{
				StatusCode: http.StatusBadRequest,
				Message:    "Failed to commit reservation",
				Detail:     err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "Stock reservation committed permanently",
	})
}

// AIAssistHandler runs the autonomous Instamart agent loop
func (r *Router) AIAssistHandler(c *gin.Context) {
	var req AIAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.Error{
				StatusCode: http.StatusBadRequest,
				Message:    "Prompt and pincode required",
				Detail:     err.Error(),
			},
		})
		return
	}

	result, err := aiAgent.ProcessPrompt(c.Request.Context(), req.Prompt, req.Pincode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.Error{
				StatusCode: http.StatusInternalServerError,
				Message:    "AI Agent execution failed",
				Detail:     err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "Instamart AI shopping agent generated cart plan",
		Data:       result,
	})
}
