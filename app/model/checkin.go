package model

import (
	"time"

	"github.com/google/uuid"
)

type Checkin struct {
	ID          uuid.UUID `json:"id"`
	TicketID    uuid.UUID `json:"ticket_id"`
	CheckedInAt time.Time `json:"checked_in_at"`
	CheckedInBy uuid.UUID `json:"checked_in_by"`
}

type CheckinRequest struct {
	QRCode string `json:"qr_code" validate:"required"`
}

type CheckinTarget struct {
	TicketID           uuid.UUID
	RegistrationID     uuid.UUID
	RegistrationStatus string
	EventID            uuid.UUID
	OrganizationID     uuid.UUID  // organisasi pemilik event
	CheckedInAt        *time.Time // nil kalau tiket belum pernah check-in
}
