package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
	"eventku/helper"
)

// Actor adalah siapa yang sedang melakukan request. Diisi route dari
// middleware.CurrentUserID dan middleware.CurrentRole.
type Actor struct {
	UserID uuid.UUID
	Role   string
}

func (a Actor) IsAdmin() bool { return a.Role == model.RoleAdmin }

const defaultEventCategory = "umum"

type EventService struct {
	events  *repository.EventRepository
	members OrganizationMemberChecker // diimplementasi OrganizationRepository (Orang 1)
	now     func() time.Time
}

func NewEventService(events *repository.EventRepository, members OrganizationMemberChecker) *EventService {
	return &EventService{events: events, members: members, now: time.Now}
}

// mapEventErr menerjemahkan error repository jadi error HTTP yang ramah.
func mapEventErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NewNotFound("event tidak ditemukan")
	case errors.Is(err, repository.ErrEventConflict):
		return helper.NewConflict(repository.ErrEventConflict.Error())
	default:
		return err
	}
}

// load membaca event dan status keanggotaan actor di organisasinya.
// Event yang tidak boleh dilihat actor dijawab "tidak ditemukan" (bukan 403)
// supaya keberadaan event draft tidak bocor.
func (s *EventService) load(ctx context.Context, actor Actor, id uuid.UUID) (*model.Event, bool, error) {
	ev, err := s.events.FindByID(ctx, id)
	if err != nil {
		return nil, false, mapEventErr(err)
	}
	isMember, err := s.members.IsMember(ctx, ev.OrganizationID, actor.UserID)
	if err != nil {
		return nil, false, fmt.Errorf("cek keanggotaan organisasi: %w", err)
	}
	if !CanViewEvent(ev.Status, actor.IsAdmin(), isMember) {
		return nil, false, helper.NewNotFound("event tidak ditemukan")
	}
	return ev, isMember, nil
}

func normalizeCategory(c string) string {
	c = strings.ToLower(strings.TrimSpace(c))
	if c == "" {
		return defaultEventCategory
	}
	return c
}

func inputOf(ev *model.Event) EventInput {
	return EventInput{
		Title: ev.Title, Location: ev.Location, StartAt: ev.StartAt, EndAt: ev.EndAt,
		Quota: ev.Quota, Price: ev.Price, RegistrationDeadline: ev.RegistrationDeadline,
	}
}

