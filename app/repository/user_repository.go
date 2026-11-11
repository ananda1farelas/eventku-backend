package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventku/app/model"
)

// ErrNotFound dibalikin repo kalau query SELECT gak nemu baris apapun.
// Service yang nentuin ini mau diterjemahin jadi error HTTP apa
// (biasanya 404, tapi di login/register sengaja ditangani beda, lihat
// service/auth_service.go).
var ErrNotFound = errors.New("data tidak ditemukan")

// ErrDuplicateEmail dibalikin repo kalau INSERT users kena unique
// constraint di kolom email (user daftar pakai email yang sudah terpakai).
var ErrDuplicateEmail = errors.New("email sudah terdaftar")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create menyimpan user baru. roleID diambil dari repo.FindRoleByName(ctx, model.RoleUser)
// di layer service/rules, bukan di-hardcode di sini, supaya repository ini
// tetap "dumb" dan gampang ditest.
func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (full_name, email, password_hash, role_id, is_active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, u.FullName, u.Email, u.PasswordHash, u.RoleID).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// FindByEmail dipakai saat login. JOIN ke roles supaya RoleName langsung
// terisi dan service gak perlu query kedua buat ambil nama role.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT u.id, u.full_name, u.email, u.password_hash, u.role_id,
		       u.is_active, u.created_at, u.updated_at, r.name
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.email = $1`

	return r.scanUser(r.db.QueryRow(ctx, query, email))
}

// FindByID dipakai di middleware JWT (muat ulang user tiap request) dan
// endpoint /auth/me.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT u.id, u.full_name, u.email, u.password_hash, u.role_id,
		       u.is_active, u.created_at, u.updated_at, r.name
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1`

	return r.scanUser(r.db.QueryRow(ctx, query, id))
}

// FindRoleIDByName dipakai service/rules pas register, buat nentuin role_id
// default (selalu model.RoleUser — gak ada jalur publik buat daftar jadi admin).
func (r *UserRepository) FindRoleIDByName(ctx context.Context, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("find role by name: %w", err)
	}
	return id, nil
}

func (r *UserRepository) scanUser(row pgx.Row) (*model.User, error) {
	u := &model.User{}
	err := row.Scan(
		&u.ID, &u.FullName, &u.Email, &u.PasswordHash, &u.RoleID,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.RoleName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}
