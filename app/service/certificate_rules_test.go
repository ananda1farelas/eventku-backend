package service

import (
	"errors"
	"testing"
	"time"

	"eventku/app/model"
)

func TestCanIssueCertificates(t *testing.T) {
	start := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	after := start.Add(3 * time.Hour)
	before := start.Add(-time.Hour)

	tests := []struct {
		name     string
		status   string
		isMember bool
		now      time.Time
		wantErr  error
	}{
		{"panitia, event done, sudah mulai", model.EventStatusDone, true, after, nil},
		{"panitia, event open, sudah mulai", model.EventStatusOpen, true, after, nil},
		{"bukan panitia", model.EventStatusDone, false, after, ErrCertificateForbidden},
		{"event belum mulai", model.EventStatusOpen, true, before, ErrCertificateEventNotStarted},
		{"event masih draft", model.EventStatusDraft, true, after, ErrCertificateEventStatus},
		{"event rejected", model.EventStatusRejected, true, after, ErrCertificateEventStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := model.Event{Status: tt.status, StartAt: start}
			err := CanIssueCertificates(ev, tt.isMember, tt.now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanIssueCertificates() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestEligibleForCertificate(t *testing.T) {
	tests := []struct {
		name    string
		c       model.CertificateCandidate
		wantErr error
	}{
		{"hadir dan belum punya", model.CertificateCandidate{Attended: true}, nil},
		{"tidak hadir", model.CertificateCandidate{Attended: false}, ErrCertificateNotAttended},
		{"hadir tapi sudah punya", model.CertificateCandidate{Attended: true, HasCertificate: true}, ErrCertificateAlreadyIssued},
		{"tidak hadir dan sudah punya", model.CertificateCandidate{HasCertificate: true}, ErrCertificateNotAttended},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EligibleForCertificate(tt.c); !errors.Is(err, tt.wantErr) {
				t.Fatalf("EligibleForCertificate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatCertificateNo(t *testing.T) {
	tests := []struct {
		eventNo, year, seq int
		want               string
	}{
		{12, 2026, 7, "EVT-12-2026-0007"},
		{1, 2026, 1, "EVT-1-2026-0001"},
		{5, 2027, 12345, "EVT-5-2027-12345"},
	}
	for _, tt := range tests {
		if got := FormatCertificateNo(tt.eventNo, tt.year, tt.seq); got != tt.want {
			t.Errorf("FormatCertificateNo(%d,%d,%d) = %q, want %q", tt.eventNo, tt.year, tt.seq, got, tt.want)
		}
	}
}