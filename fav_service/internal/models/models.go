package models

import "time"

type Favourite struct {
	ID        int       `json:"id" db:"id"`
	UserSSOID int       `json:"user_id" db:"user_id"`
	SneakerID int       `json:"sneaker_id" db:"sneaker_id"`
	AddedAt   time.Time `json:"added_at" db:"added_at"`
}
