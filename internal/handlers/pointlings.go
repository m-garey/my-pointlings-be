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

// RegisterRoutes adds the pointling endpoints to the provided router
// Routes sets up all the pointling routes
func (h *PointlingHandler) Routes(rg *gin.RouterGroup) {
	pointlings := rg.Group("/pointlings")
	{
		pointlings.POST("/", h.CreatePointling)
		pointlings.GET("/:pointling_id", h.GetPointling)
		pointlings.PATCH("/:pointling_id/nickname", h.UpdateNickname)
		pointlings.GET("/user/:user_id", h.ListUserPointlings)
	}
}

type createPointlingRequest struct {
	UserID   int64   `json:"user_id"`
	Nickname *string `json:"nickname,omitempty"`
}

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

type updateNicknameRequest struct {
	Nickname *string `json:"nickname"`
}

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
