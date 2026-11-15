package service

import (
	"errors"
	"testing"
	"time"

	"eventku/app/model"
)

func validInput(now time.Time) EventInput {
	return EventInput{
		Title:                "Seminar Go",
		Location:             "Aula Kampus",
		StartAt:              now.Add(48 * time.Hour),
		EndAt:                now.Add(50 * time.Hour),
		Quota:                50,
		Price:                0,
		RegistrationDeadline: now.Add(24 * time.Hour),
	}
}

func TestValidateEventInput(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		mutate  func(in *EventInput)
		wantErr error
	}{
		{"valid", func(in *EventInput) {}, nil},
		{"judul terlalu pendek", func(in *EventInput) { in.Title = "ab" }, ErrEventTitleTooShort},
		{"judul cuma spasi", func(in *EventInput) { in.Title = "    " }, ErrEventTitleTooShort},
		{"lokasi kosong", func(in *EventInput) { in.Location = " " }, ErrEventLocationRequired},
		{"selesai sebelum mulai", func(in *EventInput) { in.EndAt = in.StartAt.Add(-time.Hour) }, ErrEventScheduleInvalid},
		{"selesai sama dengan mulai", func(in *EventInput) { in.EndAt = in.StartAt }, ErrEventScheduleInvalid},
		{"mulai di masa lalu", func(in *EventInput) {
			in.StartAt = now.Add(-2 * time.Hour)
			in.EndAt = now.Add(-time.Hour)
			in.RegistrationDeadline = now.Add(-3 * time.Hour)
		}, ErrEventStartInPast},
		{"kuota nol", func(in *EventInput) { in.Quota = 0 }, ErrEventQuotaInvalid},
		{"kuota negatif", func(in *EventInput) { in.Quota = -5 }, ErrEventQuotaInvalid},
		{"harga negatif", func(in *EventInput) { in.Price = -1 }, ErrEventPriceInvalid},
		{"harga berbayar", func(in *EventInput) { in.Price = 25000 }, nil},
		{"deadline sama dengan mulai", func(in *EventInput) { in.RegistrationDeadline = in.StartAt }, ErrEventDeadlineInvalid},
		{"deadline setelah mulai", func(in *EventInput) { in.RegistrationDeadline = in.StartAt.Add(time.Hour) }, ErrEventDeadlineInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput(now)
			tt.mutate(&in)
			err := ValidateEventInput(in, now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateEventInput() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRejectReason(t *testing.T) {
	if err := ValidateRejectReason("   "); !errors.Is(err, ErrEventRejectReasonEmpty) {
		t.Fatalf("alasan kosong: got %v", err)
	}
	if err := ValidateRejectReason("Deskripsi kurang jelas"); err != nil {
		t.Fatalf("alasan valid: got %v", err)
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to string
		wantErr  error
	}{
		{model.EventStatusDraft, model.EventStatusPendingReview, nil},
		{model.EventStatusRejected, model.EventStatusPendingReview, nil},
		{model.EventStatusPendingReview, model.EventStatusOpen, nil},
		{model.EventStatusPendingReview, model.EventStatusRejected, nil},
		{model.EventStatusOpen, model.EventStatusClosed, nil},
		{model.EventStatusOpen, model.EventStatusDone, nil},
		{model.EventStatusClosed, model.EventStatusDone, nil},
		{model.EventStatusDraft, model.EventStatusOpen, ErrEventInvalidTransition},
		{model.EventStatusDraft, model.EventStatusDone, ErrEventInvalidTransition},
		{model.EventStatusOpen, model.EventStatusDraft, ErrEventInvalidTransition},
		{model.EventStatusDone, model.EventStatusOpen, ErrEventInvalidTransition},
		{model.EventStatusRejected, model.EventStatusOpen, ErrEventInvalidTransition},
		{"status_ngawur", model.EventStatusOpen, ErrEventInvalidTransition},
	}
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			if err := CanTransition(tt.from, tt.to); !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanTransition() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCanEditEvent(t *testing.T) {
	editable := []string{model.EventStatusDraft, model.EventStatusRejected}
	locked := []string{model.EventStatusPendingReview, model.EventStatusOpen, model.EventStatusClosed, model.EventStatusDone}

	for _, s := range editable {
		if err := CanEditEvent(s); err != nil {
			t.Errorf("status %q harusnya boleh diedit, got %v", s, err)
		}
	}
	for _, s := range locked {
		if err := CanEditEvent(s); !errors.Is(err, ErrEventNotEditable) {
			t.Errorf("status %q harusnya terkunci, got %v", s, err)
		}
	}
}

func TestPermissionRules(t *testing.T) {
	if err := CanManageEvent(true); err != nil {
		t.Errorf("anggota harusnya boleh: %v", err)
	}
	if err := CanManageEvent(false); !errors.Is(err, ErrEventForbidden) {
		t.Errorf("non-anggota harusnya ditolak: %v", err)
	}
	if err := CanModerateEvent(model.RoleAdmin); err != nil {
		t.Errorf("admin harusnya boleh moderasi: %v", err)
	}
	if err := CanModerateEvent(model.RoleUser); !errors.Is(err, ErrEventModerateForbidden) {
		t.Errorf("user biasa harusnya ditolak: %v", err)
	}
	if err := CanViewParticipants(false, false); !errors.Is(err, ErrEventForbidden) {
		t.Errorf("orang luar harusnya ditolak: %v", err)
	}
	if err := CanViewParticipants(true, false); err != nil {
		t.Errorf("admin harusnya boleh: %v", err)
	}
	if err := CanViewParticipants(false, true); err != nil {
		t.Errorf("panitia harusnya boleh: %v", err)
	}
}

func TestCanViewEvent(t *testing.T) {
	tests := []struct {
		name              string
		status            string
		isAdmin, isMember bool
		want              bool
	}{
		{"open untuk umum", model.EventStatusOpen, false, false, true},
		{"done untuk umum", model.EventStatusDone, false, false, true},
		{"draft disembunyikan dari umum", model.EventStatusDraft, false, false, false},
		{"pending disembunyikan dari umum", model.EventStatusPendingReview, false, false, false},
		{"rejected disembunyikan dari umum", model.EventStatusRejected, false, false, false},
		{"draft terlihat oleh panitia", model.EventStatusDraft, false, true, true},
		{"pending terlihat oleh admin", model.EventStatusPendingReview, true, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanViewEvent(tt.status, tt.isAdmin, tt.isMember); got != tt.want {
				t.Fatalf("CanViewEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}