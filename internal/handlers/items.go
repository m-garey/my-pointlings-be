package handlers

import (
	"net/http"
	"strconv"

	"my-pointlings-be/internal/models"
	"my-pointlings-be/internal/repository"

	"github.com/gin-gonic/gin"
)

type ItemHandler struct {
	repo repository.API
}

func NewItemHandler(
	repo repository.API,
) *ItemHandler {
	return &ItemHandler{
		repo: repo,
	}
}

// Routes sets up all the item-related routes
func (h *ItemHandler) Routes(rg *gin.RouterGroup) {
	// Item catalog endpoints
	items := rg.Group("/items")
	{
		items.GET("/", h.ListItems)       // Get available items
		items.GET("/:item_id", h.GetItem) // Get item details
		items.POST("/", h.CreateItem)     // Admin only
	}

	// Pointling inventory endpoints
	inventory := rg.Group("/pointlings/:pointling_id/items")
	{
		inventory.GET("/", h.GetInventory)                   // List owned items
		inventory.POST("/:item_id", h.AcquireItem)           // Buy/unlock item
		inventory.PATCH("/:item_id/equip", h.ToggleEquipped) // Equip/unequip item
	}
}

func (h *ItemHandler) ListItems(c *gin.Context) {
	var category *models.ItemCategory
	var rarity *models.ItemRarity
	var slot *models.ItemSlot

	if cat := c.Query("category"); cat != "" {
		catVal := models.ItemCategory(cat)
		if !catVal.ValidateCategory() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
			return
		}
		category = &catVal
	}

	if rar := c.Query("rarity"); rar != "" {
		rarVal := models.ItemRarity(rar)
		if !rarVal.ValidateRarity() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rarity"})
			return
		}
		rarity = &rarVal
	}

	if s := c.Query("slot"); s != "" {
		slotVal := models.ItemSlot(s)
		if !slotVal.ValidateSlot() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot"})
			return
		}
		slot = &slotVal
	}

	items, err := h.repo.ListItems(category, rarity, slot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *ItemHandler) GetItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	item, err := h.repo.GetItemByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get item"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

type createItemRequest struct {
	Category    models.ItemCategory `json:"category"`
	Slot        *models.ItemSlot    `json:"slot,omitempty"`
	AssetID     string              `json:"asset_id"`
	Name        string              `json:"name"`
	Rarity      models.ItemRarity   `json:"rarity"`
	PricePoints *int                `json:"price_points,omitempty"`
	UnlockLevel *int                `json:"unlock_level,omitempty"`
}

func (h *ItemHandler) CreateItem(c *gin.Context) {
	var req createItemRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !req.Category.ValidateCategory() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}
	if !req.Rarity.ValidateRarity() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rarity"})
		return
	}
	if req.Slot != nil && !req.Slot.ValidateSlot() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot"})
		return
	}
	if req.Name == "" || req.AssetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and asset_id are required"})
		return
	}

	item := &models.Item{
		Category:    req.Category,
		Slot:        req.Slot,
		AssetID:     req.AssetID,
		Name:        req.Name,
		Rarity:      req.Rarity,
		PricePoints: req.PricePoints,
		UnlockLevel: req.UnlockLevel,
	}

	if err := h.repo.CreateItem(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *ItemHandler) GetInventory(c *gin.Context) {
	pointlingID, err := strconv.ParseInt(c.Param("pointling_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointling ID"})
		return
	}

	var equipped *bool
	if e := c.Query("equipped"); e != "" {
		b := e == "true"
		equipped = &b
	}

	items, err := h.repo.GetItems(pointlingID, equipped)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *ItemHandler) AcquireItem(c *gin.Context) {
	pointlingID, err := strconv.ParseInt(c.Param("pointling_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointling ID"})
		return
	}

	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Get item details
	item, err := h.repo.GetItemByID(itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get item"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Get pointling details
	pointling, err := h.repo.GetPointlingByID(pointlingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pointling"})
		return
	}
	if pointling == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pointling not found"})
		return
	}

	// Validate requirements
	if item.UnlockLevel != nil && pointling.Level < *item.UnlockLevel {
		c.JSON(http.StatusForbidden, gin.H{"error": "Level requirement not met"})
		return
	}

	// Add item to inventory
	err = h.repo.AddItem(pointlingID, itemID)
	if err == models.ErrAlreadyOwned {
		c.JSON(http.StatusConflict, gin.H{"error": "Item already owned"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acquire item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

type toggleEquippedRequest struct {
	Equipped bool `json:"equipped"`
}

func (h *ItemHandler) ToggleEquipped(c *gin.Context) {
	pointlingID, err := strconv.ParseInt(c.Param("pointling_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointling ID"})
		return
	}

	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req toggleEquippedRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.repo.InTransaction(func(repo *repository.PointlingRepository) error {
		return repo.ToggleEquipped(pointlingID, itemID, req.Equipped)
	})
	if err != nil {
		if err.Error() == "pointling does not own this item" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle equipped status"})
		return
	}

	c.Status(http.StatusNoContent)
}
