package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"eventku/app/model"
)

var ErrTicketExists = errors.New("tiket untuk registrasi ini sudah ada")

type TicketRepository struct {
	db DBTX
}

func NewTicketRepository(db DBTX) *TicketRepository {
	return &TicketRepository{db: db}
}

// WithTx mengembalikan salinan repository yang query-nya jalan di tx.
func (r *TicketRepository) WithTx(tx DBTX) *TicketRepository {
	return &TicketRepository{db: tx}
}

func (r *TicketRepository) Create(ctx context.Context, t *model.Ticket) error {
	query := `
		INSERT INTO tickets (registration_id, qr_code)
		VALUES ($1, $2)
		RETURNING id, issued_at`

	err := r.db.QueryRow(ctx, query, t.RegistrationID, t.QRCode).
		Scan(&t.ID, &t.IssuedAt)
	if err != nil {
		if isUniqueViolation(err, "tickets_registration_key") {
			return ErrTicketExists
		}
		return fmt.Errorf("create ticket: %w", err)
	}
	return nil
}

func (r *TicketRepository) FindByRegistrationID(ctx context.Context, registrationID uuid.UUID) (*model.Ticket, error) {
	query := `
		SELECT id, registration_id, qr_code, issued_at
		FROM tickets
		WHERE registration_id = $1`

	return scanTicket(r.db.QueryRow(ctx, query, registrationID))
}

// FindByIDForUser mengambil tiket HANYA kalau milik user itu dan
// registrasinya masih berstatus registered. Dipakai GET /tickets/{id}/qr,
// jadi orang lain tidak bisa melihat QR tiket yang bukan miliknya.
func (r *TicketRepository) FindByIDForUser(ctx context.Context, ticketID, userID uuid.UUID) (*model.Ticket, error) {
	query := `
		SELECT t.id, t.registration_id, t.qr_code, t.issued_at
		FROM tickets t
		JOIN registrations r ON r.id = t.registration_id
		WHERE t.id = $1 AND r.user_id = $2 AND r.status = $3`

	return scanTicket(r.db.QueryRow(ctx, query, ticketID, userID, model.RegistrationStatusRegistered))
}

func (r *TicketRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.TicketDetail, error) {
	query := `
		SELECT t.id, t.registration_id, r.event_id, t.qr_code, t.issued_at,
		       (c.id IS NOT NULL) AS checked_in
		FROM tickets t
		JOIN registrations r ON r.id = t.registration_id
		LEFT JOIN checkins c ON c.ticket_id = t.id
		WHERE r.user_id = $1 AND r.status = $2
		ORDER BY t.issued_at DESC`

	rows, err := r.db.Query(ctx, query, userID, model.RegistrationStatusRegistered)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	tickets := []model.TicketDetail{}
	for rows.Next() {
		var t model.TicketDetail
		if err := rows.Scan(&t.ID, &t.RegistrationID, &t.EventID,
			&t.QRCode, &t.IssuedAt, &t.CheckedIn); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	return tickets, nil
}

// scanTicket menerima pgx.Row (QueryRow) maupun pgx.Rows (Query).
func scanTicket(row interface{ Scan(dest ...any) error }) (*model.Ticket, error) {
	var t model.Ticket
	err := row.Scan(&t.ID, &t.RegistrationID, &t.QRCode, &t.IssuedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan ticket: %w", err)
	}
	return &t, nil
}
