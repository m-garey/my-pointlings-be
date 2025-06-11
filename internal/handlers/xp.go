package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"my-pointlings-be/internal/models"

	"github.com/go-chi/chi/v5"
)

type XPHandler struct {
	xpRepo        models.XPRepository
	pointlingRepo models.PointlingRepository
}

func NewXPHandler(xpRepo models.XPRepository, pointlingRepo models.PointlingRepository) *XPHandler {
	return &XPHandler{
		xpRepo:        xpRepo,
		pointlingRepo: pointlingRepo,
	}
}

func (h *XPHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/pointlings/{pointlingID}", func(r chi.Router) {
		r.Post("/xp", h.addXP)
		r.Get("/xp/history", h.getXPHistory)
	})
}

type addXPRequest struct {
	Source models.XPEventSource `json:"source"`
	Amount int                  `json:"amount"`
}

type addXPResponse struct {
	XPGained    int             `json:"xp_gained"`
	NewTotal    int             `json:"new_total"`
	LeveledUp   bool            `json:"leveled_up"`
	NewLevel    int             `json:"new_level,omitempty"`
	RequiredXP  int             `json:"required_xp"`
	PointlingID int64           `json:"pointling_id"`
	Event       *models.XPEvent `json:"event"`
}

func (h *XPHandler) addXP(w http.ResponseWriter, r *http.Request) {
	pointlingID, err := strconv.ParseInt(chi.URLParam(r, "pointlingID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid pointling ID")
		return
	}

	var req addXPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !req.Source.ValidateSource() {
		respondError(w, http.StatusBadRequest, "Invalid XP source")
		return
	}

	if req.Amount <= 0 || req.Amount > req.Source.GetXPPerAction() {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid XP amount, max is %d", req.Source.GetXPPerAction()))
		return
	}

	// Get current pointling state
	pointling, err := h.pointlingRepo.GetByID(pointlingID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get pointling")
		return
	}
	if pointling == nil {
		respondError(w, http.StatusNotFound, "Pointling not found")
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
	err = h.xpRepo.InTransaction(func(txRepo models.XPRepository) error {
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
			if err := h.pointlingRepo.UpdateLevel(pointlingID, pointling.Level); err != nil {
				return fmt.Errorf("update level: %w", err)
			}
			resp.RequiredXP = newRequired

			// Update XP with new requirement
			if err := h.pointlingRepo.UpdateXP(pointlingID, resp.NewTotal, newRequired); err != nil {
				return fmt.Errorf("update xp: %w", err)
			}
		} else {
			// Update XP without level up
			resp.RequiredXP = pointling.RequiredXP
			if err := h.pointlingRepo.UpdateXP(pointlingID, resp.NewTotal, pointling.RequiredXP); err != nil {
				return fmt.Errorf("update xp: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		if err == models.ErrDailyXPLimitReached {
			respondError(w, http.StatusTooManyRequests, "Daily XP limit reached for this source")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to process XP gain")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *XPHandler) getXPHistory(w http.ResponseWriter, r *http.Request) {
	pointlingID, err := strconv.ParseInt(chi.URLParam(r, "pointlingID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid pointling ID")
		return
	}

	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	events, err := h.xpRepo.GetEventsByPointling(pointlingID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get XP history")
		return
	}

	respondJSON(w, http.StatusOK, events)
}