// Create: panitia (anggota organisasi) membuat event baru berstatus draft.
func (s *EventService) Create(ctx context.Context, actor Actor, req model.CreateEventRequest) (*model.Event, error) {
	ev := &model.Event{
		OrganizationID: req.OrganizationID, CreatedBy: actor.UserID,
		Title: strings.TrimSpace(req.Title), Description: req.Description,
		Category: normalizeCategory(req.Category), Location: strings.TrimSpace(req.Location),
		StartAt: req.StartAt, EndAt: req.EndAt, Quota: req.Quota, Price: req.Price,
		RegistrationDeadline: req.RegistrationDeadline, Status: model.EventStatusDraft,
	}
	if err := ValidateEventInput(inputOf(ev), s.now()); err != nil {
		return nil, err
	}

	isMember, err := s.members.IsMember(ctx, ev.OrganizationID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("cek keanggotaan organisasi: %w", err)
	}
	if err := CanManageEvent(isMember); err != nil {
		return nil, err
	}

	if err := s.events.Create(ctx, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

// Get: detail satu event.
func (s *EventService) Get(ctx context.Context, actor Actor, id uuid.UUID) (*model.Event, error) {
	ev, _, err := s.load(ctx, actor, id)
	return ev, err
}

// ListQuery parameter list event dari query string.
type ListEventsQuery struct {
	Page, Limit    int
	Search         string
	Category       string
	Status         string
	OrganizationID *uuid.UUID
}

// List: admin melihat semua; panitia melihat semua event organisasinya
// (kalau memfilter organization_id); selain itu hanya event publik.
func (s *EventService) List(ctx context.Context, actor Actor, q ListEventsQuery) ([]model.Event, *model.Meta, error) {
	f := model.EventFilter{
		Page: q.Page, Limit: q.Limit, Search: q.Search,
		Category: strings.ToLower(q.Category), OrganizationID: q.OrganizationID,
	}

	seeAll := actor.IsAdmin()
	if !seeAll && q.OrganizationID != nil {
		isMember, err := s.members.IsMember(ctx, *q.OrganizationID, actor.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("cek keanggotaan organisasi: %w", err)
		}
		seeAll = isMember
	}

	switch {
	case q.Status != "" && (seeAll || isPublicStatus(q.Status)):
		f.Statuses = []string{q.Status}
	case q.Status != "":
		f.Statuses = []string{} // status privat diminta orang luar: hasil kosong
	case !seeAll:
		f.Statuses = publicEventStatuses
	}

	events, total, err := s.events.List(ctx, f)
	if err != nil {
		return nil, nil, err
	}
	return events, helper.NewMeta(q.Page, q.Limit, total), nil
}

// Update: panitia mengubah isi event (hanya draft/rejected).
func (s *EventService) Update(ctx context.Context, actor Actor, id uuid.UUID, req model.UpdateEventRequest) (*model.Event, error) {
	ev, isMember, err := s.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if err := CanManageEvent(isMember); err != nil {
		return nil, err
	}
	if err := CanEditEvent(ev.Status); err != nil {
		return nil, err
	}

	previousStatus := ev.Status
	ev.Title, ev.Description = strings.TrimSpace(req.Title), req.Description
	ev.Category, ev.Location = normalizeCategory(req.Category), strings.TrimSpace(req.Location)
	ev.StartAt, ev.EndAt, ev.Quota, ev.Price = req.StartAt, req.EndAt, req.Quota, req.Price
	ev.RegistrationDeadline = req.RegistrationDeadline

	if err := ValidateEventInput(inputOf(ev), s.now()); err != nil {
		return nil, err
	}
	if err := s.events.Update(ctx, ev, previousStatus); err != nil {
		return nil, mapEventErr(err)
	}
	return ev, nil
}

// Delete: soft delete, hanya draft/rejected.
func (s *EventService) Delete(ctx context.Context, actor Actor, id uuid.UUID) error {
	ev, isMember, err := s.load(ctx, actor, id)
	if err != nil {
		return err
	}
	if err := CanManageEvent(isMember); err != nil {
		return err
	}
	if err := CanEditEvent(ev.Status); err != nil {
		return err
	}
	return mapEventErr(s.events.SoftDelete(ctx, ev.ID, ev.Status))
}

// Submit: panitia mengajukan event ke admin (draft/rejected -> pending_review).
func (s *EventService) Submit(ctx context.Context, actor Actor, id uuid.UUID) (*model.Event, error) {
	ev, isMember, err := s.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if err := CanManageEvent(isMember); err != nil {
		return nil, err
	}
	if err := CanTransition(ev.Status, model.EventStatusPendingReview); err != nil {
		return nil, err
	}
	// Jadwal bisa saja sudah lewat sejak draft dibuat.
	if err := ValidateEventInput(inputOf(ev), s.now()); err != nil {
		return nil, err
	}
	return s.changeStatus(ctx, ev, model.EventStatusPendingReview, nil)
}

// Approve: admin menyetujui (pending_review -> open).
func (s *EventService) Approve(ctx context.Context, actor Actor, id uuid.UUID) (*model.Event, error) {
	if err := CanModerateEvent(actor.Role); err != nil {
		return nil, err
	}
	ev, _, err := s.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if err := CanTransition(ev.Status, model.EventStatusOpen); err != nil {
		return nil, err
	}
	return s.changeStatus(ctx, ev, model.EventStatusOpen, nil)
}

// Reject: admin menolak disertai alasan (pending_review -> rejected).
func (s *EventService) Reject(ctx context.Context, actor Actor, id uuid.UUID, reason string) (*model.Event, error) {
	if err := CanModerateEvent(actor.Role); err != nil {
		return nil, err
	}
	if err := ValidateRejectReason(reason); err != nil {
		return nil, err
	}
	ev, _, err := s.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if err := CanTransition(ev.Status, model.EventStatusRejected); err != nil {
		return nil, err
	}
	reason = strings.TrimSpace(reason)
	return s.changeStatus(ctx, ev, model.EventStatusRejected, &reason)
}

func (s *EventService) changeStatus(ctx context.Context, ev *model.Event, to string, reason *string) (*model.Event, error) {
	if err := s.events.UpdateStatus(ctx, ev.ID, ev.Status, to, reason); err != nil {
		return nil, mapEventErr(err)
	}
	ev.Status = to
	ev.RejectionReason = reason
	ev.UpdatedAt = s.now()
	return ev, nil
}