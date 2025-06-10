package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pointlings/backend/internal/models"
)

type PointHandler struct {
	pointSpendRepo    models.PointSpendRepository
	itemRepo          models.ItemRepository
	pointlingItemRepo models.PointlingItemRepository
}

func NewPointHandler(
	pointSpendRepo models.PointSpendRepository,
	itemRepo models.ItemRepository,
	pointlingItemRepo models.PointlingItemRepository,
) *PointHandler {
	return &PointHandler{
		pointSpendRepo:    pointSpendRepo,
		itemRepo:          itemRepo,
		pointlingItemRepo: pointlingItemRepo,
	}
}

func (h *PointHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/users/{userID}", func(r chi.Router) {
		r.Post("/points/spend", h.spendPoints)
		r.Get("/points/history", h.getSpendHistory)
	})
}

type spendPointsRequest struct {
	ItemID      int64 `json:"item_id"`
	PointlingID int64 `json:"pointling_id"`
}

func (h *PointHandler) spendPoints(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req spendPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get item details to verify price
	item, err := h.itemRepo.GetByID(req.ItemID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get item")
		return
	}
	if item == nil {
		RespondError(w, http.StatusNotFound, "Item not found")
		return
	}

	if item.PricePoints == nil {
		RespondError(w, http.StatusBadRequest, "Item cannot be purchased with points")
		return
	}

	// Execute purchase transaction
	var success models.TransactionSuccess
	err = h.pointSpendRepo.InTransaction(func(txRepo models.PointSpendRepository) error {
		// Spend points
		if err := txRepo.SpendPoints(userID, req.ItemID, *item.PricePoints); err != nil {
			return err
		}

		// Add item to pointling's inventory
		if err := h.pointlingItemRepo.AddItem(req.PointlingID, req.ItemID); err != nil {
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
			RespondError(w, http.StatusPaymentRequired, "Insufficient points")
		case models.ErrAlreadyOwned:
			RespondError(w, http.StatusConflict, "Item already owned")
		default:
			RespondError(w, http.StatusInternalServerError, "Failed to process purchase")
		}
		return
	}

	RespondJSON(w, http.StatusOK, success)
}

func (h *PointHandler) getSpendHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	limit := 50 // Default limit
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	history, err := h.pointSpendRepo.GetByUser(userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get spending history")
		return
	}

	RespondJSON(w, http.StatusOK, history)
}
