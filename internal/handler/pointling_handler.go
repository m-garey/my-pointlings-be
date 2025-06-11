package handler

import (
	"net/http"
	"strings"

	pointling_model "my-pointlings-be/internal/handler/model"
	"my-pointlings-be/internal/service"

	"github.com/gin-gonic/gin"
)

type PointlingHandler struct {
	service service.API
}

type API interface {
	ListUsers(c *gin.Context)
	CreateUser(c *gin.Context)
	GetUser(c *gin.Context)
	UpdateUserPoints(c *gin.Context)
	CreatePointling(c *gin.Context)
	GetPointling(c *gin.Context)
	AddXP(c *gin.Context)
	GetXPHistory(c *gin.Context)
	UpdateNickname(c *gin.Context)
	ListUserPointlings(c *gin.Context)
	ListItems(c *gin.Context)
	GetItem(c *gin.Context)
	CreateItem(c *gin.Context)
	GetInventory(c *gin.Context)
	AcquireItem(c *gin.Context)
	ToggleEquipped(c *gin.Context)
	SpendPoints(c *gin.Context)
	GetSpendHistory(c *gin.Context)
}

func New(service service.API) *PointlingHandler {
	return &PointlingHandler{service: service}
}

func (h *PointlingHandler) ListUsers(c *gin.Context) {
	if _, err := h.service.ListUsers(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "listUsers"})
}

func (h *PointlingHandler) CreateUser(c *gin.Context) {
	var user pointling_model.CreateUserRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, serviceErr := h.service.CreateUser(c.Request.Context(), user); serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "createUser"})
}

func (h *PointlingHandler) GetUser(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	user, serviceErr := h.service.GetUser(c.Request.Context(), userID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *PointlingHandler) UpdateUserPoints(c *gin.Context) {
	var pointling pointling_model.UpdateUserPointsRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.service.UpdateUserPoints(c.Request.Context(), pointling); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updateUserPoints"})
}

func (h *PointlingHandler) CreatePointling(c *gin.Context) {
	var pointling pointling_model.CreatePointlingRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.service.CreatePointling(c.Request.Context(), pointling); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "createPointling"})
}

func (h *PointlingHandler) GetPointling(c *gin.Context) {
	pointlingID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("pointling_id"), "/")))
	pointling, serviceErr := h.service.GetPointling(c.Request.Context(), pointlingID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, pointling)
}

func (h *PointlingHandler) AddXP(c *gin.Context) {
	var pointling pointling_model.AddXPRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := h.service.AddXP(c.Request.Context(), pointling)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *PointlingHandler) GetXPHistory(c *gin.Context) {
	pointlingID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("pointling_id"), "/")))
	history, serviceErr := h.service.GetXPHistory(c.Request.Context(), pointlingID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *PointlingHandler) UpdateNickname(c *gin.Context) {
	var pointling pointling_model.UpdateNicknameRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.service.UpdateNickname(c.Request.Context(), pointling); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updateNickname"})
}

func (h *PointlingHandler) ListUserPointlings(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	pointlings, err := h.service.ListUserPointlings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pointlings)
}

func (h *PointlingHandler) ListItems(c *gin.Context) {
	items, err := h.service.ListItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *PointlingHandler) GetItem(c *gin.Context) {
	itemID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("item_id"), "/")))
	item, err := h.service.GetItem(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *PointlingHandler) CreateItem(c *gin.Context) {
	var item pointling_model.CreateItemRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.service.CreateItem(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "createItem"})
}

func (h *PointlingHandler) GetInventory(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	inventory, err := h.service.GetInventory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (h *PointlingHandler) AcquireItem(c *gin.Context) {
	var item pointling_model.AcquireItemRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.service.AcquireItem(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "acquireItem"})
}

func (h *PointlingHandler) ToggleEquipped(c *gin.Context) {
	var item pointling_model.ToggleEquippedRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.service.ToggleEquipped(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "toggleEquipped"})
}

func (h *PointlingHandler) SpendPoints(c *gin.Context) {
	var item pointling_model.SpendPointsRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.service.SpendPoints(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "spendPoints"})
}

func (h *PointlingHandler) GetSpendHistory(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	history, err := h.service.GetSpendHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}
