package helper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"eventku/app/model"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	// Pakai nama field JSON di pesan error, bukan nama field Go.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

// Validate menjalankan tag `validate:"..."` pada struct. Mengembalikan nil
// kalau lolos, atau AppError (422) berisi detail per field.
func Validate(s any) *AppError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return NewValidation("validasi gagal")
	}

	details := make([]model.ErrorDetail, 0, len(verrs))
	for _, fe := range verrs {
		details = append(details, model.ErrorDetail{Field: fe.Field(), Message: messageFor(fe)})
	}
	return NewValidation("validasi gagal", details...)
}

func messageFor(fe validator.FieldError) string {
	isText := fe.Kind() == reflect.String
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "uuid":
		return "harus berupa UUID yang valid"
	case "min":
		if isText {
			return fmt.Sprintf("minimal %s karakter", fe.Param())
		}
		return fmt.Sprintf("minimal %s", fe.Param())
	case "max":
		if isText {
			return fmt.Sprintf("maksimal %s karakter", fe.Param())
		}
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "gt":
		return fmt.Sprintf("harus lebih dari %s", fe.Param())
	case "gte":
		return fmt.Sprintf("minimal %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", fe.Param())
	default:
		return "tidak valid"
	}
}