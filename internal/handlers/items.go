package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pointlings/backend/internal/models"
)

type ItemHandler struct {
	itemRepo          models.ItemRepository
	pointlingRepo     models.PointlingRepository
	pointlingItemRepo models.PointlingItemRepository
}

func NewItemHandler(
	itemRepo models.ItemRepository,
	pointlingRepo models.PointlingRepository,
	pointlingItemRepo models.PointlingItemRepository,
) *ItemHandler {
	return &ItemHandler{
		itemRepo:          itemRepo,
		pointlingRepo:     pointlingRepo,
		pointlingItemRepo: pointlingItemRepo,
	}
}

func (h *ItemHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		// Item catalog endpoints
		r.Route("/items", func(r chi.Router) {
			r.Get("/", h.listItems)       // Get available items
			r.Get("/{itemID}", h.getItem) // Get item details
			r.Post("/", h.createItem)     // Admin only
		})

		// Pointling inventory endpoints
		r.Route("/pointlings/{pointlingID}/items", func(r chi.Router) {
			r.Get("/", h.getInventory)                   // List owned items
			r.Post("/{itemID}", h.acquireItem)           // Buy/unlock item
			r.Patch("/{itemID}/equip", h.toggleEquipped) // Equip/unequip item
		})
	})
}

func (h *ItemHandler) listItems(w http.ResponseWriter, r *http.Request) {
	var category *models.ItemCategory
	var rarity *models.ItemRarity
	var slot *models.ItemSlot

	if cat := r.URL.Query().Get("category"); cat != "" {
		c := models.ItemCategory(cat)
		if !c.ValidateCategory() {
			RespondError(w, http.StatusBadRequest, "Invalid category")
			return
		}
		category = &c
	}

	if rar := r.URL.Query().Get("rarity"); rar != "" {
		r := models.ItemRarity(rar)
		if !r.ValidateRarity() {
			RespondError(w, http.StatusBadRequest, "Invalid rarity")
			return
		}
		rarity = &r
	}

	if s := r.URL.Query().Get("slot"); s != "" {
		sl := models.ItemSlot(s)
		if !sl.ValidateSlot() {
			RespondError(w, http.StatusBadRequest, "Invalid slot")
			return
		}
		slot = &sl
	}

	items, err := h.itemRepo.List(category, rarity, slot)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list items")
		return
	}

	RespondJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) getItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	item, err := h.itemRepo.GetByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get item")
		return
	}
	if item == nil {
		RespondError(w, http.StatusNotFound, "Item not found")
		return
	}

	RespondJSON(w, http.StatusOK, item)
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

func (h *ItemHandler) createItem(w http.ResponseWriter, r *http.Request) {
	var req createItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !req.Category.ValidateCategory() {
		RespondError(w, http.StatusBadRequest, "Invalid category")
		return
	}
	if !req.Rarity.ValidateRarity() {
		RespondError(w, http.StatusBadRequest, "Invalid rarity")
		return
	}
	if req.Slot != nil && !req.Slot.ValidateSlot() {
		RespondError(w, http.StatusBadRequest, "Invalid slot")
		return
	}
	if req.Name == "" || req.AssetID == "" {
		RespondError(w, http.StatusBadRequest, "Name and asset_id are required")
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

	if err := h.itemRepo.Create(item); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create item")
		return
	}

	RespondJSON(w, http.StatusCreated, item)
}

func (h *ItemHandler) getInventory(w http.ResponseWriter, r *http.Request) {
	pointlingID, err := strconv.ParseInt(chi.URLParam(r, "pointlingID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid pointling ID")
		return
	}

	var equipped *bool
	if e := r.URL.Query().Get("equipped"); e != "" {
		b := e == "true"
		equipped = &b
	}

	items, err := h.pointlingItemRepo.GetItems(pointlingID, equipped)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get inventory")
		return
	}

	RespondJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) acquireItem(w http.ResponseWriter, r *http.Request) {
	pointlingID, err := strconv.ParseInt(chi.URLParam(r, "pointlingID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid pointling ID")
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	// Get item details
	item, err := h.itemRepo.GetByID(itemID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get item")
		return
	}
	if item == nil {
		RespondError(w, http.StatusNotFound, "Item not found")
		return
	}

	// Get pointling details
	pointling, err := h.pointlingRepo.GetByID(pointlingID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get pointling")
		return
	}
	if pointling == nil {
		RespondError(w, http.StatusNotFound, "Pointling not found")
		return
	}

	// Validate requirements
	if item.UnlockLevel != nil && pointling.Level < *item.UnlockLevel {
		RespondError(w, http.StatusForbidden, "Level requirement not met")
		return
	}

	// Add item to inventory
	err = h.pointlingItemRepo.AddItem(pointlingID, itemID)
	if err == models.ErrAlreadyOwned {
		RespondError(w, http.StatusConflict, "Item already owned")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to acquire item")
		return
	}

	RespondJSON(w, http.StatusOK, item)
}

type toggleEquippedRequest struct {
	Equipped bool `json:"equipped"`
}

func (h *ItemHandler) toggleEquipped(w http.ResponseWriter, r *http.Request) {
	pointlingID, err := strconv.ParseInt(chi.URLParam(r, "pointlingID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid pointling ID")
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var req toggleEquippedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.pointlingItemRepo.InTransaction(func(repo models.PointlingItemRepository) error {
		return repo.ToggleEquipped(pointlingID, itemID, req.Equipped)
	})
	if err != nil {
		if err.Error() == "pointling does not own this item" {
			RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, "Failed to toggle equipped status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
