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
// @Summary Set up item routes
// @Description Initializes all item-related API endpoints
// @Tags Items
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

// ListItems godoc
// @Summary List available items
// @Description Get a list of items with optional filtering by category, rarity, and slot
// @Tags Items
// @Accept json
// @Produce json
// @Param category query string false "Filter by item category (e.g. COSMETIC, CONSUMABLE)"
// @Param rarity query string false "Filter by item rarity (e.g. COMMON, RARE, EPIC)"
// @Param slot query string false "Filter by item slot (e.g. HEAD, BODY, ACCESSORY)"
// @Success 200 {array} models.Item
// @Failure 400 {object} ErrorResponse "Invalid category/rarity/slot"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /items [get]
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

// GetItem godoc
// @Summary Get item details
// @Description Get detailed information about a specific item
// @Tags Items
// @Accept json
// @Produce json
// @Param item_id path int true "Item ID"
// @Success 200 {object} models.Item
// @Failure 400 {object} ErrorResponse "Invalid item ID"
// @Failure 404 {object} ErrorResponse "Item not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /items/{item_id} [get]
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

// createItemRequest represents the request body for creating a new item
type createItemRequest struct {
	Category    models.ItemCategory `json:"category" example:"COSMETIC"`
	Slot        *models.ItemSlot    `json:"slot,omitempty" example:"HEAD"`
	AssetID     string              `json:"asset_id" example:"hat_001"`
	Name        string              `json:"name" example:"Cool Hat"`
	Rarity      models.ItemRarity   `json:"rarity" example:"RARE"`
	PricePoints *int                `json:"price_points,omitempty" example:"100"`
	UnlockLevel *int                `json:"unlock_level,omitempty" example:"5"`
}

// CreateItem godoc
// @Summary Create a new item
// @Description Create a new item in the catalog (admin only)
// @Tags Items
// @Accept json
// @Produce json
// @Param request body createItemRequest true "Item creation request"
// @Success 201 {object} models.Item
// @Failure 400 {object} ErrorResponse "Invalid request body/parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /items [post]
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

// GetInventory godoc
// @Summary Get pointling's inventory
// @Description List all items owned by a pointling with optional equipped filter
// @Tags Items
// @Accept json
// @Produce json
// @Param pointling_id path int true "Pointling ID"
// @Param equipped query boolean false "Filter by equipped status"
// @Success 200 {array} models.Item
// @Failure 400 {object} ErrorResponse "Invalid pointling ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/{pointling_id}/items [get]
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

// AcquireItem godoc
// @Summary Acquire an item for a pointling
// @Description Add an item to a pointling's inventory (requires meeting level requirements)
// @Tags Items
// @Accept json
// @Produce json
// @Param pointling_id path int true "Pointling ID"
// @Param item_id path int true "Item ID"
// @Success 200 {object} models.Item
// @Failure 400 {object} ErrorResponse "Invalid pointling/item ID"
// @Failure 403 {object} ErrorResponse "Level requirement not met"
// @Failure 404 {object} ErrorResponse "Pointling/item not found"
// @Failure 409 {object} ErrorResponse "Item already owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/{pointling_id}/items/{item_id} [post]
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

// toggleEquippedRequest represents the request body for equipping/unequipping an item
type toggleEquippedRequest struct {
	Equipped bool `json:"equipped" example:"true"`
}

// ToggleEquipped godoc
// @Summary Toggle item equipped status
// @Description Equip or unequip an item for a pointling
// @Tags Items
// @Accept json
// @Produce json
// @Param pointling_id path int true "Pointling ID"
// @Param item_id path int true "Item ID"
// @Param request body toggleEquippedRequest true "Toggle equipped request"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse "Invalid pointling/item ID or request body"
// @Failure 404 {object} ErrorResponse "Pointling does not own this item"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /pointlings/{pointling_id}/items/{item_id}/equip [patch]
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
