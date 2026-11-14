package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
)

var (
	ErrRegistrationEventNotOpen     = errors.New("pendaftaran belum dibuka atau sudah ditutup untuk event ini")
	ErrRegistrationDeadlinePassed   = errors.New("batas waktu pendaftaran sudah lewat")
	ErrAlreadyRegistered            = errors.New("kamu sudah terdaftar atau masuk waitlist di event ini")
	ErrRegistrationForbidden        = errors.New("pendaftaran ini bukan milikmu")
	ErrRegistrationAlreadyCancelled = errors.New("pendaftaran ini sudah dibatalkan")
)

const registrationEventOpen = "open"

func CanRegister(info model.EventRegistrationInfo, existing *model.Registration, now time.Time) error {
	if info.Status != registrationEventOpen {
		return ErrRegistrationEventNotOpen
	}
	if now.After(info.RegistrationDeadline) {
		return ErrRegistrationDeadlinePassed
	}
	if existing != nil && existing.Status != model.RegistrationStatusCancelled {
		return ErrAlreadyRegistered
	}
	return nil
}

func DecideStatus(quota, registeredCount int) string {
	if registeredCount < quota {
		return model.RegistrationStatusRegistered
	}
	return model.RegistrationStatusWaitlist
}

// CanCancel ngecek apakah user boleh membatalkan registrasi ini: harus
// miliknya sendiri dan belum berstatus cancelled.
func CanCancel(reg *model.Registration, userID uuid.UUID) error {
	if reg.UserID != userID {
		return ErrRegistrationForbidden
	}
	if reg.Status == model.RegistrationStatusCancelled {
		return ErrRegistrationAlreadyCancelled
	}
	return nil
}

func ShouldPromoteWaitlist(previousStatus string) bool {
	return previousStatus == model.RegistrationStatusRegistered
}