package service

import (
	"strings"
	"time"

	"eventku/app/model"
	"eventku/helper"
)

// Error domain event. Dibuat sebagai *AppError supaya langsung punya kode
// HTTP, dan tetap bisa dicek dengan errors.Is di test.
var (
	ErrEventTitleTooShort     = helper.NewValidation("judul event minimal 3 karakter")
	ErrEventLocationRequired  = helper.NewValidation("lokasi event wajib diisi")
	ErrEventScheduleInvalid   = helper.NewValidation("waktu selesai harus setelah waktu mulai")
	ErrEventStartInPast       = helper.NewValidation("waktu mulai event harus di masa depan")
	ErrEventQuotaInvalid      = helper.NewValidation("kuota harus lebih dari 0")
	ErrEventPriceInvalid      = helper.NewValidation("harga tidak boleh negatif (isi 0 untuk gratis)")
	ErrEventDeadlineInvalid   = helper.NewValidation("batas pendaftaran harus sebelum waktu mulai event")
	ErrEventRejectReasonEmpty = helper.NewValidation("alasan penolakan wajib diisi")

	ErrEventForbidden         = helper.NewForbidden("hanya panitia organisasi pemilik event yang boleh melakukan ini")
	ErrEventModerateForbidden = helper.NewForbidden("hanya admin yang boleh menyetujui atau menolak event")
	ErrEventNotEditable       = helper.NewConflict("event hanya bisa diubah atau dihapus saat berstatus draft atau rejected")
	ErrEventInvalidTransition = helper.NewConflict("perpindahan status event tidak diizinkan")
)

// EventInput adalah isi event yang divalidasi (dipakai create dan update).
type EventInput struct {
	Title                string
	Location             string
	StartAt              time.Time
	EndAt                time.Time
	Quota                int
	Price                float64
	RegistrationDeadline time.Time
}

// ValidateEventInput mengecek aturan bisnis isi event. `now` dikirim dari
// luar supaya mudah dites.
func ValidateEventInput(in EventInput, now time.Time) error {
	if len(strings.TrimSpace(in.Title)) < 3 {
		return ErrEventTitleTooShort
	}
	if strings.TrimSpace(in.Location) == "" {
		return ErrEventLocationRequired
	}
	if !in.EndAt.After(in.StartAt) {
		return ErrEventScheduleInvalid
	}
	if !in.StartAt.After(now) {
		return ErrEventStartInPast
	}
	if in.Quota <= 0 {
		return ErrEventQuotaInvalid
	}
	if in.Price < 0 {
		return ErrEventPriceInvalid
	}
	if !in.RegistrationDeadline.Before(in.StartAt) {
		return ErrEventDeadlineInvalid
	}
	return nil
}

// ValidateRejectReason memastikan penolakan disertai alasan.
func ValidateRejectReason(reason string) error {
	if len(strings.TrimSpace(reason)) < 5 {
		return ErrEventRejectReasonEmpty
	}
	return nil
}

// Aturan perpindahan status. Kunci = status asal, nilai = status tujuan sah.
var eventTransitions = map[string][]string{
	model.EventStatusDraft:         {model.EventStatusPendingReview},
	model.EventStatusRejected:      {model.EventStatusPendingReview},
	model.EventStatusPendingReview: {model.EventStatusOpen, model.EventStatusRejected},
	model.EventStatusOpen:          {model.EventStatusClosed, model.EventStatusDone},
	model.EventStatusClosed:        {model.EventStatusDone},
}

// CanTransition mengecek apakah status boleh berpindah dari -> ke.
func CanTransition(from, to string) error {
	for _, allowed := range eventTransitions[from] {
		if allowed == to {
			return nil
		}
	}
	return ErrEventInvalidTransition
}

// CanEditEvent: isi event hanya boleh diubah/dihapus saat draft atau rejected.
func CanEditEvent(status string) error {
	if status == model.EventStatusDraft || status == model.EventStatusRejected {
		return nil
	}
	return ErrEventNotEditable
}

// CanManageEvent: hanya anggota (panitia) organisasi pemilik event.
func CanManageEvent(isMember bool) error {
	if !isMember {
		return ErrEventForbidden
	}
	return nil
}

// CanModerateEvent: hanya admin.
func CanModerateEvent(role string) error {
	if role != model.RoleAdmin {
		return ErrEventModerateForbidden
	}
	return nil
}

// CanViewParticipants: rekap peserta untuk admin atau panitia event itu.
func CanViewParticipants(isAdmin, isMember bool) error {
	if !isAdmin && !isMember {
		return ErrEventForbidden
	}
	return nil
}

// publicEventStatuses adalah status yang boleh dilihat semua user login.
var publicEventStatuses = []string{model.EventStatusOpen, model.EventStatusClosed, model.EventStatusDone}

func isPublicStatus(status string) bool {
	for _, s := range publicEventStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// CanViewEvent: event yang belum disetujui hanya terlihat oleh admin dan
// panitia pemiliknya.
func CanViewEvent(status string, isAdmin, isMember bool) bool {
	return isAdmin || isMember || isPublicStatus(status)
}