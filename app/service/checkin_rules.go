package service

import (
	"errors"
	"strings"

	"eventku/app/model"
)

// Error domain buat rule check-in.
var (
	ErrCheckinQRRequired    = errors.New("kode QR wajib diisi")
	ErrCheckinForbidden     = errors.New("hanya panitia event ini yang boleh melakukan check-in")
	ErrCheckinTicketInvalid = errors.New("tiket tidak valid karena pendaftarannya tidak aktif")
	ErrAlreadyCheckedIn     = errors.New("tiket ini sudah dipakai untuk check-in")
)

// ValidateCheckinInput memastikan kode QR tidak kosong.
func ValidateCheckinInput(qrCode string) error {
	if strings.TrimSpace(qrCode) == "" {
		return ErrCheckinQRRequired
	}
	return nil
}

func CanCheckIn(target model.CheckinTarget, isStaff bool) error {
	if !isStaff {
		return ErrCheckinForbidden
	}
	if target.RegistrationStatus != model.RegistrationStatusRegistered {
		return ErrCheckinTicketInvalid
	}
	if target.CheckedInAt != nil {
		return ErrAlreadyCheckedIn
	}
	return nil
}
