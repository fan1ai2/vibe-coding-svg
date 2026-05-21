package model

import "time"

type DailyQuota struct {
	ID     string    `json:"id" db:"id"`
	UserID string    `json:"user_id" db:"user_id"`
	Date   time.Time `json:"date" db:"date"`
	Count  int       `json:"count" db:"count"`
}
