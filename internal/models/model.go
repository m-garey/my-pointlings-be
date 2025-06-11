package models

import (
	"encoding/json"
	"errors"
)

var (
	ErrInsufficientPoints  = errors.New("insufficient points to purchase item")
	ErrLevelTooLow         = errors.New("level requirement not met")
	ErrAlreadyOwned        = errors.New("item already owned")
	ErrInsufficientBalance = errors.New("insufficient point balance")
)

// JSONMap for look_json fields

type JSONMap map[string]interface{}

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

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
