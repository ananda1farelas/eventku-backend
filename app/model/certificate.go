package model

import (
	"time"

	"github.com/google/uuid"
)

type Certificate struct {
	ID             uuid.UUID `json:"id"`
	RegistrationID uuid.UUID `json:"registration_id"`
	CertificateNo  string    `json:"certificate_no"`
	FileURL        *string   `json:"file_url,omitempty"`
	IssuedAt       time.Time `json:"issued_at"`
}

// CertificateDetail untuk GET /me/certificates (ditambah info event).
type CertificateDetail struct {
	ID             uuid.UUID `json:"id"`
	RegistrationID uuid.UUID `json:"registration_id"`
	EventID        uuid.UUID `json:"event_id"`
	EventTitle     string    `json:"event_title"`
	CertificateNo  string    `json:"certificate_no"`
	FileURL        *string   `json:"file_url,omitempty"`
	IssuedAt       time.Time `json:"issued_at"`
}

// CertificateCandidate adalah satu registrasi berstatus registered
// beserta fakta yang dibutuhkan rule kelayakan sertifikat.
type CertificateCandidate struct {
	RegistrationID uuid.UUID
	Attended       bool // punya check-in
	HasCertificate bool
}

// IssueCertificatesResult ringkasan hasil POST /events/{id}/certificates/issue.
type IssueCertificatesResult struct {
	Issued       int           `json:"issued"`
	Skipped      int           `json:"skipped"`
	Certificates []Certificate `json:"certificates"`
}