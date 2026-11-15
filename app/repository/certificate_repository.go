package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"eventku/app/model"
)

var ErrCertificateExists = errors.New("sertifikat untuk registrasi ini sudah ada")

type CertificateRepository struct {
	db DBTX
}

func NewCertificateRepository(db DBTX) *CertificateRepository {
	return &CertificateRepository{db: db}
}

func (r *CertificateRepository) WithTx(tx DBTX) *CertificateRepository {
	return &CertificateRepository{db: tx}
}

func (r *CertificateRepository) Create(ctx context.Context, c *model.Certificate) error {
	query := `
		INSERT INTO certificates (registration_id, certificate_no, file_url)
		VALUES ($1, $2, $3)
		RETURNING id, issued_at`

	err := r.db.QueryRow(ctx, query, c.RegistrationID, c.CertificateNo, c.FileURL).
		Scan(&c.ID, &c.IssuedAt)
	if err != nil {
		if isUniqueViolation(err, "certificates_registration_key") {
			return ErrCertificateExists
		}
		return fmt.Errorf("create certificate: %w", err)
	}
	return nil
}

// CountByEvent menghitung sertifikat yang sudah terbit di sebuah event,
// dipakai untuk nomor urut berikutnya. Aman dari balapan karena pemanggil
// mengunci baris event (EventRepository.LockByID) dalam transaction.
func (r *CertificateRepository) CountByEvent(ctx context.Context, eventID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM certificates c
		JOIN registrations r ON r.id = c.registration_id
		WHERE r.event_id = $1`

	var n int
	if err := r.db.QueryRow(ctx, query, eventID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count certificates: %w", err)
	}
	return n, nil
}

// ListCandidates mengambil semua registrasi berstatus registered di event
// itu beserta fakta hadir/sudah punya sertifikat, urut waktu check-in.
func (r *CertificateRepository) ListCandidates(ctx context.Context, eventID uuid.UUID) ([]model.CertificateCandidate, error) {
	query := `
		SELECT r.id, (ck.id IS NOT NULL), (c.id IS NOT NULL)
		FROM registrations r
		LEFT JOIN tickets t ON t.registration_id = r.id
		LEFT JOIN checkins ck ON ck.ticket_id = t.id
		LEFT JOIN certificates c ON c.registration_id = r.id
		WHERE r.event_id = $1 AND r.status = $2
		ORDER BY ck.checked_in_at ASC NULLS LAST, r.id ASC`

	rows, err := r.db.Query(ctx, query, eventID, model.RegistrationStatusRegistered)
	if err != nil {
		return nil, fmt.Errorf("list certificate candidates: %w", err)
	}
	defer rows.Close()

	out := []model.CertificateCandidate{}
	for rows.Next() {
		var c model.CertificateCandidate
		if err := rows.Scan(&c.RegistrationID, &c.Attended, &c.HasCertificate); err != nil {
			return nil, fmt.Errorf("scan certificate candidate: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list certificate candidates: %w", err)
	}
	return out, nil
}

// ListByUser untuk GET /me/certificates.
func (r *CertificateRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.CertificateDetail, error) {
	query := `
		SELECT c.id, c.registration_id, e.id, e.title, c.certificate_no, c.file_url, c.issued_at
		FROM certificates c
		JOIN registrations r ON r.id = c.registration_id
		JOIN events e ON e.id = r.event_id
		WHERE r.user_id = $1
		ORDER BY c.issued_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}
	defer rows.Close()

	out := []model.CertificateDetail{}
	for rows.Next() {
		var d model.CertificateDetail
		if err := rows.Scan(&d.ID, &d.RegistrationID, &d.EventID, &d.EventTitle,
			&d.CertificateNo, &d.FileURL, &d.IssuedAt); err != nil {
			return nil, fmt.Errorf("scan certificate: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}
	return out, nil
}