package service

import (
	"errors"
	"strings"

	"eventku/app/model"
)

var (
	ErrOrganizationNameTooShort = errors.New("nama organisasi minimal 3 karakter")
	ErrInvalidMemberPosition    = errors.New("posisi anggota harus owner atau staff")
	ErrNotOrganizationMember    = errors.New("anda bukan anggota organisasi ini")
	ErrNotOrganizationOwner     = errors.New("hanya owner organisasi yang bisa melakukan aksi ini")
)

// ValidateOrganizationName dipakai sebelum Create maupun Update.
func ValidateOrganizationName(name string) error {
	if len(strings.TrimSpace(name)) < 3 {
		return ErrOrganizationNameTooShort
	}
	return nil
}

// ValidateMemberPosition cuma ngizinin dua nilai pasti: owner atau staff.
// Dicek ulang di sini (selain tag validate:"oneof=owner staff" di
// AddMemberRequest) biar rule ini tetap jalan kalau suatu saat validator
// tag dilepas/diganti, konsisten sama pola di auth_rules.go.
func ValidateMemberPosition(position string) error {
	switch position {
	case model.OrgPositionOwner, model.OrgPositionStaff:
		return nil
	default:
		return ErrInvalidMemberPosition
	}
}
