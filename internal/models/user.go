package models

import "time"

// User represents a player in the Pointlings game system
type User struct {
	UserID       int64     `json:"user_id" db:"user_id"`
	DisplayName  string    `json:"display_name" db:"display_name"`
	PointBalance int64     `json:"point_balance" db:"point_balance"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// GetUser retrieves a user by ID
	GetUser(userID int64) (*User, error)

	// CreateUser creates a new user
	CreateUser(user *User) error

	// UpdatePointBalance updates a user's point balance
	UpdatePointBalance(userID int64, newBalance int64) error

	// ListUsers retrieves all users with optional limit/offset pagination
	ListUsers(limit, offset int) ([]*User, error)
}
