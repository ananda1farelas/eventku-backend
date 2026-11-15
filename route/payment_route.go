package route

import (
	"github.com/gofiber/fiber/v2"

	"eventku/app/model"
	"eventku/app/service"
	"eventku/helper"
)

// RegisterPaymentRoutes memasang endpoint pembayaran (mock).
func RegisterPaymentRoutes(app *fiber.App, protect fiber.Handler, payments *service.PaymentService) {
	app.Post("/registrations/:id/payments", protect, createPayment(payments))
	app.Get("/me/payments", protect, listMyPayments(payments))
	app.Get("/payments/:id", protect, getPayment(payments))
	app.Post("/payments/:id/pay", protect, payPayment(payments))
	app.Post("/payments/:id/fail", protect, failPayment(payments))
}

// @Summary     Buat tagihan pembayaran (peserta)
// @Description Hanya untuk registrasi berstatus registered pada event berbayar. Nominal diambil dari harga event.
// @Tags        Payments
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string                     true "ID registrasi (UUID)"
// @Param       body body model.CreatePaymentRequest true "Metode: transfer | ewallet | qris"
// @Success     201 {object} model.WebResponse
// @Failure     409 {object} model.WebResponse
// @Router      /registrations/{id}/payments [post]
func createPayment(svc *service.PaymentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		var req model.CreatePaymentRequest
		if err := helper.BindAndValidate(c, &req); err != nil {
			return err
		}
		p, err := svc.Create(c.UserContext(), currentActor(c), id, req)
		if err != nil {
			return err
		}
		return helper.Created(c, "tagihan pembayaran dibuat", p)
	}
}

// @Summary  Pembayaran saya
// @Tags     Payments
// @Security BearerAuth
// @Success  200 {object} model.WebResponse
// @Router   /me/payments [get]
func listMyPayments(svc *service.PaymentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items, err := svc.ListMine(c.UserContext(), currentActor(c))
		if err != nil {
			return err
		}
		return helper.Success(c, "pembayaran saya", items)
	}
}

// @Summary  Detail pembayaran (pemilik / admin)
// @Tags     Payments
// @Security BearerAuth
// @Param    id path string true "ID pembayaran (UUID)"
// @Success  200 {object} model.WebResponse
// @Router   /payments/{id} [get]
func getPayment(svc *service.PaymentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		d, err := svc.Get(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		return helper.Success(c, "detail pembayaran", d)
	}
}

// @Summary     Simulasi bayar sukses (mock)
// @Description pending -> paid. Pengganti callback payment gateway.
// @Tags        Payments
// @Security    BearerAuth
// @Param       id path string true "ID pembayaran (UUID)"
// @Success     200 {object} model.WebResponse
// @Router      /payments/{id}/pay [post]
func payPayment(svc *service.PaymentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		d, err := svc.Pay(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		return helper.Success(c, "pembayaran berhasil", d)
	}
}

// @Summary     Simulasi bayar gagal (mock)
// @Description pending -> failed. Setelah itu peserta boleh membuat tagihan baru.
// @Tags        Payments
// @Security    BearerAuth
// @Param       id path string true "ID pembayaran (UUID)"
// @Success     200 {object} model.WebResponse
// @Router      /payments/{id}/fail [post]
func failPayment(svc *service.PaymentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		d, err := svc.Fail(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		return helper.Success(c, "pembayaran gagal", d)
	}
}