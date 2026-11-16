package route

import (
	"github.com/gofiber/fiber/v2"

	"eventku/app/service"
	"eventku/helper"
)

// RegisterReportRoutes memasang endpoint rekap peserta dan export CSV.
func RegisterReportRoutes(app *fiber.App, protect fiber.Handler, reports *service.ReportService) {
	app.Get("/events/:id/participants", protect, listParticipants(reports))
	app.Get("/events/:id/participants/export", protect, exportParticipants(reports))
}

// @Summary  Rekap peserta event (admin / panitia event)
// @Tags     Reports
// @Security BearerAuth
// @Param    id     path  string true  "ID event (UUID)"
// @Param    status query string false "registered | waitlist | cancelled"
// @Success  200 {object} model.WebResponse
// @Router   /events/{id}/participants [get]
func listParticipants(svc *service.ReportService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		rows, err := svc.Participants(c.UserContext(), currentActor(c), id, c.Query("status"))
		if err != nil {
			return err
		}
		return helper.Success(c, "rekap peserta", rows)
	}
}

// @Summary  Export rekap peserta ke CSV (admin / panitia event)
// @Tags     Reports
// @Security BearerAuth
// @Produce  text/csv
// @Param    id path string true "ID event (UUID)"
// @Success  200 {file} file
// @Router   /events/{id}/participants/export [get]
func exportParticipants(svc *service.ReportService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		filename, data, err := svc.ExportCSV(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
		c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
		return c.Send(data)
	}
}