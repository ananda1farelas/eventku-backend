package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
	"eventku/helper"
)

type ReportService struct {
	events  *repository.EventRepository
	reports *repository.ReportRepository
	members OrganizationMemberChecker
}

func NewReportService(events *repository.EventRepository, reports *repository.ReportRepository, members OrganizationMemberChecker) *ReportService {
	return &ReportService{events: events, reports: reports, members: members}
}

var validParticipantStatuses = map[string]bool{
	"": true, model.RegistrationStatusRegistered: true,
	model.RegistrationStatusWaitlist: true, model.RegistrationStatusCancelled: true,
}

func (s *ReportService) authorize(ctx context.Context, actor Actor, eventID uuid.UUID) (*model.Event, error) {
	ev, err := s.events.FindByID(ctx, eventID)
	if err != nil {
		return nil, mapEventErr(err)
	}
	isMember, err := s.members.IsMember(ctx, ev.OrganizationID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("cek keanggotaan organisasi: %w", err)
	}
	// Peserta biasa dijawab "tidak ditemukan" untuk event non-publik.
	if !CanViewEvent(ev.Status, actor.IsAdmin(), isMember) {
		return nil, helper.NewNotFound("event tidak ditemukan")
	}
	if err := CanViewParticipants(actor.IsAdmin(), isMember); err != nil {
		return nil, err
	}
	return ev, nil
}

// Participants untuk GET /events/{id}/participants. status opsional.
func (s *ReportService) Participants(ctx context.Context, actor Actor, eventID uuid.UUID, status string) ([]model.ParticipantRow, error) {
	if !validParticipantStatuses[status] {
		return nil, helper.NewValidation("status harus registered, waitlist, atau cancelled")
	}
	if _, err := s.authorize(ctx, actor, eventID); err != nil {
		return nil, err
	}
	return s.reports.ListParticipants(ctx, eventID, status)
}

// ExportCSV untuk GET /events/{id}/participants/export.
// Mengembalikan nama file dan isi CSV.
func (s *ReportService) ExportCSV(ctx context.Context, actor Actor, eventID uuid.UUID) (string, []byte, error) {
	rows, err := s.Participants(ctx, actor, eventID, "")
	if err != nil {
		return "", nil, err
	}
	data, err := BuildParticipantsCSV(rows)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("peserta-%s.csv", eventID), data, nil
}

// BuildParticipantsCSV menyusun CSV rekap peserta (fungsi murni, mudah dites).
func BuildParticipantsCSV(rows []model.ParticipantRow) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF") // BOM supaya Excel membaca UTF-8 dengan benar

	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"No", "Nama", "Email", "Status", "Waktu Daftar", "Hadir", "Waktu Check-in"}); err != nil {
		return nil, err
	}
	for i, p := range rows {
		attended, checkedIn := "Tidak", ""
		if p.Attended {
			attended = "Ya"
		}
		if p.CheckedInAt != nil {
			checkedIn = p.CheckedInAt.Format(time.RFC3339)
		}
		rec := []string{
			fmt.Sprint(i + 1), csvSafe(p.FullName), csvSafe(p.Email), p.Status,
			p.RegisteredAt.Format(time.RFC3339), attended, checkedIn,
		}
		if err := w.Write(rec); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// csvSafe mencegah CSV injection: sel yang diawali = + - @ dianggap rumus
// oleh Excel, jadi diberi tanda petik di depan.
func csvSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}