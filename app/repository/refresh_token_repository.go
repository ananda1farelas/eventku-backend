package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventku/app/model"
)

// RefreshToken merepresentasikan baris tabel refresh_tokens.
// TokenHash adalah SHA-256 dari refresh token mentah — service yang hash,
// repo cuma nyimpen/nyari hash-nya, gak pernah pegang token mentah.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Store nyimpen refresh token baru pas login/refresh berhasil.
func (r *RefreshTokenRepository) Store(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

// FindByHash dipakai pas endpoint /auth/refresh: cari token yang dikirim
// client (sudah di-hash dulu oleh service), buat divalidasi masih berlaku
// atau belum (expired/revoked dicek di service/rules, bukan di sini).
func (r *RefreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*model.User, *RefreshToken, error) {
	rt := &RefreshToken{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("find refresh token: %w", err)
	}

	userRepo := &UserRepository{db: r.db}
	user, err := userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, nil, err
	}

	return user, rt, nil
}

// Revoke nandain satu refresh token spesifik sudah tidak berlaku (dipakai
// saat refresh berhasil: token lama direvoke, token baru di-Store —
// rotation, bukan dipakai ulang).
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser dipakai kalau nanti ada fitur "logout semua device" /
// ganti password paksa logout. Belum dipakai endpoint manapun sekarang,
// disiapin aja supaya service tinggal panggil.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}
