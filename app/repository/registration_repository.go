package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"eventku/app/model"
)

var ErrRegistrationExists = errors.New("registrasi untuk event ini sudah ada")

const registrationColumns = `id, event_id, user_id, status, queued_at, created_at, updated_at`

type RegistrationRepository struct {
	db DBTX
}

func NewRegistrationRepository(db DBTX) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) WithTx(tx DBTX) *RegistrationRepository {
	return &RegistrationRepository{db: tx}
}

func (r *RegistrationRepository) LockEventInfo(ctx context.Context, eventID uuid.UUID) (*model.EventRegistrationInfo, error) {
	lockQuery := `
		SELECT status, quota, registration_deadline
		FROM events
		WHERE id = $1
		FOR UPDATE`

	var info model.EventRegistrationInfo
	err := r.db.QueryRow(ctx, lockQuery, eventID).
		Scan(&info.Status, &info.Quota, &info.RegistrationDeadline)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lock event: %w", err)
	}

	countQuery := `SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = $2`
	err = r.db.QueryRow(ctx, countQuery, eventID, model.RegistrationStatusRegistered).
		Scan(&info.RegisteredCount)
	if err != nil {
		return nil, fmt.Errorf("count registered: %w", err)
	}
	return &info, nil
}

func (r *RegistrationRepository) FindByEventAndUser(ctx context.Context, eventID, userID uuid.UUID) (*model.Registration, error) {
	query := `SELECT ` + registrationColumns + `
		FROM registrations
		WHERE event_id = $1 AND user_id = $2`

	return scanRegistration(r.db.QueryRow(ctx, query, eventID, userID))
}

func (r *RegistrationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Registration, error) {
	query := `SELECT ` + registrationColumns + ` FROM registrations WHERE id = $1`

	return scanRegistration(r.db.QueryRow(ctx, query, id))
}

func (r *RegistrationRepository) FindByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Registration, error) {
	query := `SELECT ` + registrationColumns + ` FROM registrations WHERE id = $1 FOR UPDATE`

	return scanRegistration(r.db.QueryRow(ctx, query, id))
}

func (r *RegistrationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Registration, error) {
	query := `SELECT ` + registrationColumns + `
		FROM registrations
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}
	defer rows.Close()

	regs := []model.Registration{}
	for rows.Next() {
		reg, err := scanRegistration(rows)
		if err != nil {
			return nil, err
		}
		regs = append(regs, *reg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}
	return regs, nil
}

func scanRegistration(row interface{ Scan(dest ...any) error }) (*model.Registration, error) {
	var reg model.Registration
	err := row.Scan(&reg.ID, &reg.EventID, &reg.UserID, &reg.Status,
		&reg.QueuedAt, &reg.CreatedAt, &reg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan registration: %w", err)
	}
	return &reg, nil
}
