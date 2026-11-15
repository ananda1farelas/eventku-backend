package helper

import (
	"net/http"

	"eventku/app/model"
)

// AppError adalah error yang sudah membawa kode HTTP. Service/rules
// mengembalikan ini, lalu ErrorHandler (config/error_handler.go) yang
// mengubahnya jadi response JSON.
type AppError struct {
	Code    int
	Message string
	Details []model.ErrorDetail
}

func (e *AppError) Error() string { return e.Message }

func NewValidation(msg string, details ...model.ErrorDetail) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: msg, Details: details}
}

func NewNotFound(msg string) *AppError { return &AppError{Code: http.StatusNotFound, Message: msg} }

func NewConflict(msg string) *AppError { return &AppError{Code: http.StatusConflict, Message: msg} }

func NewForbidden(msg string) *AppError { return &AppError{Code: http.StatusForbidden, Message: msg} }

func NewUnauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg}
}