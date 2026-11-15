package helper

import (
	"math"

	"github.com/gofiber/fiber/v2"

	"eventku/app/model"
)

func Success(c *fiber.Ctx, msg string, data any, meta ...*model.Meta) error {
	res := model.WebResponse{Success: true, Message: msg, Data: data}
	if len(meta) > 0 {
		res.Meta = meta[0]
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func Created(c *fiber.Ctx, msg string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{Success: true, Message: msg, Data: data})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Fail menulis response error dari AppError. Error yang bukan AppError
// (dan tidak terdaftar di registry) dijawab 500 tanpa membocorkan detailnya.
func Fail(c *fiber.Ctx, err error) error {
	if ae := ResolveError(err); ae != nil {
		return c.Status(ae.Code).JSON(model.WebResponse{Success: false, Message: ae.Message, Errors: ae.Details})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(model.WebResponse{
		Success: false,
		Message: "terjadi kesalahan pada server",
	})
}

func FailValidation(c *fiber.Ctx, details []model.ErrorDetail) error {
	return Fail(c, NewValidation("validasi gagal", details...))
}

// NewMeta menghitung total halaman dari total data.
func NewMeta(page, limit int, total int64) *model.Meta {
	pages := 0
	if limit > 0 {
		pages = int(math.Ceil(float64(total) / float64(limit)))
	}
	return &model.Meta{Page: page, Limit: limit, Total: total, TotalPages: pages}
}