package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"eventku/app/model"
)

func TestValidatePaymentMethod(t *testing.T) {
	for _, m := range []string{"transfer", "ewallet", "qris"} {
		if err := ValidatePaymentMethod(m); err != nil {
			t.Errorf("metode %q harusnya valid: %v", m, err)
		}
	}
	for _, m := range []string{"", "cash", "TRANSFER", "kartu"} {
		if err := ValidatePaymentMethod(m); !errors.Is(err, ErrPaymentMethodInvalid) {
			t.Errorf("metode %q harusnya ditolak, got %v", m, err)
		}
	}
}

func TestCanCreatePayment(t *testing.T) {
	owner, other := uuid.New(), uuid.New()
	base := model.PayableRegistration{UserID: owner, Status: model.RegistrationStatusRegistered, Price: 50000}

	tests := []struct {
		name    string
		mutate  func(r *model.PayableRegistration)
		actor   uuid.UUID
		wantErr error
	}{
		{"pemilik, event berbayar", func(r *model.PayableRegistration) {}, owner, nil},
		{"bukan pemilik", func(r *model.PayableRegistration) {}, other, ErrPaymentRegistrationNotFound},
		{"waitlist belum bayar", func(r *model.PayableRegistration) { r.Status = model.RegistrationStatusWaitlist }, owner, ErrPaymentRegistrationInactive},
		{"sudah dibatalkan", func(r *model.PayableRegistration) { r.Status = model.RegistrationStatusCancelled }, owner, ErrPaymentRegistrationInactive},
		{"event gratis", func(r *model.PayableRegistration) { r.Price = 0 }, owner, ErrPaymentEventFree},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := base
			tt.mutate(&reg)
			if err := CanCreatePayment(reg, tt.actor); !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanCreatePayment() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCanAccessPayment(t *testing.T) {
	owner, other := uuid.New(), uuid.New()
	if err := CanAccessPayment(owner, owner, false); err != nil {
		t.Errorf("pemilik harusnya boleh: %v", err)
	}
	if err := CanAccessPayment(owner, other, true); err != nil {
		t.Errorf("admin harusnya boleh: %v", err)
	}
	if err := CanAccessPayment(owner, other, false); !errors.Is(err, ErrPaymentNotFound) {
		t.Errorf("orang lain harusnya ditolak: %v", err)
	}
}

func TestCanTransitionPayment(t *testing.T) {
	tests := []struct {
		from, to string
		wantErr  error
	}{
		{model.PaymentStatusPending, model.PaymentStatusPaid, nil},
		{model.PaymentStatusPending, model.PaymentStatusFailed, nil},
		{model.PaymentStatusPaid, model.PaymentStatusFailed, ErrPaymentInvalidTransition},
		{model.PaymentStatusPaid, model.PaymentStatusPending, ErrPaymentInvalidTransition},
		{model.PaymentStatusFailed, model.PaymentStatusPaid, ErrPaymentInvalidTransition},
		{model.PaymentStatusPending, model.PaymentStatusPending, ErrPaymentInvalidTransition},
	}
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			if err := CanTransitionPayment(tt.from, tt.to); !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanTransitionPayment() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}