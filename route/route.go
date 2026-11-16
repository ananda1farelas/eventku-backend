package route

import (
	"github.com/gofiber/fiber/v2"

	"eventku/app/service"
)

// RegisterRoutes jadi satu pintu pendaftaran semua route. `protect` dioper
// sebagai fiber.Handler yang SUDAH JADI (dibangun di config.SetupApp),
// supaya route tidak perlu meng-import config untuk dapat *config.Config —
// sesuai aturan import: route hanya boleh import app/service, middleware,
// helper.
func RegisterRoutes(
	app *fiber.App,
	authService *service.AuthService,
	orgService *service.OrganizationService,
	rbacService *service.RBACService,
	protect fiber.Handler,
) {
	app.Get("/health", healthCheck)

	RegisterAuthRoutes(app, authService, protect)
	RegisterOrganizationRoutes(app, orgService, protect)

	// RegisterEventRoutes(app, eventService, protect, rbacService)
	// RegisterRegistrationRoutes(app, registrationService, protect)
	// RegisterTicketRoutes(app, ticketService, protect)
	// RegisterCheckinRoutes(app, checkinService, protect)
	_ = rbacService // dipakai begitu RegisterEventRoutes (Orang 2) di-uncomment
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "eventku",
	})
}
