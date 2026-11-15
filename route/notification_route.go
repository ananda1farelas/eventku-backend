package route

import (
	"github.com/gofiber/fiber/v2"

	"eventku/app/service"
	"eventku/helper"
)

// RegisterNotificationRoutes memasang endpoint notifikasi milik user login.
func RegisterNotificationRoutes(app *fiber.App, protect fiber.Handler, notifications *service.NotificationService) {
	app.Get("/me/notifications", protect, listMyNotifications(notifications))
	app.Patch("/me/notifications/:id/read", protect, readNotification(notifications))
}

// @Summary  Notifikasi saya
// @Tags     Notifications
// @Security BearerAuth
// @Param    page   query int  false "Halaman (default 1)"
// @Param    limit  query int  false "Jumlah per halaman (default 10, maks 100)"
// @Param    unread query bool false "true = hanya yang belum dibaca"
// @Success  200 {object} model.WebResponse
// @Router   /me/notifications [get]
func listMyNotifications(svc *service.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		lq := helper.ParseListQuery(c)
		items, meta, err := svc.ListMine(c.UserContext(), currentActor(c), lq.Page, lq.Limit, c.QueryBool("unread", false))
		if err != nil {
			return err
		}
		return helper.Success(c, "notifikasi saya", items, meta)
	}
}

// @Summary  Tandai notifikasi sudah dibaca
// @Tags     Notifications
// @Security BearerAuth
// @Param    id path string true "ID notifikasi (UUID)"
// @Success  204
// @Router   /me/notifications/{id}/read [patch]
func readNotification(svc *service.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		if err := svc.MarkRead(c.UserContext(), currentActor(c), id); err != nil {
			return err
		}
		return helper.NoContent(c)
	}
}