package models

import "time"

type CartItem struct {
	ID           string    `json:"id"`
	UserSSOID    int       `json:"user_sso_id"`
	SneakerID    int       `json:"sneaker_id"`
	Quantity     int       `json:"quantity"`
	AddedAt      time.Time `json:"added_at"`
	Synchronized bool      `json:"synchronized"`
}

type Cart struct {
	UserSSOID int        `json:"user_sso_id"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
}
