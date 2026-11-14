package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
)

type OrganizationMemberChecker interface {
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
}

type CheckinService struct {
	checkins *repository.CheckinRepository
	members  OrganizationMemberChecker
}

func NewCheckinService(checkins *repository.CheckinRepository, members OrganizationMemberChecker) *CheckinService {
	return &CheckinService{checkins: checkins, members: members}
}

func (s *CheckinService) checkIn(ctx context.Context, scannerID uuid.UUID, qrCode string) (*model.Checkin, error) {
	if err := ValidateCheckinInput(qrCode); err != nil {
		return nil, err
	}

	target, err := s.checkins.FindTargetByQR(ctx, qrCode)
	if err != nil {
		return nil, err
	}

	isStaff, err := s.members.IsMember(ctx, target.OrganizationID, scannerID)
	if err != nil {
		return nil, fmt.Errorf("cek keanggotaan organisasi: %w", err)
	}

	if err := CanCheckIn(*target, isStaff); err != nil {
		return nil, err
	}

	checkin := &model.Checkin{TicketID: target.TicketID, CheckedInBy: scannerID}
	if err := s.checkins.Create(ctx, checkin); err != nil {
		if errors.Is(err, repository.ErrCheckinExists) {
			return nil, ErrAlreadyCheckedIn
		}
		return nil, err
	}
	return checkin, nil
}
