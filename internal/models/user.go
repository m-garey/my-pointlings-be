package models

import "time"

// User Models

type User struct {
	UserID       int64     `json:"user_id" db:"user_id"`
	DisplayName  string    `json:"display_name" db:"display_name"`
	PointBalance int64     `json:"point_balance" db:"point_balance"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CreateUserRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	DisplayName  string `json:"display_name" binding:"required"`
	PointBalance int    `json:"point_balance" binding:"required"`
}

type UpdateUserPointsRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	PointAmount int    `json:"point_amount" binding:"required"`
}

type UserListResponse struct {
	Users []User `json:"users"`
}
