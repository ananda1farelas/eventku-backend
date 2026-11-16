package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// PermissionChecker sengaja didefinisikan sebagai func type di sini, BUKAN
// dengan middleware nge-import package service langsung (aturan import di
// docs/PEMBAGIAN-TUGAS.md: middleware cuma boleh import helper). Bentuknya
// cocok sama method value service.RBACService.Can, jadi route.go (yang
// memang boleh import service) tinggal ngoper `rbacService.Can` sebagai
// argumen saat wiring — tanpa middleware perlu tahu package service sama
// sekali.
type PermissionChecker func(ctx context.Context, role, permission string) (bool, error)

// RequirePermission dipasang SETELAH Protect(), buat endpoint yang gate-nya
// granular per-permission (misal "events.moderate"), beda dengan RequireRole
// yang cuma ngecek role kasar. Fail closed: error dari checker ATAU hasil
// false, dua-duanya ditolak (403) — gak ada jalur yang diam-diam meloloskan.
func RequirePermission(permission string, check PermissionChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := CurrentRole(c)

		allowed, err := check(c.Context(), role, permission)
		if err != nil || !allowed {
			return fiber.NewError(fiber.StatusForbidden, "tidak punya izin untuk aksi ini")
		}

		return c.Next()
	}
}
