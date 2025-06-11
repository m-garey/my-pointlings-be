package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"my-pointlings-be/internal/models"
	"my-pointlings-be/internal/repository"

	"github.com/gin-gonic/gin"
)

type PointlingHandler struct {
	repo repository.API
}

func NewPointlingHandler(repo repository.API) *PointlingHandler {
	return &PointlingHandler{repo: repo}
}

// Routes sets up all the pointling routes
// @Summary Set up pointling routes
// @Description Initializes all pointling-related API endpoints
// @Tags Pointlings
func (h *PointlingHandler) Routes(rg *gin.RouterGroup) {
	pointlings := rg.Group("/pointlings")
	{
		pointlings.POST("/", h.CreatePointling)
		pointlings.GET("/:pointling_id", h.GetPointling)
		pointlings.PATCH("/:pointling_id/nickname", h.UpdateNickname)
		pointlings.GET("/user/:user_id", h.ListUserPointlings)
	}
}

// createPointlingRequest represents the request body for creating a new pointling
type createPointlingRequest struct {
	UserID   int64   `json:"user_id" example:"123"`
	Nickname *string `json:"nickname,omitempty" example:"MyPointling"`
}

// CreatePointling godoc
// @Summary Create a new pointling
// @Description Create a new pointling for a user with optional nickname
// @Tags Pointlings
// @Accept json
// @Produce json
// @Param request body createPointlingRequest true "Pointling creation request"
// @Success 201 {object} models.Pointling
// @Failure 400 {object} ErrorResponse "Invalid request body or missing user_id"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings [post]
func (h *PointlingHandler) CreatePointling(c *gin.Context) {
	var req createPointlingRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	pointling := models.NewPointling(req.UserID, req.Nickname)

	if err := h.repo.CreatePointling(pointling); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pointling"})
		return
	}

	c.JSON(http.StatusCreated, pointling)
}

// GetPointling godoc
// @Summary Get pointling details
// @Description Get detailed information about a specific pointling
// @Tags Pointlings
// @Accept json
// @Produce json
// @Param pointling_id path int true "Pointling ID"
// @Success 200 {object} models.Pointling
// @Failure 400 {object} ErrorResponse "Invalid pointling ID"
// @Failure 404 {object} ErrorResponse "Pointling not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/{pointling_id} [get]
func (h *PointlingHandler) GetPointling(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("pointling_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointling ID"})
		return
	}

	pointling, err := h.repo.GetPointlingByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pointling"})
		return
	}
	if pointling == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pointling not found"})
		return
	}

	c.JSON(http.StatusOK, pointling)
}

// updateNicknameRequest represents the request body for updating a pointling's nickname
type updateNicknameRequest struct {
	Nickname *string `json:"nickname" example:"CoolPointling"`
}

// UpdateNickname godoc
// @Summary Update pointling nickname
// @Description Update or remove a pointling's nickname
// @Tags Pointlings
// @Accept json
// @Produce json
// @Param pointling_id path int true "Pointling ID"
// @Param request body updateNicknameRequest true "Nickname update request"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse "Invalid pointling ID or request body"
// @Failure 404 {object} ErrorResponse "Pointling not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/{pointling_id}/nickname [patch]
func (h *PointlingHandler) UpdateNickname(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("pointling_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointling ID"})
		return
	}

	var req updateNicknameRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.repo.UpdatePointlingNickname(id, req.Nickname); err != nil {
		if strings.Contains(err.Error(), "pointling not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pointling not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update nickname"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListUserPointlings godoc
// @Summary List user's pointlings
// @Description Get all pointlings owned by a specific user
// @Tags Pointlings
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Pointling
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/user/{user_id} [get]
func (h *PointlingHandler) ListUserPointlings(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	pointlings, err := h.repo.GetPointlingByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pointlings"})
		return
	}

	c.JSON(http.StatusOK, pointlings)
}
