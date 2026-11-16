package service

import (
	"context"

	"eventku/app/repository"
)

type RBACService struct {
	userRepo *repository.UserRepository
}

func NewRBACService(userRepo *repository.UserRepository) *RBACService {
	return &RBACService{userRepo: userRepo}
}

// Can mengecek apakah role tertentu boleh melakukan permission tertentu.
//
// FAIL CLOSED secara sengaja: kalau roleName/permissionName kosong, atau
// query ke DB error (koneksi putus, dll), method ini balikin false — BUKAN
// true. Prinsipnya: kalau sistem lagi ragu/error, tolak dulu, jangan sampai
// bug infrastruktur malah meloloskan semua orang lewat gate RBAC.
func (s *RBACService) Can(ctx context.Context, roleName, permissionName string) (bool, error) {
	if roleName == "" || permissionName == "" {
		return false, nil
	}

	allowed, err := s.userRepo.HasPermission(ctx, roleName, permissionName)
	if err != nil {
		return false, err
	}
	return allowed, nil
}
