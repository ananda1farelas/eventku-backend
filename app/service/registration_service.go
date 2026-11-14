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

func (s *RegistrationService) cancel(ctx context.Context, userID, registrationID uuid.UUID) (*model.Registration, error) {
	var result *model.Registration

	err := s.tx.WithTx(ctx, func(db repository.DBTX) error {
		regs := s.regs.WithTx(db)

		peek, err := regs.FindByID(ctx, registrationID)
		if err != nil {
			return err
		}
		if err := CanCancel(peek, userID); err != nil {
			return err
		}

		if _, err := regs.LockEventInfo(ctx, peek.EventID); err != nil {
			return err
		}

		reg, err := regs.FindByIDForUpdate(ctx, registrationID)
		if err != nil {
			return err
		}
		if err := CanCancel(reg, userID); err != nil {
			return err
		}

		previousStatus := reg.Status
		if err := regs.SetStatus(ctx, reg.ID, model.RegistrationStatusCancelled); err != nil {
			return err
		}
		reg.Status = model.RegistrationStatusCancelled

		if ShouldPromoteWaitlist(previousStatus) {
			next, err := regs.FindOldestWaitlistForUpdate(ctx, reg.EventID)
			switch {
			case err == nil:
				if err := regs.SetStatus(ctx, next.ID, model.RegistrationStatusRegistered); err != nil {
					return err
				}
				// TODO(ticket): terbitkan tiket QR untuk next
				// (ditambahkan di branch feat/ticket-qr).
			case errors.Is(err, repository.ErrNotFound):
				// Waitlist kosong, tidak ada yang perlu dipromosikan.
			default:
				return err
			}
		}

		result = reg
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// listMine untuk GET /me/registrations.
func (s *RegistrationService) listMine(ctx context.Context, userID uuid.UUID) ([]model.Registration, error) {
	return s.regs.ListByUser(ctx, userID)
}
