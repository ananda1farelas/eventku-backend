package model

import (
	"time"

	"github.com/google/uuid"
)

// Role level sistem. Cuma ada dua: admin dan user (mahasiswa).
// Status "panitia" BUKAN role — itu ditentukan per-organisasi lewat
// organization_members, dicek pakai OrganizationRepository.IsMember.
type Role struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Nama role, dipakai di seluruh kode (middleware, service, rules) biar
// gak ada magic string "admin"/"user" nyebar di banyak file.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)
