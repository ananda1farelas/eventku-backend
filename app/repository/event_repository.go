package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"eventku/app/model"
)

// ErrEventConflict dibalikin kalau UPDATE gagal karena status event sudah
// berubah sejak dibaca (dua orang mengubah event yang sama bersamaan).
var ErrEventConflict = errors.New("status event sudah berubah, muat ulang lalu coba lagi")

const eventColumns = `id, event_no, organization_id, created_by, title, description, category,
	location, start_at, end_at, quota, price, registration_deadline, status,
	rejection_reason, created_at, updated_at`

type EventRepository struct {
	db DBTX
}

func NewEventRepository(db DBTX) *EventRepository {
	return &EventRepository{db: db}
}

// WithTx mengembalikan salinan repository yang query-nya jalan di tx.
func (r *EventRepository) WithTx(tx DBTX) *EventRepository {
	return &EventRepository{db: tx}
}

func scanEvent(row interface{ Scan(dest ...any) error }) (*model.Event, error) {
	var e model.Event
	err := row.Scan(&e.ID, &e.EventNo, &e.OrganizationID, &e.CreatedBy, &e.Title,
		&e.Description, &e.Category, &e.Location, &e.StartAt, &e.EndAt, &e.Quota,
		&e.Price, &e.RegistrationDeadline, &e.Status, &e.RejectionReason,
		&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan event: %w", err)
	}
	return &e, nil
}

func (r *EventRepository) Create(ctx context.Context, e *model.Event) error {
	query := `
		INSERT INTO events (organization_id, created_by, title, description, category,
			location, start_at, end_at, quota, price, registration_deadline, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, event_no, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, e.OrganizationID, e.CreatedBy, e.Title, e.Description,
		e.Category, e.Location, e.StartAt, e.EndAt, e.Quota, e.Price,
		e.RegistrationDeadline, e.Status).
		Scan(&e.ID, &e.EventNo, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

func (r *EventRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	query := `SELECT ` + eventColumns + ` FROM events WHERE id = $1 AND deleted_at IS NULL`
	return scanEvent(r.db.QueryRow(ctx, query, id))
}

// LockByID sama seperti FindByID tapi mengunci baris (FOR UPDATE).
// Hanya berguna di dalam transaction.
func (r *EventRepository) LockByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	query := `SELECT ` + eventColumns + ` FROM events WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`
	return scanEvent(r.db.QueryRow(ctx, query, id))
}

// List mengembalikan satu halaman event beserta total datanya.
func (r *EventRepository) List(ctx context.Context, f model.EventFilter) ([]model.Event, int64, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	if f.Statuses != nil {
		add("status = ANY($%d)", f.Statuses)
	}
	if f.OrganizationID != nil {
		add("organization_id = $%d", *f.OrganizationID)
	}
	if f.Category != "" {
		add("category = $%d", f.Category)
	}
	if f.Search != "" {
		add("title ILIKE $%d", "%"+f.Search+"%")
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	args = append(args, f.Limit, (f.Page-1)*f.Limit)
	query := fmt.Sprintf(`SELECT %s FROM events WHERE %s ORDER BY start_at ASC, id ASC LIMIT $%d OFFSET $%d`,
		eventColumns, whereSQL, len(args)-1, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := []model.Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, *e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	return events, total, nil
}

// Update menyimpan perubahan isi event. WHERE status = status-lama menjaga
// supaya tidak menimpa event yang statusnya baru saja diubah orang lain.
func (r *EventRepository) Update(ctx context.Context, e *model.Event, expectedStatus string) error {
	query := `
		UPDATE events
		SET title = $3, description = $4, category = $5, location = $6, start_at = $7,
		    end_at = $8, quota = $9, price = $10, registration_deadline = $11, updated_at = now()
		WHERE id = $1 AND status = $2 AND deleted_at IS NULL
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query, e.ID, expectedStatus, e.Title, e.Description, e.Category,
		e.Location, e.StartAt, e.EndAt, e.Quota, e.Price, e.RegistrationDeadline).
		Scan(&e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventConflict
		}
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

// UpdateStatus mengubah status dari `from` ke `to`. reason hanya terisi untuk
// reject; nil menghapus alasan lama.
func (r *EventRepository) UpdateStatus(ctx context.Context, id uuid.UUID, from, to string, reason *string) error {
	query := `
		UPDATE events
		SET status = $3, rejection_reason = $4, updated_at = now()
		WHERE id = $1 AND status = $2 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, id, from, to, reason)
	if err != nil {
		return fmt.Errorf("update event status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrEventConflict
	}
	return nil
}

// SoftDelete menandai event terhapus (baris tetap ada untuk riwayat).
func (r *EventRepository) SoftDelete(ctx context.Context, id uuid.UUID, expectedStatus string) error {
	query := `
		UPDATE events SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND status = $2 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, id, expectedStatus)
	if err != nil {
		return fmt.Errorf("soft delete event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrEventConflict
	}
	return nil
}

// ListOpenStartingBetween mengambil event open yang mulai di [from, to).
// Dipakai pengingat H-1.
func (r *EventRepository) ListOpenStartingBetween(ctx context.Context, from, to time.Time) ([]model.Event, error) {
	query := `SELECT ` + eventColumns + `
		FROM events
		WHERE status = $1 AND deleted_at IS NULL AND start_at >= $2 AND start_at < $3
		ORDER BY start_at ASC`

	rows, err := r.db.Query(ctx, query, model.EventStatusOpen, from, to)
	if err != nil {
		return nil, fmt.Errorf("list events for reminder: %w", err)
	}
	defer rows.Close()

	events := []model.Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list events for reminder: %w", err)
	}
	return events, nil
}