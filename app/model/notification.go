package model

import (
	"time"

	"github.com/google/uuid"
)

const NotificationTypeReminderH1 = "reminder_h1"

type Notification struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"-"`
	EventID   *uuid.UUID `json:"event_id,omitempty"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}