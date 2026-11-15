package service

import (
	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/helper"
)

var (
	ErrPaymentMethodInvalid        = helper.NewValidation("metode pembayaran harus transfer, ewallet, atau qris")
	ErrPaymentRegistrationNotFound = helper.NewNotFound("registrasi tidak ditemukan")
	ErrPaymentRegistrationInactive = helper.NewConflict("hanya registrasi berstatus registered yang bisa dibayar")
	ErrPaymentEventFree            = helper.NewConflict("event ini gratis, tidak perlu pembayaran")
	ErrPaymentNotFound             = helper.NewNotFound("pembayaran tidak ditemukan")
	ErrPaymentInvalidTransition    = helper.NewConflict("pembayaran hanya bisa diubah selama berstatus pending")
)

// ValidatePaymentMethod: metode yang didukung mock.
func ValidatePaymentMethod(method string) error {
	switch method {
	case model.PaymentMethodTransfer, model.PaymentMethodEwallet, model.PaymentMethodQRIS:
		return nil
	}
	return ErrPaymentMethodInvalid
}

// CanCreatePayment: hanya pemilik registrasi, registrasi harus registered
// (peserta waitlist belum perlu bayar), dan event harus berbayar.
// Registrasi milik orang lain dijawab "tidak ditemukan" supaya tidak bocor.
func CanCreatePayment(reg model.PayableRegistration, actorID uuid.UUID) error {
	if reg.UserID != actorID {
		return ErrPaymentRegistrationNotFound
	}
	if reg.Status != model.RegistrationStatusRegistered {
		return ErrPaymentRegistrationInactive
	}
	if reg.Price <= 0 {
		return ErrPaymentEventFree
	}
	return nil
}

// CanAccessPayment: pemilik pembayaran atau admin.
func CanAccessPayment(ownerID, actorID uuid.UUID, isAdmin bool) error {
	if isAdmin || ownerID == actorID {
		return nil
	}
	return ErrPaymentNotFound
}

// CanTransitionPayment: hanya pending -> paid atau pending -> failed.
func CanTransitionPayment(from, to string) error {
	if from == model.PaymentStatusPending &&
		(to == model.PaymentStatusPaid || to == model.PaymentStatusFailed) {
		return nil
	}
	return ErrPaymentInvalidTransition
}