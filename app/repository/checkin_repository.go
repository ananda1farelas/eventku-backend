package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"eventku/app/model"
)

var ErrCheckinExists = errors.New("tiket ini sudah check-in")

type CheckinRepository struct {
	db DBTX
}

func NewCheckinRepository(db DBTX) *CheckinRepository {
	return &CheckinRepository{db: db}
}

func (r *CheckinRepository) FindTargetByQR(ctx context.Context, qrCode string) (*model.CheckinTarget, error) {
	query := `
		SELECT t.id, r.id, r.status, r.event_id, e.organization_id, c.checked_in_at
		FROM tickets t
		JOIN registrations r ON r.id = t.registration_id
		JOIN events e ON e.id = r.event_id
		LEFT JOIN checkins c ON c.ticket_id = t.id
		WHERE t.qr_code = $1`

	var target model.CheckinTarget
	err := r.db.QueryRow(ctx, query, qrCode).Scan(
		&target.TicketID, &target.RegistrationID, &target.RegistrationStatus,
		&target.EventID, &target.OrganizationID, &target.CheckedInAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find checkin target: %w", err)
	}
	return &target, nil
}

func (r *CheckinRepository) Create(ctx context.Context, c *model.Checkin) error {
	query := `
		INSERT INTO checkins (ticket_id, checked_in_by)
		VALUES ($1, $2)
		RETURNING id, checked_in_at`

	err := r.db.QueryRow(ctx, query, c.TicketID, c.CheckedInBy).
		Scan(&c.ID, &c.CheckedInAt)
	if err != nil {
		if isUniqueViolation(err, "checkins_ticket_key") {
			return ErrCheckinExists
		}
		return fmt.Errorf("create checkin: %w", err)
	}
	return nil
}
