package models

import "time"

type CartEvent struct {
	EventType string    `json:"event_type"` // "item_added", "item_updated", "item_removed"
	UserSSOID int       `json:"user_sso_id"`
	ItemID    string    `json:"item_id"` // Используем UUID
	SneakerID int       `json:"sneaker_id,omitempty"`
	Quantity  int       `json:"quantity,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
