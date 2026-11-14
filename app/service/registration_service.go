package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
)

type RegistrationService struct {
	regs *repository.RegistrationRepository
	tx   *repository.TxManager
}

func NewRegistrationService(regs *repository.RegistrationRepository, tx *repository.TxManager) *RegistrationService {
	return &RegistrationService{regs: regs, tx: tx}
}

func (s *RegistrationService) register(ctx context.Context, userID, eventID uuid.UUID) (*model.Registration, error) {
	var result *model.Registration

	err := s.tx.WithTx(ctx, func(db repository.DBTX) error {
		regs := s.regs.WithTx(db)

		info, err := regs.LockEventInfo(ctx, eventID)
		if err != nil {
			return err // ErrNotFound kalau event tidak ada
		}

		existing, err := regs.FindByEventAndUser(ctx, eventID, userID)
		if err != nil {
			if !errors.Is(err, repository.ErrNotFound) {
				return err
			}
			existing = nil // belum pernah daftar
		}

		if err := CanRegister(*info, existing, time.Now()); err != nil {
			return err
		}

		status := DecideStatus(info.Quota, info.RegisteredCount)

		// Pernah batal: pakai baris lama (UNIQUE event_id + user_id).
		if existing != nil {
			if err := regs.Reactivate(ctx, existing, status); err != nil {
				return err
			}
			result = existing
			return nil
		}

		reg := &model.Registration{EventID: eventID, UserID: userID, Status: status}
		if err := regs.Create(ctx, reg); err != nil {
			return err
		}
		result = reg
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
