package storage

import "time"

type Notification struct {
	EventID   string    `json:"event_id"`
	Title     string    `json:"title"`
	EventDate time.Time `json:"event_date"`
	UserID    string    `json:"user_id"`
}
