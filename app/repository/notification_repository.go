package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"eventku/app/model"
)

type NotificationRepository struct {
	db DBTX
}

func NewNotificationRepository(db DBTX) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create menyimpan notifikasi. Mengembalikan created=false (tanpa error)
// kalau notifikasi yang sama (user, event, type) sudah pernah dibuat.
func (r *NotificationRepository) Create(ctx context.Context, n *model.Notification) (bool, error) {
	query := `
		INSERT INTO notifications (user_id, event_id, type, title, message)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query, n.UserID, n.EventID, n.Type, n.Title, n.Message).
		Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("create notification: %w", err)
	}
	return true, nil
}

// ListByUser mengembalikan satu halaman notifikasi milik user (terbaru dulu).
func (r *NotificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit, offset int) ([]model.Notification, int64, error) {
	where := `user_id = $1`
	if unreadOnly {
		where += ` AND read_at IS NULL`
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE `+where, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, event_id, type, title, message, read_at, created_at
		FROM notifications WHERE `+where+`
		ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	out := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.EventID, &n.Type, &n.Title, &n.Message, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	return out, total, nil
}

// MarkRead menandai notifikasi milik userID sebagai sudah dibaca.
// Memanggil ulang aman (waktu baca pertama dipertahankan).
func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE notifications SET read_at = COALESCE(read_at, now()) WHERE id = $1 AND user_id = $2`,
		id, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}