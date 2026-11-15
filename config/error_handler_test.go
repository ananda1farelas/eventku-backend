package config

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"eventku/app/model"
	"eventku/helper"
)

func TestErrorHandler(t *testing.T) {
	registered := errors.New("error terdaftar di registry")
	helper.RegisterError(registered, fiber.StatusConflict)

	app := fiber.New(fiber.Config{ErrorHandler: NewErrorHandler(nil)})
	app.Get("/app-error", func(c *fiber.Ctx) error { return helper.NewForbidden("dilarang") })
	app.Get("/registered", func(c *fiber.Ctx) error { return registered })
	app.Get("/fiber-error", func(c *fiber.Ctx) error { return fiber.NewError(fiber.StatusUnauthorized, "token salah") })
	app.Get("/boom", func(c *fiber.Ctx) error { return errors.New("password db: rahasia") })
	app.Get("/validation", func(c *fiber.Ctx) error {
		return helper.NewValidation("gagal", model.ErrorDetail{Field: "title", Message: "wajib diisi"})
	})

	tests := []struct {
		path     string
		wantCode int
		wantMsg  string
	}{
		{"/app-error", 403, "dilarang"},
		{"/registered", 409, "error terdaftar di registry"},
		{"/fiber-error", 401, "token salah"},
		{"/boom", 500, "terjadi kesalahan pada server"}, // detail error tidak bocor
		{"/tidak-ada", 404, "Cannot GET /tidak-ada"},
		{"/validation", 422, "gagal"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest("GET", tt.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantCode)
			}
			body, _ := io.ReadAll(resp.Body)
			var res model.WebResponse
			if err := json.Unmarshal(body, &res); err != nil {
				t.Fatalf("body bukan JSON WebResponse: %s", body)
			}
			if res.Success || res.Message != tt.wantMsg {
				t.Fatalf("response = %+v, want success=false message=%q", res, tt.wantMsg)
			}
		})
	}
}