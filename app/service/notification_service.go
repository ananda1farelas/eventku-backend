package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
	"eventku/helper"
)

type NotificationService struct {
	notifications *repository.NotificationRepository
}

func NewNotificationService(notifications *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

// ListMine untuk GET /me/notifications.
func (s *NotificationService) ListMine(ctx context.Context, actor Actor, page, limit int, unreadOnly bool) ([]model.Notification, *model.Meta, error) {
	items, total, err := s.notifications.ListByUser(ctx, actor.UserID, unreadOnly, limit, (page-1)*limit)
	if err != nil {
		return nil, nil, err
	}
	return items, helper.NewMeta(page, limit, total), nil
}

// MarkRead untuk PATCH /me/notifications/{id}/read.
func (s *NotificationService) MarkRead(ctx context.Context, actor Actor, id uuid.UUID) error {
	if err := s.notifications.MarkRead(ctx, id, actor.UserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NewNotFound("notifikasi tidak ditemukan")
		}
		return err
	}
	return nil
}

// DBNotifier adalah Notifier yang menyimpan pengingat ke tabel notifications
// (muncul di GET /me/notifications) sekaligus menulisnya ke log sebagai
// "pengiriman" palsu. Pakai ini menggantikan LogNotifier di main.go.
type DBNotifier struct {
	Repo   *repository.NotificationRepository
	Logger *log.Logger
}

func (n DBNotifier) Notify(ctx context.Context, to model.Recipient, ev model.Event) error {
	title, message := BuildReminderMessage(to, ev)
	evID := ev.ID
	created, err := n.Repo.Create(ctx, &model.Notification{
		UserID: to.UserID, EventID: &evID,
		Type: model.NotificationTypeReminderH1, Title: title, Message: message,
	})
	if err != nil {
		return err
	}
	if created {
		n.Logger.Printf("[reminder H-1] terkirim ke %s <%s>: %s", to.FullName, to.Email, title)
	}
	return nil
}