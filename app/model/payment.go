package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"

	PaymentMethodTransfer = "transfer"
	PaymentMethodEwallet  = "ewallet"
	PaymentMethodQRIS     = "qris"
)

type Payment struct {
	ID             uuid.UUID  `json:"id"`
	RegistrationID uuid.UUID  `json:"registration_id"`
	Amount         float64    `json:"amount"`
	Method         string     `json:"method"`
	Status         string     `json:"status"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// PaymentDetail = Payment + info pemilik dan event (untuk cek akses dan tampilan).
type PaymentDetail struct {
	Payment
	UserID     uuid.UUID `json:"-"`
	EventID    uuid.UUID `json:"event_id"`
	EventTitle string    `json:"event_title"`
}

// PayableRegistration adalah fakta tentang sebuah registrasi yang dibutuhkan
// rule pembayaran (dibaca read-only dari registrations + events).
type PayableRegistration struct {
	RegistrationID uuid.UUID
	UserID         uuid.UUID
	Status         string
	EventID        uuid.UUID
	EventTitle     string
	Price          float64
}

// CreatePaymentRequest untuk POST /registrations/{id}/payments.
type CreatePaymentRequest struct {
	Method string `json:"method" validate:"required,oneof=transfer ewallet qris"`
}