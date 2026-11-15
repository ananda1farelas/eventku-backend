package helper

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ParamID membaca path param bertipe UUID, misal ParamID(c, "id").
func ParamID(c *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, NewValidation("parameter " + name + " harus berupa UUID yang valid")
	}
	return id, nil
}

// BindAndValidate membaca body JSON ke dst lalu menjalankan validasi tag.
// Kembalian error sudah berupa AppError, jadi handler cukup `return err`.
func BindAndValidate(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return NewValidation("body request tidak valid (pastikan JSON dan tipe datanya benar)")
	}
	if ae := Validate(dst); ae != nil {
		return ae
	}
	return nil
}

// ListQuery adalah parameter umum endpoint list.
type ListQuery struct {
	Page   int
	Limit  int
	Search string
}

const (
	defaultLimit = 10
	maxLimit     = 100
)

// ParseListQuery membaca ?page=&limit=&search= dengan batas aman.
func ParseListQuery(c *fiber.Ctx) ListQuery {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", defaultLimit)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > maxLimit {
		limit = defaultLimit
	}
	return ListQuery{Page: page, Limit: limit, Search: strings.TrimSpace(c.Query("search"))}
}

func (q ListQuery) Offset() int { return (q.Page - 1) * q.Limit }