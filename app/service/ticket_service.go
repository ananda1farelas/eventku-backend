package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
	"eventku/helper"
)

type TicketService struct {
	tickets *repository.TicketRepository
}

func NewTicketService(tickets *repository.TicketRepository) *TicketService {
	return &TicketService{tickets: tickets}
}

// listMine untuk GET /me/tickets.
func (s *TicketService) listMine(ctx context.Context, userID uuid.UUID) ([]model.TicketDetail, error) {
	return s.tickets.ListByUser(ctx, userID)
}

func (s *TicketService) getForUser(ctx context.Context, ticketID, userID uuid.UUID) (*model.Ticket, error) {
	return s.tickets.FindByIDForUser(ctx, ticketID, userID)
}

func issueTicket(ctx context.Context, tickets *repository.TicketRepository, registrationID uuid.UUID) error {
	_, err := tickets.FindByRegistrationID(ctx, registrationID)
	if err == nil {
		return nil // sudah punya tiket
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	token, err := helper.NewQRToken()
	if err != nil {
		return err
	}

	return tickets.Create(ctx, &model.Ticket{
		RegistrationID: registrationID,
		QRCode:         token,
	})
}
