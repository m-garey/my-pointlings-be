package models

import "time"

// Transaction Models

type SpendPointsRequest struct {
	UserID string `json:"user_id" binding:"required"`
	ItemID string `json:"item_id" binding:"required"`
	Amount int    `json:"amount" binding:"required"`
}

type PointSpend struct {
	SpendID     int64     `json:"spend_id" db:"spend_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	ItemID      int64     `json:"item_id" db:"item_id"`
	PointsSpent int       `json:"points_spent" db:"points_spent"`
	SpendTS     time.Time `json:"spend_ts" db:"spend_ts"`
	Item        *Item     `json:"item,omitempty" db:"-"`
}

type TransactionSuccess struct {
	ItemID        int64 `json:"item_id"`
	PointsSpent   int   `json:"points_spent"`
	NewBalance    int64 `json:"new_balance"`
	PreviousSpend *int  `json:"previous_spend,omitempty"`
}
