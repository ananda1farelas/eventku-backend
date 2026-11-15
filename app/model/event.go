package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventStatusDraft         = "draft"
	EventStatusPendingReview = "pending_review"
	EventStatusOpen          = "open"
	EventStatusRejected      = "rejected"
	EventStatusClosed        = "closed"
	EventStatusDone          = "done"
)

// Event merepresentasikan satu baris tabel events (yang belum di-soft-delete).
type Event struct {
	ID                   uuid.UUID `json:"id"`
	EventNo              int       `json:"event_no"`
	OrganizationID       uuid.UUID `json:"organization_id"`
	CreatedBy            uuid.UUID `json:"created_by"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	Category             string    `json:"category"`
	Location             string    `json:"location"`
	StartAt              time.Time `json:"start_at"`
	EndAt                time.Time `json:"end_at"`
	Quota                int       `json:"quota"`
	Price                float64   `json:"price"` // 0 = gratis
	RegistrationDeadline time.Time `json:"registration_deadline"`
	Status               string    `json:"status"`
	RejectionReason      *string   `json:"rejection_reason,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CreateEventRequest untuk POST /events.
type CreateEventRequest struct {
	OrganizationID       uuid.UUID `json:"organization_id" validate:"required"`
	Title                string    `json:"title" validate:"required,min=3,max=200"`
	Description          string    `json:"description" validate:"max=5000"`
	Category             string    `json:"category" validate:"max=50"`
	Location             string    `json:"location" validate:"required,max=200"`
	StartAt              time.Time `json:"start_at" validate:"required"`
	EndAt                time.Time `json:"end_at" validate:"required"`
	Quota                int       `json:"quota" validate:"required,gt=0"`
	Price                float64   `json:"price" validate:"gte=0"`
	RegistrationDeadline time.Time `json:"registration_deadline" validate:"required"`
}

// UpdateEventRequest untuk PUT /events/{id} (kirim data lengkap).
// Organisasi pemilik event tidak bisa dipindah.
type UpdateEventRequest struct {
	Title                string    `json:"title" validate:"required,min=3,max=200"`
	Description          string    `json:"description" validate:"max=5000"`
	Category             string    `json:"category" validate:"max=50"`
	Location             string    `json:"location" validate:"required,max=200"`
	StartAt              time.Time `json:"start_at" validate:"required"`
	EndAt                time.Time `json:"end_at" validate:"required"`
	Quota                int       `json:"quota" validate:"required,gt=0"`
	Price                float64   `json:"price" validate:"gte=0"`
	RegistrationDeadline time.Time `json:"registration_deadline" validate:"required"`
}

// RejectEventRequest untuk PATCH /events/{id}/reject.
type RejectEventRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

// EventFilter dipakai repository untuk query list. Statuses nil artinya
// semua status (hanya diisi nil untuk admin/panitia, diputuskan di service).
type EventFilter struct {
	Page           int
	Limit          int
	Search         string
	Category       string
	Statuses       []string
	OrganizationID *uuid.UUID
}