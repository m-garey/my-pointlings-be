package models

import (
	"errors"
	"time"
)

var (
	ErrInsufficientBalance = errors.New("insufficient point balance")
)

// PointSpend represents a record of points spent on items
type PointSpend struct {
	SpendID     int64     `json:"spend_id" db:"spend_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	ItemID      int64     `json:"item_id" db:"item_id"`
	PointsSpent int       `json:"points_spent" db:"points_spent"`
	SpendTS     time.Time `json:"spend_ts" db:"spend_ts"`
	Item        *Item     `json:"item,omitempty" db:"-"` // Joined data
}

// TransactionSuccess represents a successful point transaction
type TransactionSuccess struct {
	ItemID        int64 `json:"item_id"`
	PointsSpent   int   `json:"points_spent"`
	NewBalance    int64 `json:"new_balance"`
	PreviousSpend *int  `json:"previous_spend,omitempty"`
}
