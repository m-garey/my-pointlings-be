package pointling_handler

import (
	"net/http"

	pointling_model "my-pointlings-be/internal/handler/model"
	pointling_service "my-pointlings-be/internal/service/pointling_service"

	"github.com/gin-gonic/gin"
)

type PointlingHandler struct {
	PointlingService pointling_service.API
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
	return &PointlingHandler{PointlingService: pointlingService}
}

func (h *PointlingHandler) listUsers(c *gin.Context) {
	err := h.PointlingService.listUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "listUsers"})
}

func (h *PointlingHandler) createUser(c *gin.Context) {
	var user pointling_model.CreateUserRequest

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "createUser"})
}

func (h *PointlingHandler) getUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "getUser"})
}

func (h *PointlingHandler) updateUserPoints(c *gin.Context) {
	var pointling pointling_model.UpdateUserPointsRequest

	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, gin.H{"message": "createPointling"})
}

func (h *PointlingHandler) getPointling(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "getPointling"})
}

func (h *PointlingHandler) addXP(c *gin.Context) {
	var pointling pointling_model.AddXPRequest

	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "addXP"})
}

func (h *PointlingHandler) getXPHistory(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "getXPHistory"})
}

func (h *PointlingHandler) updateNickname(c *gin.Context) {
	var pointling pointling_model.UpdateNicknameRequest

	if err := c.ShouldBindJSON(&pointling); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updateNickname"})
}

func (h *PointlingHandler) listUserPointlings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "listUserPointlings"})
}

func (h *PointlingHandler) listItems(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "listItems"})
}

func (h *PointlingHandler) getItem(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "getItem"})
}

func (h *PointlingHandler) createItem(c *gin.Context) {
	var item pointling_model.CreateItemRequest

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "createItem"})
}

func (h *PointlingHandler) getInventory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "getInventory"})
}

func (h *PointlingHandler) acquireItem(c *gin.Context) {
	var item pointling_model.AcquireItemRequest

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, gin.H{"message": "toggleEquipped"})
}

func (h *PointlingHandler) spendPoints(c *gin.Context) {
	var item pointling_model.SpendPointsRequest

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "spendPoints"})
}

func (h *PointlingHandler) getSpendHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "getSpendHistory"})
}
