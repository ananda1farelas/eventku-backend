package service

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
)

func TestBuildParticipantsCSV(t *testing.T) {
	in := time.Date(2026, 11, 20, 9, 5, 0, 0, time.UTC)
	reg := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)

	rows := []model.ParticipantRow{
		{RegistrationID: uuid.New(), FullName: "Budi, S.Kom", Email: "budi@kampus.ac.id",
			Status: "registered", RegisteredAt: reg, Attended: true, CheckedInAt: &in},
		{RegistrationID: uuid.New(), FullName: "=HYPERLINK(\"x\")", Email: "sari@kampus.ac.id",
			Status: "waitlist", RegisteredAt: reg},
	}

	data, err := BuildParticipantsCSV(rows)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)

	if !strings.HasPrefix(out, "\xEF\xBB\xBFNo,Nama,Email,Status") {
		t.Errorf("header/BOM salah: %q", out)
	}
	if !strings.Contains(out, `"Budi, S.Kom"`) {
		t.Errorf("koma di dalam nama harus di-quote: %q", out)
	}
	if !strings.Contains(out, ",Ya,2026-11-20T09:05:00Z") {
		t.Errorf("data kehadiran tidak muncul: %q", out)
	}
	if strings.Contains(out, ",=HYPERLINK") || !strings.Contains(out, "'=HYPERLINK") {
		t.Errorf("CSV injection belum dinetralkan: %q", out)
	}
	if got := strings.Count(out, "\n"); got != 3 {
		t.Errorf("jumlah baris = %d, want 3 (header + 2 peserta)", got)
	}
}

func TestBuildParticipantsCSVEmpty(t *testing.T) {
	data, err := BuildParticipantsCSV(nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "\n") != 1 {
		t.Errorf("CSV kosong harus berisi header saja: %q", data)
	}
}