package model

import "github.com/google/uuid"

// === Request ===
// Validasi di sini cuma yang gampang dicek lewat tag (required, format,
// panjang). Aturan yang lebih "bisnis" (kekuatan password, dst) ada di
// service/auth_rules.go, bukan di struct tag.

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=3,max=150"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Password string `json:"password" validate:"required,min=8,max=72"` // 72 byte = batas bcrypt
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// === Response ===

// AuthResponse dibalikin setelah register, login, dan refresh.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"` // selalu "Bearer"
	ExpiresIn    int    `json:"expires_in"` // detik sampai access token expired
}

// MeResponse dibalikin di GET /auth/me.
type MeResponse struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
}

// NewMeResponse bikin MeResponse dari User. Dipisah dari struct User biar
// gampang kontrol field apa aja yang boleh keluar di /auth/me, tanpa
// nyentuh struct User yang juga dipakai internal di repo/service.
func NewMeResponse(u *User) *MeResponse {
	return &MeResponse{
		ID:       u.ID,
		FullName: u.FullName,
		Email:    u.Email,
		Role:     u.RoleName,
	}
}
