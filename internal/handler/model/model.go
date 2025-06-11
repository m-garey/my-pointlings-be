package pointling_model

// Structures to bind requests for Pointlings API endpoints

// USERS

// POST /users
// CreateUserRequest holds data to create a user
type CreateUserRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	UserName    string `json:"name" binding:"required"`
	PointlingID string `json:"pointling_id" binding:"required"`
	PointAmount int    `json:"point_amount" binding:"required"`
}

type User struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	PointlingID string `json:"pointling_id"`
	PointAmount int    `json:"point_amount"`
}

// PATCH /users/:user_id/points
// UpdateUserPointsRequest holds data to update user points
type UpdateUserPointsRequest struct {
	PointAmount int `json:"point_amount" binding:"required"`
}

// POINTLINGS

type Pointling struct {
	PointlingID  string `json:"pointling_id"`
	Name         string `json:"pointling_name"`
	CurrentXP    int    `json:"current_xp"`
	RequiredXP   int    `json:"required_xp"`
	Level        int    `json:"current_level"`
	AppearanceID string `json:"appearance_id"`
	WardrobeID   string `json:"wardrobe_id"`
}

// POST /pointlings
// CreatePointlingRequest holds data to create a pointling
type CreatePointlingRequest struct {
	PointlingName string `json:"pointling_name" binding:"required"`
}

// POST /pointlings/:pointling_id/xp
// AddXPRequest holds XP to be added
type AddXPRequest struct {
	XPGain int `json:"xp_gain" binding:"required"`
}

type XPHistoryEntry struct {
	XP        int    `json:"xp"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}

type XPHistoryResponse struct {
	History []XPHistoryEntry `json:"history"`
}

// PATCH /pointlings/:pointling_id/nickname
// UpdateNicknameRequest holds new nickname for a pointling
type UpdateNicknameRequest struct {
	Nickname string `json:"nickname" binding:"required"`
}

// ITEMS

// POST /items
// CreateItemRequest holds data to create an item
// (Assuming structure based on accessories and items)
type CreateItemRequest struct {
	Name    string `json:"name" binding:"required"`
	Cost    int    `json:"cost" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Rarity  string `json:"rarity"`
	AssetID string `json:"asset_id" binding:"required"`
}

// POINTLING INVENTORY

// POST /pointlings/:pointling_id/items/:item_id
// AcquireItemRequest holds optional metadata if needed for acquisition
// (Empty struct here assuming path params are sufficient)
type AcquireItemRequest struct{}

// PATCH /pointlings/:pointling_id/items/:item_id/equip
// ToggleEquippedRequest handles toggling equip state
type ToggleEquippedRequest struct {
	Equipped bool `json:"equipped" binding:"required"`
}

// POINT SPENDING

// POST /users/:user_id/points/spend
// SpendPointsRequest holds data to spend user points
type SpendPointsRequest struct {
	Amount int    `json:"amount" binding:"required"`
	Reason string `json:"reason"`
}
