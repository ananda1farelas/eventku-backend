package model

import (
	"time"

	"github.com/google/uuid"
)

// ParticipantRow satu baris rekap peserta sebuah event.
type ParticipantRow struct {
	RegistrationID uuid.UUID  `json:"registration_id"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email"`
	Status         string     `json:"status"` // registered / waitlist / cancelled
	RegisteredAt   time.Time  `json:"registered_at"`
	Attended       bool       `json:"attended"`
	CheckedInAt    *time.Time `json:"checked_in_at,omitempty"`
}

// Recipient penerima notifikasi pengingat.
type Recipient struct {
	UserID   uuid.UUID
	FullName string
	Email    string
}