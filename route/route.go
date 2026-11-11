package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterRoutes jadi satu pintu pendaftaran semua route.
// Supaya main.go dan file ini tidak sering bentrok pas merge, tiap modul
// (auth, event, dll) bikin file route/<modul>_route.go sendiri dengan satu
// fungsi RegisterXxxRoutes, lalu dipanggil satu baris di sini.
func RegisterRoutes(app *fiber.App, pool *pgxpool.Pool) {
	app.Get("/health", healthCheck)

	// RegisterAuthRoutes(app, pool)
	// RegisterOrganizationRoutes(app, pool)
	// RegisterEventRoutes(app, pool)
	// RegisterRegistrationRoutes(app, pool)
	// RegisterTicketRoutes(app, pool)
	// RegisterCheckinRoutes(app, pool)
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "eventku",
	})
}
