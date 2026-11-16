package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/service"
	"eventku/helper"
	"eventku/middleware"
)

// currentActor membaca siapa yang login dari context (diisi middleware.Protect).
func currentActor(c *fiber.Ctx) service.Actor {
	return service.Actor{UserID: middleware.CurrentUserID(c), Role: middleware.CurrentRole(c)}
}

// RegisterEventRoutes memasang endpoint event dan moderasi.
// protect = middleware.Protect(cfg), dikirim dari route.go supaya file ini
// tidak perlu tahu soal config.
func RegisterEventRoutes(app *fiber.App, protect fiber.Handler, events *service.EventService) {
	g := app.Group("/events", protect)

	g.Get("/", listEvents(events))
	g.Post("/", createEvent(events))
	g.Get("/:id", getEvent(events))
	g.Put("/:id", updateEvent(events))
	g.Delete("/:id", deleteEvent(events))

	g.Patch("/:id/submit", submitEvent(events))
	g.Patch("/:id/approve", middleware.RequireRole(model.RoleAdmin), approveEvent(events))
	g.Patch("/:id/reject", middleware.RequireRole(model.RoleAdmin), rejectEvent(events))
}

// @Summary     Daftar event
// @Description Admin melihat semua event. Panitia melihat semua event organisasinya (isi organization_id). Selain itu hanya open/closed/done.
// @Tags        Events
// @Security    BearerAuth
// @Produce     json
// @Param       page            query int    false "Halaman (default 1)"
// @Param       limit           query int    false "Jumlah per halaman (default 10, maks 100)"
// @Param       search          query string false "Cari judul"
// @Param       category        query string false "Filter kategori"
// @Param       status          query string false "Filter status"
// @Param       organization_id query string false "Filter organisasi (UUID)"
// @Success     200 {object} model.WebResponse
// @Router      /events [get]
func listEvents(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		lq := helper.ParseListQuery(c)
		q := service.ListEventsQuery{
			Page: lq.Page, Limit: lq.Limit, Search: lq.Search,
			Category: c.Query("category"), Status: c.Query("status"),
		}
		if raw := c.Query("organization_id"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				return helper.NewValidation("organization_id harus berupa UUID yang valid")
			}
			q.OrganizationID = &id
		}

		events, meta, err := svc.List(c.UserContext(), currentActor(c), q)
		if err != nil {
			return err
		}
		return helper.Success(c, "daftar event", events, meta)
	}
}

// @Summary  Detail event
// @Tags     Events
// @Security BearerAuth
// @Param    id path string true "ID event (UUID)"
// @Success  200 {object} model.WebResponse
// @Failure  404 {object} model.WebResponse
// @Router   /events/{id} [get]
func getEvent(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		ev, err := svc.Get(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		return helper.Success(c, "detail event", ev)
	}
}

// @Summary     Buat event (panitia)
// @Description Event baru berstatus draft. Harus anggota organisasi pemilik event.
// @Tags        Events
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body model.CreateEventRequest true "Data event"
// @Success     201 {object} model.WebResponse
// @Failure     403 {object} model.WebResponse
// @Failure     422 {object} model.WebResponse
// @Router      /events [post]
func createEvent(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.CreateEventRequest
		if err := helper.BindAndValidate(c, &req); err != nil {
			return err
		}
		ev, err := svc.Create(c.UserContext(), currentActor(c), req)
		if err != nil {
			return err
		}
		return helper.Created(c, "event berhasil dibuat", ev)
	}
}

// @Summary  Ubah event (panitia, hanya draft/rejected)
// @Tags     Events
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id   path string                   true "ID event (UUID)"
// @Param    body body model.UpdateEventRequest true "Data event lengkap"
// @Success  200 {object} model.WebResponse
// @Router   /events/{id} [put]
func updateEvent(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		var req model.UpdateEventRequest
		if err := helper.BindAndValidate(c, &req); err != nil {
			return err
		}
		ev, err := svc.Update(c.UserContext(), currentActor(c), id, req)
		if err != nil {
			return err
		}
		return helper.Success(c, "event berhasil diubah", ev)
	}
}

// @Summary  Hapus event (panitia, hanya draft/rejected)
// @Tags     Events
// @Security BearerAuth
// @Param    id path string true "ID event (UUID)"
// @Success  204
// @Router   /events/{id} [delete]
func deleteEvent(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		if err := svc.Delete(c.UserContext(), currentActor(c), id); err != nil {
			return err
		}
		return helper.NoContent(c)
	}
}

// @Summary  Ajukan event ke admin (panitia)
// @Tags     Events
// @Security BearerAuth
// @Param    id path string true "ID event (UUID)"
// @Success  200 {object} model.WebResponse
// @Router   /events/{id}/submit [patch]
func submitEvent(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		ev, err := svc.Submit(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		return helper.Success(c, "event diajukan untuk ditinjau", ev)
	}
}

// @Summary  Setujui event (admin)
// @Tags     Events
// @Security BearerAuth
// @Param    id path string true "ID event (UUID)"
// @Success  200 {object} model.WebResponse
// @Router   /events/{id}/approve [patch]
func approveEvent(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		ev, err := svc.Approve(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		return helper.Success(c, "event disetujui dan dibuka", ev)
	}
}

// @Summary  Tolak event (admin)
// @Tags     Events
// @Security BearerAuth
// @Accept   json
// @Param    id   path string                   true "ID event (UUID)"
// @Param    body body model.RejectEventRequest true "Alasan penolakan"
// @Success  200 {object} model.WebResponse
// @Router   /events/{id}/reject [patch]
func rejectEvent(svc *service.EventService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		var req model.RejectEventRequest
		if err := helper.BindAndValidate(c, &req); err != nil {
			return err
		}
		ev, err := svc.Reject(c.UserContext(), currentActor(c), id, req.Reason)
		if err != nil {
			return err
		}
		return helper.Success(c, "event ditolak", ev)
	}
}