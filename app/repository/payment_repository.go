package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"eventku/app/model"
)

var (
	ErrPaymentExists   = errors.New("registrasi ini sudah punya pembayaran yang pending atau lunas")
	ErrPaymentConflict = errors.New("status pembayaran sudah berubah, muat ulang lalu coba lagi")
)

type PaymentRepository struct {
	db DBTX
}

func NewPaymentRepository(db DBTX) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) WithTx(tx DBTX) *PaymentRepository {
	return &PaymentRepository{db: tx}
}

// FindPayable membaca registrasi beserta harga event-nya (read-only).
func (r *PaymentRepository) FindPayable(ctx context.Context, registrationID uuid.UUID) (*model.PayableRegistration, error) {
	query := `
		SELECT r.id, r.user_id, r.status, e.id, e.title, e.price
		FROM registrations r
		JOIN events e ON e.id = r.event_id
		WHERE r.id = $1`

	var p model.PayableRegistration
	err := r.db.QueryRow(ctx, query, registrationID).
		Scan(&p.RegistrationID, &p.UserID, &p.Status, &p.EventID, &p.EventTitle, &p.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find payable registration: %w", err)
	}
	return &p, nil
}

func (r *PaymentRepository) Create(ctx context.Context, p *model.Payment) error {
	query := `
		INSERT INTO payments (registration_id, amount, method, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query, p.RegistrationID, p.Amount, p.Method, p.Status).
		Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		if isUniqueViolation(err, "payments_one_active_per_registration") {
			return ErrPaymentExists
		}
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

const paymentDetailSelect = `
	SELECT p.id, p.registration_id, p.amount, p.method, p.status, p.paid_at, p.created_at,
	       r.user_id, e.id, e.title
	FROM payments p
	JOIN registrations r ON r.id = p.registration_id
	JOIN events e ON e.id = r.event_id`

func scanPaymentDetail(row interface{ Scan(dest ...any) error }) (*model.PaymentDetail, error) {
	var d model.PaymentDetail
	err := row.Scan(&d.ID, &d.RegistrationID, &d.Amount, &d.Method, &d.Status, &d.PaidAt,
		&d.CreatedAt, &d.UserID, &d.EventID, &d.EventTitle)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan payment: %w", err)
	}
	return &d, nil
}

func (r *PaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.PaymentDetail, error) {
	return scanPaymentDetail(r.db.QueryRow(ctx, paymentDetailSelect+` WHERE p.id = $1`, id))
}

func (r *PaymentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.PaymentDetail, error) {
	rows, err := r.db.Query(ctx, paymentDetailSelect+` WHERE r.user_id = $1 ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	out := []model.PaymentDetail{}
	for rows.Next() {
		d, err := scanPaymentDetail(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	return out, nil
}

// UpdateStatus mengubah status dari `from` ke `to`. paidAt hanya diisi untuk paid.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, from, to string, paidAt *time.Time) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE payments SET status = $3, paid_at = $4 WHERE id = $1 AND status = $2`,
		id, from, to, paidAt)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPaymentConflict
	}
	return nil
}