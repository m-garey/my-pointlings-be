package models

import (
	"errors"
	"time"
)

// XPEventSource represents the type of activity that generated XP
type XPEventSource string

const (
	XPSourceReceipt XPEventSource = "RECEIPT"
	XPSourcePlay    XPEventSource = "PLAY"
	XPSourceDaily   XPEventSource = "DAILY"
)

// XP limits per source
const (
	MaxDailyReceiptXP = 50  // 10 XP per receipt, max 5 receipts
	MaxDailyPlayXP    = 100 // Can be earned through games
	MaxDailyLoginXP   = 10  // Once per day bonus
)

// ErrDailyXPLimitReached indicates that no more XP can be earned from this source today
var ErrDailyXPLimitReached = errors.New("daily XP limit reached for this source")

// XPEvent represents a record of XP gain
type XPEvent struct {
	EventID     int64         `json:"event_id" db:"event_id"`
	PointlingID int64         `json:"pointling_id" db:"pointling_id"`
	Source      XPEventSource `json:"source" db:"source"`
	XPAmount    int           `json:"xp_amount" db:"xp_amount"`
	EventTS     time.Time     `json:"event_ts" db:"event_ts"`
}

// XPRepository defines the interface for XP event data access
type XPRepository interface {
	// AddXP creates a new XP event and updates the pointling's XP
	AddXP(event *XPEvent) error

	// GetEventsByPointling retrieves recent XP events for a pointling
	GetEventsByPointling(pointlingID int64, limit int) ([]*XPEvent, error)

	// GetDailyXPBySource gets total XP gained from a source today
	GetDailyXPBySource(pointlingID int64, source XPEventSource) (int, error)

	// InTransaction executes the given function within a database transaction
	InTransaction(fn func(XPRepository) error) error
}

// ValidateSource checks if the XP source is valid
func (s XPEventSource) ValidateSource() bool {
	switch s {
	case XPSourceReceipt, XPSourcePlay, XPSourceDaily:
		return true
	default:
		return false
	}
}

// GetMaxDailyXP returns the maximum XP allowed per day for this source
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

// GetXPPerAction returns the base XP earned per action of this type
func (s XPEventSource) GetXPPerAction() int {
	switch s {
	case XPSourceReceipt:
		return 10 // 10 XP per receipt
	case XPSourcePlay:
		return 20 // 20 XP per game completion
	case XPSourceDaily:
		return 10 // 10 XP for daily login
	default:
		return 0
	}
}
