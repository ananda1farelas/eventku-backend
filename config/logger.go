package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger membuat logger yang menulis ke stdout SEKALIGUS ke file rotating
// (biar gampang dipantau pas dev, dan tetap tersimpan rapi pas deploy).
func NewLogger(cfg *Config) *log.Logger {
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.LogPath,
		MaxSize:    cfg.LogMaxSizeMB,  // MB sebelum rotasi
		MaxBackups: cfg.LogMaxBackups, // jumlah file backup yang disimpan
		MaxAge:     cfg.LogMaxAgeDays, // umur maksimum file backup (hari)
		Compress:   true,
	}

	multi := io.MultiWriter(os.Stdout, fileWriter)
	return log.New(multi, "["+cfg.AppName+"] ", log.LstdFlags|log.Lshortfile)
}
