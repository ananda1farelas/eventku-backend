package helper

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
)

func TestValidateCreateEventRequest(t *testing.T) {
	valid := model.CreateEventRequest{
		OrganizationID: uuid.New(), Title: "Seminar Go", Location: "Aula",
		StartAt: time.Now().Add(48 * time.Hour), EndAt: time.Now().Add(50 * time.Hour),
		Quota: 10, RegistrationDeadline: time.Now().Add(24 * time.Hour),
	}
	if ae := Validate(valid); ae != nil {
		t.Fatalf("request valid ditolak: %+v", ae.Details)
	}

	bad := model.CreateEventRequest{Title: "ab", Quota: 0, Price: -1}
	ae := Validate(bad)
	if ae == nil {
		t.Fatal("request kosong/ngawur harusnya ditolak")
	}

	got := map[string]string{}
	for _, d := range ae.Details {
		got[d.Field] = d.Message
	}
	for _, field := range []string{"organization_id", "title", "location", "start_at", "end_at", "quota", "price", "registration_deadline"} {
		if _, ok := got[field]; !ok {
			t.Errorf("field %q harusnya gagal validasi, details: %v", field, got)
		}
	}
	if got["title"] != "minimal 3 karakter" {
		t.Errorf("pesan title = %q", got["title"])
	}
}

func TestResolveError(t *testing.T) {
	custom := errors.New("error kustom")
	RegisterError(custom, 409)

	if ae := ResolveError(NewNotFound("x")); ae == nil || ae.Code != 404 {
		t.Errorf("AppError harus dikembalikan apa adanya: %+v", ae)
	}
	wrapped := errors.Join(errors.New("konteks"), custom)
	if ae := ResolveError(wrapped); ae == nil || ae.Code != 409 {
		t.Errorf("error terdaftar (dibungkus) harus jadi 409: %+v", ae)
	}
	if ae := ResolveError(errors.New("tidak dikenal")); ae != nil {
		t.Errorf("error tak dikenal harus nil (jadi 500): %+v", ae)
	}
}

func TestNewMeta(t *testing.T) {
	tests := []struct {
		page, limit int
		total       int64
		wantPages   int
	}{{1, 10, 0, 0}, {1, 10, 10, 1}, {1, 10, 11, 2}, {3, 5, 23, 5}}
	for _, tt := range tests {
		if got := NewMeta(tt.page, tt.limit, tt.total).TotalPages; got != tt.wantPages {
			t.Errorf("NewMeta(%d,%d,%d).TotalPages = %d, want %d", tt.page, tt.limit, tt.total, got, tt.wantPages)
		}
	}
}