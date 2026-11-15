package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
	"eventku/helper"
)

// PaymentService mengurus pembayaran MOCK: tidak ada gateway, status diubah
// lewat endpoint pay/fail untuk mensimulasikan hasil pembayaran.
type PaymentService struct {
	payments *repository.PaymentRepository
	now      func() time.Time
}

func NewPaymentService(payments *repository.PaymentRepository) *PaymentService {
	return &PaymentService{payments: payments, now: time.Now}
}

func mapPaymentErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrPaymentNotFound
	case errors.Is(err, repository.ErrPaymentExists):
		return helper.NewConflict(repository.ErrPaymentExists.Error())
	case errors.Is(err, repository.ErrPaymentConflict):
		return helper.NewConflict(repository.ErrPaymentConflict.Error())
	default:
		return err
	}
}

// Create membuat tagihan (pending) senilai harga event.
func (s *PaymentService) Create(ctx context.Context, actor Actor, registrationID uuid.UUID, req model.CreatePaymentRequest) (*model.Payment, error) {
	if err := ValidatePaymentMethod(req.Method); err != nil {
		return nil, err
	}

	reg, err := s.payments.FindPayable(ctx, registrationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrPaymentRegistrationNotFound
		}
		return nil, err
	}
	if err := CanCreatePayment(*reg, actor.UserID); err != nil {
		return nil, err
	}

	p := &model.Payment{
		RegistrationID: reg.RegistrationID, Amount: reg.Price,
		Method: req.Method, Status: model.PaymentStatusPending,
	}
	if err := s.payments.Create(ctx, p); err != nil {
		return nil, mapPaymentErr(err)
	}
	return p, nil
}

func (s *PaymentService) load(ctx context.Context, actor Actor, id uuid.UUID) (*model.PaymentDetail, error) {
	d, err := s.payments.FindByID(ctx, id)
	if err != nil {
		return nil, mapPaymentErr(err)
	}
	if err := CanAccessPayment(d.UserID, actor.UserID, actor.IsAdmin()); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *PaymentService) Get(ctx context.Context, actor Actor, id uuid.UUID) (*model.PaymentDetail, error) {
	return s.load(ctx, actor, id)
}

func (s *PaymentService) ListMine(ctx context.Context, actor Actor) ([]model.PaymentDetail, error) {
	return s.payments.ListByUser(ctx, actor.UserID)
}

// Pay (mock): pending -> paid.
func (s *PaymentService) Pay(ctx context.Context, actor Actor, id uuid.UUID) (*model.PaymentDetail, error) {
	return s.transition(ctx, actor, id, model.PaymentStatusPaid)
}

// Fail (mock): pending -> failed. Setelah failed, peserta boleh membuat tagihan baru.
func (s *PaymentService) Fail(ctx context.Context, actor Actor, id uuid.UUID) (*model.PaymentDetail, error) {
	return s.transition(ctx, actor, id, model.PaymentStatusFailed)
}

func (s *PaymentService) transition(ctx context.Context, actor Actor, id uuid.UUID, to string) (*model.PaymentDetail, error) {
	d, err := s.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if err := CanTransitionPayment(d.Status, to); err != nil {
		return nil, err
	}

	var paidAt *time.Time
	if to == model.PaymentStatusPaid {
		now := s.now()
		paidAt = &now
	}
	if err := s.payments.UpdateStatus(ctx, d.ID, d.Status, to, paidAt); err != nil {
		return nil, mapPaymentErr(err)
	}
	d.Status, d.PaidAt = to, paidAt
	return d, nil
}