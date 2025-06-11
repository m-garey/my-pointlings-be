package service_model

// ---------- USERS ----------
type User struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	PointlingID string `json:"pointling_id"`
	PointAmount int    `json:"point_amount"`
}

type CreateUserRequest struct {
	UserName string `json:"user_name"`
}

type UserListResponse struct {
	Users []User `json:"users"`
}

type UpdateUserPointsRequest struct {
	UserID      string
	PointAmount int `json:"point_amount"`
}

type SpendPointsRequest struct {
	UserID string
	Amount int    `json:"amount"`
	ItemID string `json:"item_id"`
}

type SpendHistoryEntry struct {
	ItemID    string `json:"item_id"`
	Amount    int    `json:"amount"`
	Timestamp string `json:"timestamp"`
}

type SpendHistoryResponse struct {
	History []SpendHistoryEntry `json:"history"`
}

// ---------- POINTLINGS ----------
type Pointling struct {
	PointlingID  string `json:"pointling_id"`
	Name         string `json:"pointling_name"`
	CurrentXP    int    `json:"current_xp"`
	RequiredXP   int    `json:"required_xp"`
	Level        int    `json:"current_level"`
	AppearanceID string `json:"appearance_id"`
	WardrobeID   string `json:"wardrobe_id"`
}

type PointlingListResponse struct {
	Pointlings []Pointling `json:"pointlings"`
}

type CreatePointlingRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type XPUpdateRequest struct {
	PointlingID string `json:"pointling_id"`
	XPGain      int    `json:"xp_gain"`
}

type XPUpdateResponse struct {
	LeveledUp  bool `json:"leveled_up"`
	NewLevel   int  `json:"new_level"`
	RequiredXP int  `json:"required_xp"`
}

type XPHistoryEntry struct {
	XP        int    `json:"xp"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}

type XPHistoryResponse struct {
	History []XPHistoryEntry `json:"history"`
}

type UpdateNicknameRequest struct {
	PointlingID string `json:"pointling_id"`
	Nickname    string `json:"nickname"`
}

type LevelUpOptionsResponse struct {
	Options []struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"options"`
}

type ChooseLevelUpItemRequest struct {
	ChoiceID string `json:"choice_id"`
}

// ---------- ITEMS ----------
type Item struct {
	ItemID string `json:"item_id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Rarity string `json:"rarity"`
	Cost   int    `json:"cost"`
	Image  string `json:"image"`
}

type CreateItemRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Rarity string `json:"rarity"`
	Cost   int    `json:"cost"`
	Image  string `json:"image"`
}

type ItemListResponse struct {
	Items []Item `json:"items"`
}

// ---------- INVENTORY ----------
type InventoryResponse struct {
	Items []Item `json:"items"`
}

type AcquireItemRequest struct {
	ItemID      string `json:"item_id"`
	PointlingID string `json:"pointling_id"`
}

type ToggleEquippedRequest struct {
	ItemID      string `json:"item_id"`
	PointlingID string `json:"pointling_id"`
	Equipped    bool   `json:"equipped"`
}

// ---------- GENERIC RESPONSE ----------
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
