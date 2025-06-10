package models

import (
	"encoding/json"
	"time"
)

// JSONMap is a type alias for JSON objects stored in the database
type JSONMap map[string]interface{}

// Scan implements sql.Scanner for JSONMap to handle JSONB columns
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		*m = make(JSONMap)
	}
	return nil
}

// Pointling represents a user's virtual pet character
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

// PointlingRepository defines the interface for pointling data access
type PointlingRepository interface {
	// Create stores a new pointling
	Create(pointling *Pointling) error

	// GetByID retrieves a pointling by its ID
	GetByID(id int64) (*Pointling, error)

	// GetByUserID retrieves all pointlings owned by a user
	GetByUserID(userID int64) ([]*Pointling, error)

	// UpdateLook updates a pointling's appearance
	UpdateLook(id int64, look JSONMap) error

	// UpdateXP updates a pointling's XP and required XP values
	UpdateXP(id int64, currentXP, requiredXP int) error

	// UpdateLevel updates a pointling's level
	UpdateLevel(id int64, level int) error

	// UpdateNickname sets a pointling's nickname
	UpdateNickname(id int64, nickname *string) error
}

// NewPointling creates a new Pointling with default values
func NewPointling(userID int64, nickname *string) *Pointling {
	return &Pointling{
		UserID:     userID,
		Nickname:   nickname,
		Level:      1,
		CurrentXP:  0,
		RequiredXP: 3, // Initial XP requirement
		LookJSON:   make(JSONMap),
	}
}

// CalculateNextLevelXP returns the XP required for the next level
// XP grows linearly from 3 to 120
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
