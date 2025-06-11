package models

import "time"

// Item Models

type ItemCategory string

type ItemSlot string

type ItemRarity string

const (
	CategoryAccessory ItemCategory = "ACCESSORY"
	CategoryFeature   ItemCategory = "FEATURE"

	SlotHat   ItemSlot = "HAT"
	SlotShoes ItemSlot = "SHOES"
	SlotFace  ItemSlot = "FACE"
	SlotWings ItemSlot = "WINGS"

	RarityCommon    ItemRarity = "COMMON"
	RarityRare      ItemRarity = "RARE"
	RarityEpic      ItemRarity = "EPIC"
	RarityLegendary ItemRarity = "LEGENDARY"
)

type Item struct {
	ItemID      int64        `json:"item_id" db:"item_id"`
	Category    ItemCategory `json:"category" db:"category"`
	Slot        *ItemSlot    `json:"slot,omitempty" db:"slot"`
	AssetID     string       `json:"asset_id" db:"asset_id"`
	Name        string       `json:"name" db:"name"`
	Rarity      ItemRarity   `json:"rarity" db:"rarity"`
	PricePoints *int         `json:"price_points,omitempty" db:"price_points"`
	UnlockLevel *int         `json:"unlock_level,omitempty" db:"unlock_level"`
}

type PointlingItem struct {
	PointlingID int64     `json:"pointling_id" db:"pointling_id"`
	ItemID      int64     `json:"item_id" db:"item_id"`
	AcquiredAt  time.Time `json:"acquired_at" db:"acquired_at"`
	Equipped    bool      `json:"equipped" db:"equipped"`
	Item        *Item     `json:"item,omitempty" db:"-"`
}

type CreateItemRequest struct {
	Name     string `json:"name" binding:"required"`
	Cost     int    `json:"cost" binding:"required"`
	Category string `json:"category" binding:"required"`
	Rarity   string `json:"rarity"`
	AssetID  string `json:"asset_id" binding:"required"`
}

type AcquireItemRequest struct {
	PointlingID string `json:"pointling_id" binding:"required"`
	ItemID      string `json:"item_id" binding:"required"`
}

type ToggleEquippedRequest struct {
	PointlingID string `json:"pointling_id" binding:"required"`
	ItemID      string `json:"item_id" binding:"required"`
	Equipped    bool   `json:"equipped" binding:"required"`
}

type ItemListResponse struct {
	Items []Item `json:"items"`
}

type InventoryResponse struct {
	Items []Item `json:"items"`
}
