package model

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID             uuid.UUID `json:"id"`
	RegistrationID uuid.UUID `json:"registration_id"`
	QRCode         string    `json:"qr_code"`
	IssuedAt       time.Time `json:"issued_at"`
}

type TicketDetail struct {
	ID             uuid.UUID `json:"id"`
	RegistrationID uuid.UUID `json:"registration_id"`
	EventID        uuid.UUID `json:"event_id"`
	QRCode         string    `json:"qr_code"`
	IssuedAt       time.Time `json:"issued_at"`
	CheckedIn      bool      `json:"checked_in"`
}
