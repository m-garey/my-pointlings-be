package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"my-pointlings-be/internal/models"

	"github.com/go-chi/chi/v5"
)

type PointlingHandler struct {
	repo models.PointlingRepository
}

func NewPointlingHandler(repo models.PointlingRepository) *PointlingHandler {
	return &PointlingHandler{repo: repo}
}

// RegisterRoutes adds the pointling endpoints to the provided router
func (h *PointlingHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/pointlings", func(r chi.Router) {
		r.Post("/", h.createPointling)
		r.Get("/{pointlingID}", h.getPointling)
		r.Patch("/{pointlingID}/nickname", h.updateNickname)
		r.Get("/user/{userID}", h.listUserPointlings)
	})
}

type createPointlingRequest struct {
	UserID   int64   `json:"user_id"`
	Nickname *string `json:"nickname,omitempty"`
}

func (h *PointlingHandler) createPointling(w http.ResponseWriter, r *http.Request) {
	var req createPointlingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	pointling := models.NewPointling(req.UserID, req.Nickname)

	if err := h.repo.Create(pointling); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create pointling")
		return
	}

	respondJSON(w, http.StatusCreated, pointling)
}

func (h *PointlingHandler) getPointling(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "pointlingID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid pointling ID")
		return
	}

	pointling, err := h.repo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get pointling")
		return
	}
	if pointling == nil {
		respondError(w, http.StatusNotFound, "Pointling not found")
		return
	}

	respondJSON(w, http.StatusOK, pointling)
}

type updateNicknameRequest struct {
	Nickname *string `json:"nickname"`
}

func (h *PointlingHandler) updateNickname(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "pointlingID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid pointling ID")
		return
	}

	var req updateNicknameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.repo.UpdateNickname(id, req.Nickname); err != nil {
		if strings.Contains(err.Error(), "pointling not found") {
			respondError(w, http.StatusNotFound, "Pointling not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update nickname")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PointlingHandler) listUserPointlings(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	pointlings, err := h.repo.GetByUserID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list pointlings")
		return
	}

	respondJSON(w, http.StatusOK, pointlings)
}
