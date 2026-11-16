package route

import (
	"github.com/gofiber/fiber/v2"

	"eventku/app/model"
	"eventku/app/service"
	"eventku/helper"
	"eventku/middleware"
)

// RegisterAuthRoutes memetakan endpoint /auth langsung ke method
// AuthService. Tidak ada layer handler terpisah (sesuai ketentuan modul):
// parsing request, validasi format (helper.Validate), dan pemanggilan
// service semua terjadi di closure di bawah ini. Error dari service
// cukup di-`return` apa adanya — ErrorHandler terpusat (config.NewErrorHandler)
// yang menerjemahkannya lewat helper.ResolveError.
func RegisterAuthRoutes(app *fiber.App, authService *service.AuthService, protect fiber.Handler) {
	auth := app.Group("/auth")

	auth.Post("/register", func(c *fiber.Ctx) error {
		var req model.RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.NewValidation("body request tidak valid")
		}
		if ae := helper.Validate(&req); ae != nil {
			return ae
		}

		res, err := authService.Register(c.Context(), req.FullName, req.Email, req.Password)
		if err != nil {
			return err
		}
		return helper.Created(c, "registrasi berhasil", res)
	})

	auth.Post("/login", func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.NewValidation("body request tidak valid")
		}
		if ae := helper.Validate(&req); ae != nil {
			return ae
		}

		res, err := authService.Login(c.Context(), req.Email, req.Password)
		if err != nil {
			return err
		}
		return helper.Success(c, "login berhasil", res)
	})

	auth.Post("/refresh", func(c *fiber.Ctx) error {
		var req model.RefreshRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.NewValidation("body request tidak valid")
		}
		if ae := helper.Validate(&req); ae != nil {
			return ae
		}

		res, err := authService.Refresh(c.Context(), req.RefreshToken)
		if err != nil {
			return err
		}
		return helper.Success(c, "token berhasil diperbarui", res)
	})

	auth.Get("/me", protect, func(c *fiber.Ctx) error {
		userID := middleware.CurrentUserID(c)

		res, err := authService.Me(c.Context(), userID)
		if err != nil {
			return err
		}
		return helper.Success(c, "berhasil mengambil profil", res)
	})
}
