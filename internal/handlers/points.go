package handlers

import (
	"net/http"
	"strconv"

	"my-pointlings-be/internal/models"
	"my-pointlings-be/internal/repository"

	"github.com/gin-gonic/gin"
)

type PointHandler struct {
	repo repository.API
}

func NewPointHandler(
	repo repository.API,
) *PointHandler {
	return &PointHandler{
		repo: repo,
	}
}

func (h *PointHandler) Routes(rg *gin.RouterGroup) {
	users := rg.Group("/points/:userID")
	{
		users.POST("/spend", h.spendPoints)
		users.GET("/history", h.getSpendHistory)
	}
}

type spendPointsRequest struct {
	ItemID      int64 `json:"item_id"`
	PointlingID int64 `json:"pointling_id"`
}

func (h *PointHandler) spendPoints(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		RespondError(c.Writer, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req spendPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c.Writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get item details to verify price
	item, err := h.repo.GetItemByID(req.ItemID)
	if err != nil {
		RespondError(c.Writer, http.StatusInternalServerError, "Failed to get item")
		return
	}
	if item == nil {
		RespondError(c.Writer, http.StatusNotFound, "Item not found")
		return
	}

	if item.PricePoints == nil {
		RespondError(c.Writer, http.StatusBadRequest, "Item cannot be purchased with points")
		return
	}

	// Execute purchase transaction
	var success models.TransactionSuccess
	err = h.repo.InTransaction(func(txRepo *repository.PointlingRepository) error {
		// Spend points
		if err := txRepo.SpendPoints(userID, req.ItemID, *item.PricePoints); err != nil {
			return err
		}

		// Add item to pointling's inventory
		if err := h.repo.AddItem(req.PointlingID, req.ItemID); err != nil {
			return err
		}

		// Get updated balance for response
		total, err := txRepo.GetTotalSpentByUser(userID)
		if err != nil {
			return err
		}

		success = models.TransactionSuccess{
			ItemID:      req.ItemID,
			PointsSpent: *item.PricePoints,
			NewBalance:  total,
		}
		return nil
	})

	if err != nil {
		switch err {
		case models.ErrInsufficientBalance:
			RespondError(c.Writer, http.StatusPaymentRequired, "Insufficient points")
		case models.ErrAlreadyOwned:
			RespondError(c.Writer, http.StatusConflict, "Item already owned")
		default:
			RespondError(c.Writer, http.StatusInternalServerError, "Failed to process purchase")
		}
		return
	}

	RespondJSON(c.Writer, http.StatusOK, success)
}

func (h *PointHandler) getSpendHistory(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		RespondError(c.Writer, http.StatusBadRequest, "Invalid user ID")
		return
	}

	limit := 50 // Default limit
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	history, err := h.repo.GetUserPointHistory(userID, limit, offset)
	if err != nil {
		RespondError(c.Writer, http.StatusInternalServerError, "Failed to get spending history")
		return
	}

	RespondJSON(c.Writer, http.StatusOK, history)
}
