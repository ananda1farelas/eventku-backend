package route

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/service"
	"eventku/helper"
	"eventku/middleware"
)

// RegisterOrganizationRoutes memetakan endpoint /organizations ke method
// OrganizationService. Semua endpoint di sini wajib login (protect
// dipasang di level group), tapi otorisasi lebih detail (harus anggota,
// harus owner) dicek DI DALAM service lewat IsMember/IsOwner — bukan di
// route, supaya rule-nya testable dan gak nempel ke fiber.Ctx.
func RegisterOrganizationRoutes(app *fiber.App, orgService *service.OrganizationService, protect fiber.Handler) {
	org := app.Group("/organizations", protect)

	org.Post("/", func(c *fiber.Ctx) error {
		var req model.CreateOrganizationRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.NewValidation("body request tidak valid")
		}
		if ae := helper.Validate(&req); ae != nil {
			return ae
		}

		ownerID := middleware.CurrentUserID(c)
		created, err := orgService.Create(c.Context(), ownerID, req.Name, req.Description)
		if err != nil {
			return err
		}
		return helper.Created(c, "organisasi berhasil dibuat", created)
	})

	org.Get("/", func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * limit

		orgs, err := orgService.List(c.Context(), limit, offset)
		if err != nil {
			return err
		}
		return helper.Success(c, "berhasil mengambil daftar organisasi", orgs)
	})

	org.Put("/:id", func(c *fiber.Ctx) error {
		orgID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return helper.NewValidation("id organisasi tidak valid")
		}

		var req model.UpdateOrganizationRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.NewValidation("body request tidak valid")
		}
		if ae := helper.Validate(&req); ae != nil {
			return ae
		}

		requesterID := middleware.CurrentUserID(c)
		if err := orgService.Update(c.Context(), requesterID, orgID, req.Name, req.Description); err != nil {
			return err
		}
		return helper.NoContent(c)
	})

	org.Post("/:id/members", func(c *fiber.Ctx) error {
		orgID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return helper.NewValidation("id organisasi tidak valid")
		}

		var req model.AddMemberRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.NewValidation("body request tidak valid")
		}
		if ae := helper.Validate(&req); ae != nil {
			return ae
		}

		requesterID := middleware.CurrentUserID(c)
		if err := orgService.AddMember(c.Context(), requesterID, orgID, req.UserID, req.Position); err != nil {
			return err
		}
		return helper.NoContent(c)
	})

	org.Get("/:id/members", func(c *fiber.Ctx) error {
		orgID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return helper.NewValidation("id organisasi tidak valid")
		}

		requesterID := middleware.CurrentUserID(c)
		members, err := orgService.ListMembers(c.Context(), requesterID, orgID)
		if err != nil {
			return err
		}
		return helper.Success(c, "berhasil mengambil daftar anggota", members)
	})
}
