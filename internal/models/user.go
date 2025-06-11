package models

import "time"

// User represents a player in the Pointlings game system
type User struct {
	UserID       int64     `json:"user_id" db:"user_id"`
	DisplayName  string    `json:"display_name" db:"display_name"`
	PointBalance int64     `json:"point_balance" db:"point_balance"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
