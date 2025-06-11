package models

import (
	"errors"
	"time"
)

// XP Models

type XPEventSource string

const (
	XPSourceReceipt XPEventSource = "RECEIPT"
	XPSourcePlay    XPEventSource = "PLAY"
	XPSourceDaily   XPEventSource = "DAILY"

	MaxDailyReceiptXP = 50
	MaxDailyPlayXP    = 100
	MaxDailyLoginXP   = 10
)

var ErrDailyXPLimitReached = errors.New("daily XP limit reached for this source")

type XPEvent struct {
	EventID     int64         `json:"event_id" db:"event_id"`
	PointlingID int64         `json:"pointling_id" db:"pointling_id"`
	Source      XPEventSource `json:"source" db:"source"`
	XPAmount    int           `json:"xp_amount" db:"xp_amount"`
	EventTS     time.Time     `json:"event_ts" db:"event_ts"`
}

type AddXPRequest struct {
	PointlingID string `json:"pointling_id" binding:"required"`
	XPGain      int    `json:"xp_gain" binding:"required"`
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

func (s XPEventSource) ValidateSource() bool {
	switch s {
	case XPSourceReceipt, XPSourcePlay, XPSourceDaily:
		return true
	default:
		return false
	}
}

func (s XPEventSource) GetMaxDailyXP() int {
	switch s {
	case XPSourceReceipt:
		return MaxDailyReceiptXP
	case XPSourcePlay:
		return MaxDailyPlayXP
	case XPSourceDaily:
		return MaxDailyLoginXP
	default:
		return 0
	}
}

func (s XPEventSource) GetXPPerAction() int {
	switch s {
	case XPSourceReceipt:
		return 10
	case XPSourcePlay:
		return 20
	case XPSourceDaily:
		return 10
	default:
		return 0
	}
}
