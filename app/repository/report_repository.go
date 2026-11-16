package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"eventku/app/model"
)

// ReportRepository membaca data lintas tabel untuk rekap dan pengingat.
// Hanya SELECT; tidak mengubah data modul lain.
type ReportRepository struct {
	db DBTX
}

func NewReportRepository(db DBTX) *ReportRepository {
	return &ReportRepository{db: db}
}

// ListParticipants rekap peserta satu event. status kosong = semua status.
func (r *ReportRepository) ListParticipants(ctx context.Context, eventID uuid.UUID, status string) ([]model.ParticipantRow, error) {
	query := `
		SELECT r.id, u.full_name, u.email, r.status, r.created_at,
		       (ck.id IS NOT NULL), ck.checked_in_at
		FROM registrations r
		JOIN users u ON u.id = r.user_id
		LEFT JOIN tickets t ON t.registration_id = r.id
		LEFT JOIN checkins ck ON ck.ticket_id = t.id
		WHERE r.event_id = $1 AND ($2 = '' OR r.status = $2)
		ORDER BY r.created_at ASC, r.id ASC`

	rows, err := r.db.Query(ctx, query, eventID, status)
	if err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}
	defer rows.Close()

	out := []model.ParticipantRow{}
	for rows.Next() {
		var p model.ParticipantRow
		if err := rows.Scan(&p.RegistrationID, &p.FullName, &p.Email, &p.Status,
			&p.RegisteredAt, &p.Attended, &p.CheckedInAt); err != nil {
			return nil, fmt.Errorf("scan participant: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}
	return out, nil
}

// ListRegisteredRecipients penerima pengingat: peserta berstatus registered.
func (r *ReportRepository) ListRegisteredRecipients(ctx context.Context, eventID uuid.UUID) ([]model.Recipient, error) {
	query := `
		SELECT u.id, u.full_name, u.email
		FROM registrations r
		JOIN users u ON u.id = r.user_id
		WHERE r.event_id = $1 AND r.status = $2
		ORDER BY r.created_at ASC`

	rows, err := r.db.Query(ctx, query, eventID, model.RegistrationStatusRegistered)
	if err != nil {
		return nil, fmt.Errorf("list recipients: %w", err)
	}
	defer rows.Close()

	out := []model.Recipient{}
	for rows.Next() {
		var rc model.Recipient
		if err := rows.Scan(&rc.UserID, &rc.FullName, &rc.Email); err != nil {
			return nil, fmt.Errorf("scan recipient: %w", err)
		}
		out = append(out, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list recipients: %w", err)
	}
	return out, nil
}