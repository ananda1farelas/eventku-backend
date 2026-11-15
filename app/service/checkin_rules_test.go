package service

import (
	"errors"
	"testing"
	"time"

	"eventku/app/model"
)

func TestValidateCheckinInput(t *testing.T) {
	tests := []struct {
		name    string
		qr      string
		wantErr error
	}{
		{"qr terisi", "abc123", nil},
		{"qr kosong", "", ErrCheckinQRRequired},
		{"qr cuma spasi", "   ", ErrCheckinQRRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCheckinInput(tt.qr)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateCheckinInput() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCanCheckIn(t *testing.T) {
	now := time.Now()

	valid := model.CheckinTarget{RegistrationStatus: model.RegistrationStatusRegistered}
	used := valid
	used.CheckedInAt = &now
	cancelled := valid
	cancelled.RegistrationStatus = model.RegistrationStatusCancelled
	waitlist := valid
	waitlist.RegistrationStatus = model.RegistrationStatusWaitlist

	tests := []struct {
		name    string
		target  model.CheckinTarget
		isStaff bool
		wantErr error
	}{
		{"panitia dan tiket valid", valid, true, nil},
		{"bukan panitia", valid, false, ErrCheckinForbidden},
		{"registrasi dibatalkan", cancelled, true, ErrCheckinTicketInvalid},
		{"registrasi masih waitlist", waitlist, true, ErrCheckinTicketInvalid},
		{"tiket sudah dipakai", used, true, ErrAlreadyCheckedIn},
		{"bukan panitia dan tiket sudah dipakai", used, false, ErrCheckinForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CanCheckIn(tt.target, tt.isStaff)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanCheckIn() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
