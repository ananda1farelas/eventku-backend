package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"eventku/app/model"
	"eventku/app/repository"
	"eventku/config"
)

// AuthService isi semua use-case auth: register, login, refresh, me.
// Validasi "bentuk" input (format, kekuatan password) sudah selesai di
// auth_rules.go sebelum method di sini dipanggil — service fokus ke
// orkestrasi: hash password, generate token, baca/tulis repo.
type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	cfg              *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, refreshTokenRepo *repository.RefreshTokenRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
	}
}

// jwtClaims adalah isi payload access token. RoleName diselipin di sini
// (bukan cuma UserID) supaya middleware bisa langsung tau role tanpa query
// DB tiap request yang gak butuh data user lengkap.
type jwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Register memvalidasi (lewat ValidateRegisterInput di pemanggil/handler,
// sudah dipastikan lolos sebelum sampai sini), hash password dengan bcrypt,
// lalu simpan user baru dengan role default "user" — TIDAK ADA jalur publik
// untuk daftar jadi admin.
func (s *AuthService) Register(ctx context.Context, fullName, email, password string) (*model.AuthResponse, error) {
	email = NormalizeEmail(email)

	roleID, err := s.userRepo.FindRoleIDByName(ctx, model.RoleUser)
	if err != nil {
		return nil, fmt.Errorf("ambil role default: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		FullName:     fullName,
		Email:        email,
		PasswordHash: string(hash),
		RoleID:       roleID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err // bisa ErrDuplicateEmail, biar handler yang terjemahin ke 409
	}
	user.RoleName = model.RoleUser

	return s.issueTokenPair(ctx, user)
}

// Login verifikasi email+password. Pesan error SENGAJA sama persis
// (ErrInvalidCredentials) baik email gak ketemu maupun password salah,
// supaya gak jadi celah buat enumerasi "email mana yang terdaftar".
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	email = NormalizeEmail(email)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("cari user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, user)
}

// Refresh melakukan token rotation: refresh token lama divalidasi lalu
// DIREVOKE, dan sepasang token baru (access+refresh) diterbitkan. Kalau
// refresh token yang sama dipakai dua kali, percobaan kedua akan gagal
// karena sudah revoked — tanda kemungkinan token itu bocor.
func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*model.AuthResponse, error) {
	hash := hashToken(rawRefreshToken)

	user, rt, err := s.refreshTokenRepo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("cari refresh token: %w", err)
	}

	if rt.RevokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}
	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if err := s.refreshTokenRepo.Revoke(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("revoke refresh token lama: %w", err)
	}

	return s.issueTokenPair(ctx, user)
}

// Me dipanggil GET /auth/me setelah middleware JWT nge-set CurrentUserID.
func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*model.MeResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return model.NewMeResponse(user), nil
}

// issueTokenPair bikin access token (JWT, short-lived) dan refresh token
// (random string, long-lived, disimpan dalam bentuk hash) sekaligus,
// dipakai bareng oleh Register/Login/Refresh biar format response konsisten.
func (s *AuthService) issueTokenPair(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	accessToken, expiresIn, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	rawRefreshToken, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshExpiresAt := time.Now().Add(time.Duration(s.cfg.JWTRefreshExpiryDays) * 24 * time.Hour)
	if err := s.refreshTokenRepo.Store(ctx, user.ID, hashToken(rawRefreshToken), refreshExpiresAt); err != nil {
		return nil, fmt.Errorf("simpan refresh token: %w", err)
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *AuthService) generateAccessToken(user *model.User) (string, int, error) {
	expiry := time.Duration(s.cfg.JWTAccessExpiryMinutes) * time.Minute
	now := time.Now()

	claims := jwtClaims{
		UserID: user.ID.String(),
		Role:   user.RoleName,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}
	return signed, int(expiry.Seconds()), nil
}

// generateRandomToken bikin refresh token mentah dari random bytes
// (bukan JWT), karena refresh token gak perlu bisa "dibaca" isinya —
// cukup unik, panjang, dan susah ditebak.
func generateRandomToken(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken nge-hash refresh token mentah pakai SHA-256 sebelum
// disimpan/dicari di DB. Beda dengan bcrypt buat password: refresh token
// sudah random & panjang, jadi gak butuh salt+cost factor, cukup hash cepat.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
