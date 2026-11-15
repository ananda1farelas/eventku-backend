package service

import (
	"strings"
	"testing"
	"time"

	"eventku/app/model"
)

func TestBuildReminderMessage(t *testing.T) {
	ev := model.Event{
		Title: "Seminar Go", Location: "Aula Kampus",
		StartAt: time.Date(2026, 11, 16, 2, 0, 0, 0, time.UTC), // 09:00 WIB
	}
	title, msg := BuildReminderMessage(model.Recipient{FullName: "Budi"}, ev)

	if title != "Pengingat: Seminar Go dimulai besok" {
		t.Errorf("title = %q", title)
	}
	for _, want := range []string{"Halo Budi", "Seminar Go", "pukul 09:00 WIB", "Aula Kampus"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message tidak memuat %q: %s", want, msg)
		}
	}
}