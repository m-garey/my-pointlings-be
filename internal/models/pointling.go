package models

import "time"

// Pointling Models
type Pointling struct {
	PointlingID   int64     `json:"pointling_id" db:"pointling_id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	Nickname      *string   `json:"nickname,omitempty" db:"nickname"`
	Level         int       `json:"level" db:"level"`
	CurrentXP     int       `json:"current_xp" db:"current_xp"`
	RequiredXP    int       `json:"required_xp" db:"required_xp"`
	PersonalityID *int      `json:"personality_id,omitempty" db:"personality_id"`
	LookJSON      JSONMap   `json:"look_json" db:"look_json"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type CreatePointlingRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

type UpdateNicknameRequest struct {
	PointlingID string `json:"pointling_id" binding:"required"`
	Nickname    string `json:"nickname" binding:"required"`
}

type PointlingListResponse struct {
	Pointlings []*Pointling `json:"pointlings"`
}

func NewPointling(userID int64, nickname *string) *Pointling {
	return &Pointling{
		UserID:     userID,
		Nickname:   nickname,
		Level:      1,
		CurrentXP:  0,
		RequiredXP: 3,
		LookJSON:   make(JSONMap),
	}
}

func CalculateNextLevelXP(currentLevel int) int {
	baseXP := 3
	xpPerLevel := 3
	maxXP := 120

	required := baseXP + (currentLevel-1)*xpPerLevel
	if required > maxXP {
		required = maxXP
	}
	return required
}
