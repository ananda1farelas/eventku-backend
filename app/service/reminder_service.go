package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"eventku/app/model"
	"eventku/app/repository"
)

// Notifier mengirim satu pengingat. Implementasi nyata (email/WhatsApp)
// tinggal menggantikan LogNotifier tanpa mengubah ReminderService.
type Notifier interface {
	Notify(ctx context.Context, to model.Recipient, ev model.Event) error
}

// LogNotifier adalah versi mock: pengingat hanya ditulis ke log.
type LogNotifier struct{ Logger *log.Logger }

func (n LogNotifier) Notify(_ context.Context, to model.Recipient, ev model.Event) error {
	n.Logger.Printf("[reminder H-1] kirim ke %s <%s>: event %q besok pukul %s di %s",
		to.FullName, to.Email, ev.Title, ev.StartAt.Format("15:04 02-01-2006"), ev.Location)
	return nil
}

// H1Window mengembalikan rentang [awal, akhir) HARI BESOK menurut zona waktu loc.
// Event yang mulai di rentang ini dapat pengingat H-1.
func H1Window(now time.Time, loc *time.Location) (time.Time, time.Time) {
	local := now.In(loc)
	tomorrow := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, loc)
	return tomorrow, tomorrow.AddDate(0, 0, 1)
}

// WIB menyediakan zona waktu Asia/Jakarta, dengan cadangan UTC+7 kalau
// database zona waktu tidak ada di mesin/container.
func WIB() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*60*60)
}

type ReminderService struct {
	events   *repository.EventRepository
	reports  *repository.ReportRepository
	notifier Notifier
	logger   *log.Logger
	loc      *time.Location

	mu   sync.Mutex
	sent map[string]bool // "eventID|tanggal" yang sudah dikirimi pengingat
}

func NewReminderService(events *repository.EventRepository, reports *repository.ReportRepository, notifier Notifier, logger *log.Logger) *ReminderService {
	return &ReminderService{
		events: events, reports: reports, notifier: notifier, logger: logger,
		loc: WIB(), sent: map[string]bool{},
	}
}

// RunH1 mencari event yang mulai besok lalu mengirim pengingat ke semua
// peserta terdaftar. Mengembalikan jumlah pengingat terkirim. Event yang
// sudah diproses di hari yang sama dilewati, jadi aman dijalankan berkala.
func (s *ReminderService) RunH1(ctx context.Context, now time.Time) (int, error) {
	from, to := H1Window(now, s.loc)
	events, err := s.events.ListOpenStartingBetween(ctx, from, to)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, ev := range events {
		key := fmt.Sprintf("%s|%s", ev.ID, from.Format("2006-01-02"))
		if !s.markSent(key) {
			continue
		}

		recipients, err := s.reports.ListRegisteredRecipients(ctx, ev.ID)
		if err != nil {
			s.unmark(key) // coba lagi pada putaran berikutnya
			return total, err
		}
		for _, rc := range recipients {
			if err := s.notifier.Notify(ctx, rc, ev); err != nil {
				s.logger.Printf("gagal kirim pengingat ke %s: %v", rc.Email, err)
				continue
			}
			total++
		}
	}
	return total, nil
}

// Start menjalankan RunH1 berkala di background sampai ctx dibatalkan.
//
//	reminders.Start(ctx, time.Hour)
func (s *ReminderService) Start(ctx context.Context, every time.Duration) {
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			if n, err := s.RunH1(ctx, time.Now()); err != nil {
				s.logger.Printf("reminder H-1 gagal: %v", err)
			} else if n > 0 {
				s.logger.Printf("reminder H-1: %d pengingat terkirim", n)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (s *ReminderService) markSent(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sent[key] {
		return false
	}
	s.sent[key] = true
	return true
}

func (s *ReminderService) unmark(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sent, key)
}