package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
)

// OrganizationService isi use-case organisasi: create, list, update,
// add member, list member. Keanggotaan (IsMember/IsOwner) dicek di sini
// SEBELUM repo dipanggil buat tulis, bukan dicek belakangan — kalau gagal
// cek, request ditolak lebih awal.
type OrganizationService struct {
	orgRepo *repository.OrganizationRepository
}

func NewOrganizationService(orgRepo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{orgRepo: orgRepo}
}

// Create: siapapun user yang login boleh bikin organisasi baru (lihat
// permission "organizations.create" di seed RBAC), dan otomatis jadi
// owner-nya (ditangani OrganizationRepository.Create dalam satu transaksi).
func (s *OrganizationService) Create(ctx context.Context, ownerID uuid.UUID, name, description string) (*model.Organization, error) {
	if err := ValidateOrganizationName(name); err != nil {
		return nil, err
	}

	org := &model.Organization{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}
	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}
	return org, nil
}

// List dipakai GET /organizations, publik buat siapapun yang login
// (gak ada pengecekan keanggotaan — lihat-lihat daftar organisasi gak
// perlu jadi anggota).
func (s *OrganizationService) List(ctx context.Context, limit, offset int) ([]model.Organization, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.orgRepo.List(ctx, limit, offset)
}

// Update cuma boleh dilakukan anggota organisasi itu sendiri (owner
// maupun staff) — konsisten dengan prinsip "panitia bisa kelola
// organisasinya sendiri" dari dokumen pembagian tugas.
func (s *OrganizationService) Update(ctx context.Context, requesterID, orgID uuid.UUID, name, description string) error {
	if err := ValidateOrganizationName(name); err != nil {
		return err
	}

	isMember, err := s.orgRepo.IsMember(ctx, orgID, requesterID)
	if err != nil {
		return fmt.Errorf("cek keanggotaan: %w", err)
	}
	if !isMember {
		return ErrNotOrganizationMember
	}

	return s.orgRepo.Update(ctx, orgID, name, description)
}

// AddMember cuma boleh dilakukan OWNER organisasi (bukan staff biasa) —
// biar staff gak bisa sembarangan nambah anggota baru atas nama organisasi.
func (s *OrganizationService) AddMember(ctx context.Context, requesterID, orgID, newUserID uuid.UUID, position string) error {
	if err := ValidateMemberPosition(position); err != nil {
		return err
	}

	isOwner, err := s.orgRepo.IsOwner(ctx, orgID, requesterID)
	if err != nil {
		return fmt.Errorf("cek owner: %w", err)
	}
	if !isOwner {
		return ErrNotOrganizationOwner
	}

	return s.orgRepo.AddMember(ctx, orgID, newUserID, position)
}

// ListMembers cuma boleh dilihat sesama anggota organisasi itu (bukan
// orang luar), supaya daftar anggota + email gak bocor ke publik.
func (s *OrganizationService) ListMembers(ctx context.Context, requesterID, orgID uuid.UUID) ([]model.OrganizationMember, error) {
	isMember, err := s.orgRepo.IsMember(ctx, orgID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("cek keanggotaan: %w", err)
	}
	if !isMember {
		return nil, ErrNotOrganizationMember
	}

	return s.orgRepo.ListMembers(ctx, orgID)
}
