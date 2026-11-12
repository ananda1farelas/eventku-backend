package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"eventku/config"
)

// Key context lokal Fiber. Pakai tipe custom (bukan string polos) biar
// gak collision sama locals key lain yang mungkin dipasang middleware lain.
type ctxKey string

const (
	localUserID ctxKey = "current_user_id"
	localRole   ctxKey = "current_role"
)

var (
	ErrMissingToken = errors.New("header Authorization tidak ada atau salah format")
	ErrInvalidToken = errors.New("token tidak valid atau sudah kedaluwarsa")
)

// jwtClaims HARUS sama persis bentuknya dengan yang di-generate
// service.AuthService.generateAccessToken (app/service/auth_service.go).
// Kalau salah satu berubah, satunya juga wajib diupdate.
type jwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Protect adalah middleware wajib-login. Pasang di route yang butuh user
// sudah login, lalu di handler/service tinggal panggil CurrentUserID(c)
// dan CurrentRole(c) buat tau siapa yang request.
func Protect(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, ErrMissingToken.Error())
		}
		rawToken := strings.TrimPrefix(header, "Bearer ")

		claims := &jwtClaims{}
		token, err := jwt.ParseWithClaims(rawToken, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("signing method tidak sesuai")
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, ErrInvalidToken.Error())
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, ErrInvalidToken.Error())
		}

		c.Locals(localUserID, userID)
		c.Locals(localRole, claims.Role)

		return c.Next()
	}
}

// RequireRole dipasang SETELAH Protect() di route yang mau dibatasi role
// tertentu, misal endpoint moderasi event yang cuma boleh admin:
//
//	app.Post("/events/:id/moderate", middleware.Protect(cfg), middleware.RequireRole(model.RoleAdmin), handler)
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := CurrentRole(c)
		for _, allowed := range allowedRoles {
			if role == allowed {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, "tidak punya akses untuk aksi ini")
	}
}

// CurrentUserID dipanggil di handler/service SETELAH route dipasangi
// Protect(). Kalau dipanggil tanpa Protect() lebih dulu, bakal balikin
// uuid.Nil (bukan panic) — tapi secara desain ini seharusnya gak pernah
// kejadian kalau routing-nya benar.
func CurrentUserID(c *fiber.Ctx) uuid.UUID {
	id, ok := c.Locals(localUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}

// CurrentRole dipanggil di handler/service SETELAH route dipasangi
// Protect(). Balikin string kosong kalau belum login / middleware belum
// dipasang.
func CurrentRole(c *fiber.Ctx) string {
	role, ok := c.Locals(localRole).(string)
	if !ok {
		return ""
	}
	return role
}
