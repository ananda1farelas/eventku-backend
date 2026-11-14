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

func (r *RegistrationRepository) Create(ctx context.Context, reg *model.Registration) error {
	query := `
		INSERT INTO registrations (event_id, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, queued_at, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, reg.EventID, reg.UserID, reg.Status).
		Scan(&reg.ID, &reg.QueuedAt, &reg.CreatedAt, &reg.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err, "registrations_event_user_key") {
			return ErrRegistrationExists
		}
		return fmt.Errorf("create registration: %w", err)
	}
	return nil
}

func (r *RegistrationRepository) Reactivate(ctx context.Context, reg *model.Registration, status string) error {
	query := `
		UPDATE registrations
		SET status = $2, queued_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING queued_at, updated_at`

	err := r.db.QueryRow(ctx, query, reg.ID, status).Scan(&reg.QueuedAt, &reg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("reactivate registration: %w", err)
	}
	reg.Status = status
	return nil
}

func (r *RegistrationRepository) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE registrations SET status = $2, updated_at = now() WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("set registration status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// FindOldestWaitlistForUpdate mengambil (dan mengunci) antrean waitlist
// paling depan di sebuah event: queued_at paling lama, id sebagai
// pemutus kalau waktunya persis sama. Balikin ErrNotFound kalau waitlist
// kosong, artinya tidak ada yang perlu dipromosikan.
func (r *RegistrationRepository) FindOldestWaitlistForUpdate(ctx context.Context, eventID uuid.UUID) (*model.Registration, error) {
	query := `SELECT ` + registrationColumns + `
		FROM registrations
		WHERE event_id = $1 AND status = $2
		ORDER BY queued_at ASC, id ASC
		LIMIT 1
		FOR UPDATE`

	return scanRegistration(r.db.QueryRow(ctx, query, eventID, model.RegistrationStatusWaitlist))
}
