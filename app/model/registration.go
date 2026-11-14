package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	RegistrationStatusRegistered = "registered"
	RegistrationStatusWaitlist   = "waitlist"
	RegistrationStatusCancelled  = "cancelled"
)

type Registration struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"`
	QueuedAt  time.Time `json:"queued_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegistrationResponse struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	Status    string    `json:"status"`
	QueuedAt  time.Time `json:"queued_at"`
	CreatedAt time.Time `json:"created_at"`
}

func NewRegistrationResponse(r *Registration) *RegistrationResponse {
	return &RegistrationResponse{
		ID:        r.ID,
		EventID:   r.EventID,
		Status:    r.Status,
		QueuedAt:  r.QueuedAt,
		CreatedAt: r.CreatedAt,
	}
}