package service

import (
	"net/http"

	"eventku/app/repository"
	"eventku/helper"
)

// RegisterErrorMappings memberi tahu ErrorHandler kode HTTP untuk error
// domain yang dibuat sebagai errors.New biasa (modul Auth, Registrasi,
// Check-in, dan error repository). Panggil SEKALI di main.go sebelum server jalan.
func RegisterErrorMappings() {
	table := map[int][]error{
		http.StatusNotFound: {repository.ErrNotFound},
		http.StatusConflict: {
			repository.ErrDuplicateEmail, repository.ErrRegistrationExists,
			repository.ErrCheckinExists, repository.ErrTicketExists,
			ErrRegistrationEventNotOpen, ErrRegistrationDeadlinePassed,
			ErrAlreadyRegistered, ErrRegistrationAlreadyCancelled,
			ErrCheckinTicketInvalid, ErrAlreadyCheckedIn,
		},
		http.StatusUnprocessableEntity: {
			ErrPasswordTooWeak, ErrInvalidEmailFormat, ErrFullNameTooShort,
			ErrCheckinQRRequired,
		},
		http.StatusUnauthorized: {
			ErrInvalidCredentials, ErrRefreshTokenExpired,
			ErrRefreshTokenRevoked, ErrRefreshTokenNotFound,
		},
		http.StatusForbidden: {
			ErrAccountInactive, ErrRegistrationForbidden, ErrCheckinForbidden,
		},
	}
	for code, errs := range table {
		for _, e := range errs {
			helper.RegisterError(e, code)
		}
	}
}