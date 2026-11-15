package config

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"eventku/helper"
)

// NewErrorHandler membuat ErrorHandler terpusat untuk Fiber. Semua handler
// cukup `return err`; di sinilah error diubah jadi response JSON baku.
//
//	app := config.NewFiberApp(config.NewErrorHandler(logger))
func NewErrorHandler(logger *log.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// 1. AppError, atau error biasa yang sudah didaftarkan di registry.
		if ae := helper.ResolveError(err); ae != nil {
			return helper.Fail(c, ae)
		}

		// 2. Error bawaan Fiber (route tidak ada, JWT middleware, dst).
		var fe *fiber.Error
		if errors.As(err, &fe) {
			return helper.Fail(c, &helper.AppError{Code: fe.Code, Message: fe.Message})
		}

		// 3. Sisanya bug/kegagalan sistem: catat detailnya, jawab generik.
		if logger != nil {
			logger.Printf("unhandled error: %s %s: %v", c.Method(), c.Path(), err)
		}
		return helper.Fail(c, err)
	}
}