package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"my-pointlings-be/internal/models"
	"my-pointlings-be/internal/repository"

	"github.com/gin-gonic/gin"
)

type XPHandler struct {
	repo repository.API
}

func NewXPHandler(repo repository.API) *XPHandler {
	return &XPHandler{
		repo: repo,
	}
}

// Routes sets up all the XP-related routes
// @Summary Set up XP routes
// @Description Initializes all XP-related API endpoints
// @Tags XP
func (h *XPHandler) Routes(rg *gin.RouterGroup) {
	pointlings := rg.Group("/pointlings/:pointling_id")
	{
		pointlings.POST("/xp", h.AddXP)
		pointlings.GET("/xp/history", h.GetXPHistory)
	}
}

// addXPRequest represents the request body for adding XP
type addXPRequest struct {
	Source models.XPEventSource `json:"source" example:"QUEST_COMPLETE" enums:"QUEST_COMPLETE,DAILY_LOGIN,ACHIEVEMENT"`
	Amount int                  `json:"amount" example:"100" minimum:"1"`
}

// addXPResponse represents the response for a successful XP addition
type addXPResponse struct {
	XPGained    int             `json:"xp_gained" example:"100"`
	NewTotal    int             `json:"new_total" example:"1250"`
	LeveledUp   bool            `json:"leveled_up" example:"true"`
	NewLevel    int             `json:"new_level,omitempty" example:"5"`
	RequiredXP  int             `json:"required_xp" example:"2000"`
	PointlingID int64           `json:"pointling_id" example:"123"`
	Event       *models.XPEvent `json:"event"`
}

// AddXP godoc
// @Summary Add XP to pointling
// @Description Award experience points to a pointling from a specific source
// @Tags XP
// @Accept json
// @Produce json
// @Param pointling_id path int true "Pointling ID"
// @Param request body addXPRequest true "XP addition request"
// @Success 200 {object} addXPResponse
// @Failure 400 {object} ErrorResponse "Invalid pointling ID, request body, or XP amount"
// @Failure 404 {object} ErrorResponse "Pointling not found"
// @Failure 429 {object} ErrorResponse "Daily XP limit reached for this source"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/{pointling_id}/xp [post]
func (h *XPHandler) AddXP(c *gin.Context) {
	pointlingID, err := strconv.ParseInt(c.Param("pointling_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointling ID"})
		return
	}

	var req addXPRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !req.Source.ValidateSource() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid XP source"})
		return
	}

	if req.Amount <= 0 || req.Amount > req.Source.GetXPPerAction() {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid XP amount, max is %d", req.Source.GetXPPerAction())})
		return
	}

	// Get current pointling state
	pointling, err := h.repo.GetPointlingByID(pointlingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pointling"})
		return
	}
	if pointling == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pointling not found"})
		return
	}

	// Create XP event
	event := &models.XPEvent{
		PointlingID: pointlingID,
		Source:      req.Source,
		XPAmount:    req.Amount,
	}

	// Process XP gain in transaction
	var resp addXPResponse
	err = h.repo.InTransaction(func(txRepo *repository.PointlingRepository) error {
		// Add XP event
		if err := txRepo.AddXP(event); err != nil {
			if err == models.ErrDailyXPLimitReached {
				return err
			}
			return fmt.Errorf("add xp: %w", err)
		}

		// Update response with new XP state
		resp.XPGained = req.Amount
		resp.NewTotal = pointling.CurrentXP + req.Amount
		resp.PointlingID = pointlingID
		resp.Event = event

		// Check for level up
		if resp.NewTotal >= pointling.RequiredXP {
			pointling.Level++
			resp.LeveledUp = true
			resp.NewLevel = pointling.Level

			// Update level and calculate new XP requirement
			newRequired := models.CalculateNextLevelXP(pointling.Level)
			if err := h.repo.UpdatePointlingLevel(pointlingID, pointling.Level); err != nil {
				return fmt.Errorf("update level: %w", err)
			}
			resp.RequiredXP = newRequired

			// Update XP with new requirement
			if err := h.repo.UpdatePointlingXP(pointlingID, resp.NewTotal, newRequired); err != nil {
				return fmt.Errorf("update xp: %w", err)
			}
		} else {
			// Update XP without level up
			resp.RequiredXP = pointling.RequiredXP
			if err := h.repo.UpdatePointlingXP(pointlingID, resp.NewTotal, pointling.RequiredXP); err != nil {
				return fmt.Errorf("update xp: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		if err == models.ErrDailyXPLimitReached {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Daily XP limit reached for this source"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process XP gain"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetXPHistory godoc
// @Summary Get XP history
// @Description Get paginated history of XP gains for a pointling
// @Tags XP
// @Accept json
// @Produce json
// @Param pointling_id path int true "Pointling ID"
// @Param limit query int false "Number of records to return (max 100)" default(50)
// @Success 200 {array} models.XPEvent
// @Failure 400 {object} ErrorResponse "Invalid pointling ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/{pointling_id}/xp/history [get]
func (h *XPHandler) GetXPHistory(c *gin.Context) {
	pointlingID, err := strconv.ParseInt(c.Param("pointling_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointling ID"})
		return
	}

	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	events, err := h.repo.GetEventsByPointling(pointlingID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get XP history"})
		return
	}

	c.JSON(http.StatusOK, events)
}
