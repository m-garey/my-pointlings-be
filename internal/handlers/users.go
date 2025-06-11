package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"my-pointlings-be/internal/models"
	"my-pointlings-be/internal/repository"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	repo repository.API
}

func NewUserHandler(repo repository.API) *UserHandler {
	return &UserHandler{repo: repo}
}

// Routes sets up all the user routes
func (h *UserHandler) Routes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("/", h.ListUsers)
		users.POST("/", h.CreateUser)
		users.GET("/:user_id", h.GetUser)
		users.PATCH("/:user_id/points", h.UpdatePoints)
	}
}

type createUserRequest struct {
	UserID      int64  `json:"user_id"`
	DisplayName string `json:"display_name"`
}

type updatePointsRequest struct {
	NewBalance int64 `json:"new_balance"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.UserID == 0 || req.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and display_name are required"})
		return
	}

	user := &models.User{
		UserID:       req.UserID,
		DisplayName:  req.DisplayName,
		PointBalance: 0, // New users start with 0 points
	}

	if err := h.repo.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.repo.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdatePoints(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req updatePointsRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.NewBalance < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Point balance cannot be negative"})
		return
	}

	err = h.repo.UpdatePointBalance(userID, req.NewBalance)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update points"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
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

	users, err := h.repo.ListUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, users)
}
