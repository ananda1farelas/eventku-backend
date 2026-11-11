package model

import (
	"time"

	"github.com/google/uuid"
)

// User merepresentasikan baris tabel users.
// PasswordHash sengaja json:"-" supaya gak pernah ke-marshal ke response,
// sekecil apapun kemungkinannya.
type User struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	RoleID       uuid.UUID `json:"role_id"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// RoleName diisi repo lewat JOIN ke tabel roles saat dibutuhkan
	// (login, me, JWT claims). Bukan kolom asli tabel users, jadi jangan
	// di-assume selalu terisi kalau query-nya cuma SELECT * FROM users.
	RoleName string `json:"role,omitempty"`
}
