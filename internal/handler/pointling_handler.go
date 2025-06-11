package pointling_handler

import (
	"net/http"
	"strings"

	pointling_model "my-pointlings-be/internal/handler/model"
	"my-pointlings-be/internal/service/pointling_service"

	"github.com/gin-gonic/gin"
)

type PointlingHandler struct {
	pointlingService pointling_service.API
}

type API interface {
	listUsers(c *gin.Context)
	createUser(c *gin.Context)
	getUser(c *gin.Context)
	updateUserPoints(c *gin.Context)
	createPointling(c *gin.Context)
	getPointling(c *gin.Context)
	addXP(c *gin.Context)
	getXPHistory(c *gin.Context)
	updateNickname(c *gin.Context)
	listUserPointlings(c *gin.Context)
	listItems(c *gin.Context)
	getItem(c *gin.Context)
	createItem(c *gin.Context)
	getInventory(c *gin.Context)
	acquireItem(c *gin.Context)
	toggleEquipped(c *gin.Context)
	spendPoints(c *gin.Context)
	getSpendHistory(c *gin.Context)
}

func New(pointlingService pointling_service.API) *PointlingHandler {
	return &PointlingHandler{pointlingService: pointlingService}
}

func (h *PointlingHandler) listUsers(c *gin.Context) {
	if err := h.pointlingService.ListUsers(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "listUsers"})
}

func (h *PointlingHandler) createUser(c *gin.Context) {
	var user pointling_model.CreateUserRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, serviceErr := h.pointlingService.CreateUser(c.Request.Context(), user); serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "createUser"})
}

func (h *PointlingHandler) getUser(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	user, serviceErr := h.pointlingService.GetUser(c.Request.Context(), userID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *PointlingHandler) updateUserPoints(c *gin.Context) {
	var pointling pointling_model.UpdateUserPointsRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pointlingService.UpdateUserPoints(c.Request.Context(), pointling); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updateUserPoints"})
}

func (h *PointlingHandler) createPointling(c *gin.Context) {
	var pointling pointling_model.CreatePointlingRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pointlingService.CreatePointling(c.Request.Context(), pointling); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "createPointling"})
}

func (h *PointlingHandler) getPointling(c *gin.Context) {
	pointlingID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("pointling_id"), "/")))
	pointling, serviceErr := h.pointlingService.GetPointling(c.Request.Context(), pointlingID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, pointling)
}

func (h *PointlingHandler) addXP(c *gin.Context) {
	var pointling pointling_model.AddXPRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := h.pointlingService.AddXP(c.Request.Context(), pointling)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *PointlingHandler) getXPHistory(c *gin.Context) {
	pointlingID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("pointling_id"), "/")))
	history, serviceErr := h.pointlingService.GetXPHistory(c.Request.Context(), pointlingID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *PointlingHandler) updateNickname(c *gin.Context) {
	var pointling pointling_model.UpdateNicknameRequest
	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pointlingService.UpdateNickname(c.Request.Context(), pointling); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updateNickname"})
}

func (h *PointlingHandler) listUserPointlings(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	pointlings, err := h.pointlingService.ListUserPointlings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pointlings)
}

func (h *PointlingHandler) listItems(c *gin.Context) {
	items, err := h.pointlingService.ListItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *PointlingHandler) getItem(c *gin.Context) {
	itemID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("item_id"), "/")))
	item, err := h.pointlingService.GetItem(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *PointlingHandler) createItem(c *gin.Context) {
	var item pointling_model.CreateItemRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pointlingService.CreateItem(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "createItem"})
}

func (h *PointlingHandler) getInventory(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	inventory, err := h.pointlingService.GetInventory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (h *PointlingHandler) acquireItem(c *gin.Context) {
	var item pointling_model.AcquireItemRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pointlingService.AcquireItem(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "acquireItem"})
}

func (h *PointlingHandler) toggleEquipped(c *gin.Context) {
	var item pointling_model.ToggleEquippedRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pointlingService.ToggleEquipped(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "toggleEquipped"})
}

func (h *PointlingHandler) spendPoints(c *gin.Context) {
	var item pointling_model.SpendPointsRequest
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pointlingService.SpendPoints(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "spendPoints"})
}

func (h *PointlingHandler) getSpendHistory(c *gin.Context) {
	userID := strings.TrimSpace(strings.ToUpper(strings.TrimPrefix(c.Param("user_id"), "/")))
	history, err := h.pointlingService.GetSpendHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}
