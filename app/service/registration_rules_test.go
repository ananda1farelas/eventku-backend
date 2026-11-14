package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
)

func TestCanRegister(t *testing.T) {
	now := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)

	open := model.EventRegistrationInfo{
		Status:               "open",
		Quota:                10,
		RegistrationDeadline: now.Add(24 * time.Hour),
	}
	closed := open
	closed.Status = "closed"
	expired := open
	expired.RegistrationDeadline = now.Add(-time.Hour)
	atDeadline := open
	atDeadline.RegistrationDeadline = now

	tests := []struct {
		name     string
		info     model.EventRegistrationInfo
		existing *model.Registration
		wantErr  error
	}{
		{"event open dan belum terdaftar", open, nil, nil},
		{"event tidak open", closed, nil, ErrRegistrationEventNotOpen},
		{"deadline sudah lewat", expired, nil, ErrRegistrationDeadlinePassed},
		{"tepat di deadline masih boleh", atDeadline, nil, nil},
		{"sudah registered", open, &model.Registration{Status: model.RegistrationStatusRegistered}, ErrAlreadyRegistered},
		{"sudah di waitlist", open, &model.Registration{Status: model.RegistrationStatusWaitlist}, ErrAlreadyRegistered},
		{"pernah batal boleh daftar lagi", open, &model.Registration{Status: model.RegistrationStatusCancelled}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CanRegister(tt.info, tt.existing, now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanRegister() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecideStatus(t *testing.T) {
	tests := []struct {
		name            string
		quota           int
		registeredCount int
		want            string
	}{
		{"masih kosong", 5, 0, model.RegistrationStatusRegistered},
		{"tinggal satu kursi", 5, 4, model.RegistrationStatusRegistered},
		{"tepat penuh", 5, 5, model.RegistrationStatusWaitlist},
		{"sudah melebihi kuota", 5, 6, model.RegistrationStatusWaitlist},
		{"kuota nol", 0, 0, model.RegistrationStatusWaitlist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecideStatus(tt.quota, tt.registeredCount); got != tt.want {
				t.Fatalf("DecideStatus(%d, %d) = %q, want %q", tt.quota, tt.registeredCount, got, tt.want)
			}
		})
	}
}

func TestCanCancel(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()

	tests := []struct {
		name    string
		reg     *model.Registration
		userID  uuid.UUID
		wantErr error
	}{
		{"pemilik membatalkan registered", &model.Registration{UserID: owner, Status: model.RegistrationStatusRegistered}, owner, nil},
		{"pemilik membatalkan waitlist", &model.Registration{UserID: owner, Status: model.RegistrationStatusWaitlist}, owner, nil},
		{"bukan pemilik", &model.Registration{UserID: owner, Status: model.RegistrationStatusRegistered}, other, ErrRegistrationForbidden},
		{"sudah dibatalkan", &model.Registration{UserID: owner, Status: model.RegistrationStatusCancelled}, owner, ErrRegistrationAlreadyCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CanCancel(tt.reg, tt.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanCancel() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestShouldPromoteWaitlist(t *testing.T) {
	tests := []struct {
		previousStatus string
		want           bool
	}{
		{model.RegistrationStatusRegistered, true},
		{model.RegistrationStatusWaitlist, false},
		{model.RegistrationStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.previousStatus, func(t *testing.T) {
			if got := ShouldPromoteWaitlist(tt.previousStatus); got != tt.want {
				t.Fatalf("ShouldPromoteWaitlist(%q) = %v, want %v", tt.previousStatus, got, tt.want)
			}
		})
	}
}
