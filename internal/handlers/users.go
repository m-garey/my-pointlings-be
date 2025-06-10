package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pointlings/backend/internal/models"
)

type UserHandler struct {
	repo models.UserRepository
}

func NewUserHandler(repo models.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Get("/", h.listUsers)
		r.Post("/", h.createUser)
		r.Get("/{userID}", h.getUser)
		r.Patch("/{userID}/points", h.updatePoints)
	})
}

type createUserRequest struct {
	UserID      int64  `json:"user_id"`
	DisplayName string `json:"display_name"`
}

type updatePointsRequest struct {
	NewBalance int64 `json:"new_balance"`
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 || req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "user_id and display_name are required")
		return
	}

	user := &models.User{
		UserID:       req.UserID,
		DisplayName:  req.DisplayName,
		PointBalance: 0, // New users start with 0 points
	}

	if err := h.repo.CreateUser(user); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.repo.GetUser(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) updatePoints(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req updatePointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.NewBalance < 0 {
		respondError(w, http.StatusBadRequest, "Point balance cannot be negative")
		return
	}

	err = h.repo.UpdatePointBalance(userID, req.NewBalance)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			respondError(w, http.StatusNotFound, "User not found")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to update points")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) {
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

	users, err := h.repo.ListUsers(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	respondJSON(w, http.StatusOK, users)
}
