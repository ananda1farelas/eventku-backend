package model
package model

import (
	"time"

	"github.com/google/uuid"
)

// Organization adalah organisasi mahasiswa pemilik event (BEM, HMJ, UKM, dst).
type Organization struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Posisi anggota organisasi. "owner" otomatis diisi ID pembuat organisasi
// saat Create, "staff" ditambahkan belakangan lewat endpoint add member.
// Ini BUKAN role sistem (model.RoleAdmin/RoleUser) — dua konsep yang
// terpisah: role sistem nentuin hak akses global, posisi organisasi
// nentuin hak akses di dalam satu organisasi tertentu.
const (
	OrgPositionOwner = "owner"
	OrgPositionStaff = "staff"
)

type OrganizationMember struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Position       string    `json:"position"`
	CreatedAt      time.Time `json:"created_at"`

	// Diisi repo lewat JOIN ke users saat dibutuhkan (list anggota),
	// bukan kolom asli tabel organization_members.
	UserFullName string `json:"user_full_name,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
}

// === Request ===

type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=150"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=150"`
	Description string `json:"description" validate:"max=1000"`
}

type AddMemberRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Position string    `json:"position" validate:"required,oneof=owner staff"`
}