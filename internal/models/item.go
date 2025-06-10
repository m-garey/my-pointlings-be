package models

import (
	"errors"
	"time"
)

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

var (
	ErrInsufficientPoints = errors.New("insufficient points to purchase item")
	ErrLevelTooLow        = errors.New("level requirement not met")
	ErrAlreadyOwned       = errors.New("item already owned")
)

// Item represents a purchasable or unlockable item
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

// PointlingItem represents an item owned by a pointling
type PointlingItem struct {
	PointlingID int64     `json:"pointling_id" db:"pointling_id"`
	ItemID      int64     `json:"item_id" db:"item_id"`
	AcquiredAt  time.Time `json:"acquired_at" db:"acquired_at"`
	Equipped    bool      `json:"equipped" db:"equipped"`
	Item        *Item     `json:"item,omitempty" db:"-"` // Joined data
}

// ItemRepository defines the interface for item data access
type ItemRepository interface {
	// GetByID retrieves an item by its ID
	GetByID(id int64) (*Item, error)

	// List retrieves items with optional filters
	List(category *ItemCategory, rarity *ItemRarity, slot *ItemSlot) ([]*Item, error)

	// GetUnlocksForLevel gets items available at a specific level
	GetUnlocksForLevel(level int) ([]*Item, error)

	// Create creates a new item (admin only)
	Create(item *Item) error
}

// PointlingItemRepository defines the interface for pointling-item relationship
type PointlingItemRepository interface {
	// AddItem gives an item to a pointling
	AddItem(pointlingID, itemID int64) error

	// GetItems lists items owned by a pointling
	GetItems(pointlingID int64, equipped *bool) ([]*PointlingItem, error)

	// ToggleEquipped equips or unequips an item
	ToggleEquipped(pointlingID, itemID int64, equipped bool) error

	// GetEquippedInSlot gets the currently equipped item in a slot
	GetEquippedInSlot(pointlingID int64, slot ItemSlot) (*PointlingItem, error)

	// InTransaction executes operations in a transaction
	InTransaction(fn func(PointlingItemRepository) error) error
}

// ValidateCategory checks if the category is valid
func (c ItemCategory) ValidateCategory() bool {
	switch c {
	case CategoryAccessory, CategoryFeature:
		return true
	default:
		return false
	}
}

// ValidateSlot checks if the slot is valid
func (s ItemSlot) ValidateSlot() bool {
	switch s {
	case SlotHat, SlotShoes, SlotFace, SlotWings:
		return true
	default:
		return false
	}
}

// ValidateRarity checks if the rarity is valid
func (r ItemRarity) ValidateRarity() bool {
	switch r {
	case RarityCommon, RarityRare, RarityEpic, RarityLegendary:
		return true
	default:
		return false
	}
}

// GetRarityValue returns a numeric value for sorting by rarity
func (r ItemRarity) GetRarityValue() int {
	switch r {
	case RarityCommon:
		return 1
	case RarityRare:
		return 2
	case RarityEpic:
		return 3
	case RarityLegendary:
		return 4
	default:
		return 0
	}
}
