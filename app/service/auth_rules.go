package service

import (
	"errors"
	"regexp"
	"strings"
)

// Error domain buat rule-rule auth. Dipisah dari error repository supaya
// handler bisa bedain "400 salah input" vs "500/404 masalah data".
var (
	ErrPasswordTooWeak      = errors.New("password minimal 8 karakter dan harus mengandung huruf serta angka")
	ErrInvalidEmailFormat   = errors.New("format email tidak valid")
	ErrFullNameTooShort     = errors.New("nama lengkap minimal 3 karakter")
	ErrInvalidCredentials   = errors.New("email atau password salah")
	ErrAccountInactive      = errors.New("akun tidak aktif, hubungi admin")
	ErrRefreshTokenExpired  = errors.New("refresh token sudah kedaluwarsa")
	ErrRefreshTokenRevoked  = errors.New("refresh token sudah tidak berlaku")
	ErrRefreshTokenNotFound = errors.New("refresh token tidak valid")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateFullName ngecek nama lengkap bukan cuma whitespace dan minimal
// 3 karakter (konsisten sama tag validate:"min=3" di RegisterRequest,
// tapi di-cek ulang di sini supaya rule ini tetap jalan walau validator
// tag suatu saat dilepas/diganti).
func ValidateFullName(name string) error {
	if len(strings.TrimSpace(name)) < 3 {
		return ErrFullNameTooShort
	}
	return nil
}

// ValidateEmailFormat ngecek format email secara longgar (bukan verifikasi
// email beneran ada). Pengecekan "sudah dipakai atau belum" itu tanggung
// jawab repository (ErrDuplicateEmail), bukan rule ini.
func ValidateEmailFormat(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmailFormat
	}
	return nil
}

// ValidatePasswordStrength adalah rule inti registrasi: minimal 8 karakter,
// dan harus ada minimal 1 huruf DAN 1 angka. Sengaja gak maksa simbol
// khusus/huruf besar supaya gak terlalu ribet buat user kampus, tapi tetap
// lebih kuat dari sekadar "panjang minimal 8".
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooWeak
	}

	var hasLetter, hasDigit bool
	for _, c := range password {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
			hasLetter = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return ErrPasswordTooWeak
	}
	return nil
}

// ValidateRegisterInput jadi satu pintu validasi sebelum service.Register
// nyentuh database sama sekali (hash password, dst). Urutan pengecekan
// sengaja nama dulu baru email baru password, biar pesan error yang
// dibalikin ke user intuitif (dari field paling atas form).
func ValidateRegisterInput(fullName, email, password string) error {
	if err := ValidateFullName(fullName); err != nil {
		return err
	}
	if err := ValidateEmailFormat(email); err != nil {
		return err
	}
	if err := ValidatePasswordStrength(password); err != nil {
		return err
	}
	return nil
}

// NormalizeEmail dipanggil sebelum register DAN sebelum login, supaya
// "User@Mail.com" dan "user@mail.com" dianggap akun yang sama secara
// konsisten di kedua alur.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
