package service

import (
	"testing"
	"time"
)

func TestH1Window(t *testing.T) {
	wib := time.FixedZone("WIB", 7*60*60)

	tests := []struct {
		name     string
		now      time.Time
		wantFrom time.Time
	}{
		{"siang hari", time.Date(2026, 11, 15, 12, 0, 0, 0, wib), time.Date(2026, 11, 16, 0, 0, 0, 0, wib)},
		{"tepat tengah malam", time.Date(2026, 11, 15, 0, 0, 0, 0, wib), time.Date(2026, 11, 16, 0, 0, 0, 0, wib)},
		{"akhir bulan", time.Date(2026, 11, 30, 23, 59, 0, 0, wib), time.Date(2026, 12, 1, 0, 0, 0, 0, wib)},
		{"akhir tahun", time.Date(2026, 12, 31, 8, 0, 0, 0, wib), time.Date(2027, 1, 1, 0, 0, 0, 0, wib)},
		// 20:00 UTC tanggal 15 sudah 03:00 WIB tanggal 16, jadi "besok" = tanggal 17.
		{"UTC dikonversi ke WIB", time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC), time.Date(2026, 11, 17, 0, 0, 0, 0, wib)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to := H1Window(tt.now, wib)
			if !from.Equal(tt.wantFrom) {
				t.Errorf("from = %v, want %v", from, tt.wantFrom)
			}
			if !to.Equal(tt.wantFrom.AddDate(0, 0, 1)) {
				t.Errorf("to = %v, want satu hari setelah from", to)
			}
		})
	}
}