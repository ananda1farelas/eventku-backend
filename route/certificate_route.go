package route

import (
	"github.com/gofiber/fiber/v2"

	"eventku/app/service"
	"eventku/helper"
)

// RegisterCertificateRoutes memasang endpoint sertifikat.
func RegisterCertificateRoutes(app *fiber.App, protect fiber.Handler, certs *service.CertificateService) {
	app.Get("/me/certificates", protect, listMyCertificates(certs))
	app.Post("/events/:id/certificates/issue", protect, issueCertificates(certs))
}

// @Summary  Sertifikat saya (peserta)
// @Tags     Certificates
// @Security BearerAuth
// @Success  200 {object} model.WebResponse
// @Router   /me/certificates [get]
func listMyCertificates(svc *service.CertificateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items, err := svc.ListMine(c.UserContext(), currentActor(c))
		if err != nil {
			return err
		}
		return helper.Success(c, "sertifikat saya", items)
	}
}

// @Summary     Terbitkan sertifikat (panitia event)
// @Description Menerbitkan sertifikat untuk semua peserta yang sudah check-in. Aman dipanggil berulang.
// @Tags        Certificates
// @Security    BearerAuth
// @Param       id path string true "ID event (UUID)"
// @Success     200 {object} model.WebResponse
// @Failure     403 {object} model.WebResponse
// @Router      /events/{id}/certificates/issue [post]
func issueCertificates(svc *service.CertificateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := helper.ParamID(c, "id")
		if err != nil {
			return err
		}
		res, err := svc.IssueForEvent(c.UserContext(), currentActor(c), id)
		if err != nil {
			return err
		}
		return helper.Success(c, "sertifikat diterbitkan", res)
	}
}