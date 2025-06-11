package models

type UserListResponse struct {
	Users []User `json:"users"`
}

type PointlingListResponse struct {
	Pointlings []Pointling `json:"pointlings"`
}

type XPUpdateResponse struct {
	LeveledUp  bool `json:"leveled_up"`
	NewLevel   int  `json:"new_level"`
	RequiredXP int  `json:"required_xp"`
}

type LevelUpOptionsResponse struct {
	Options []struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"options"`
}

type ItemListResponse struct {
	Items []Item `json:"items"`
}

type InventoryResponse struct {
	Items []Item `json:"items"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
