package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config menampung seluruh environment variable yang dipakai aplikasi.
type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret              string
	JWTAccessExpiryMinutes int
	JWTRefreshExpiryDays   int

	LogLevel      string
	LogPath       string
	LogMaxSizeMB  int
	LogMaxBackups int
	LogMaxAgeDays int
}

// Load membaca file .env (kalau ada) lalu merakit Config dari environment variable.
// Tidak fatal kalau .env tidak ditemukan, karena di production env var bisa diset langsung
// oleh sistem/container, bukan lewat file.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env tidak ditemukan, lanjut pakai environment variable sistem")
	}

	cfg := &Config{
		AppName: getEnv("APP_NAME", "EventKu"),
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "3000"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "eventku"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTAccessExpiryMinutes: getEnvAsInt("JWT_ACCESS_EXPIRY_MINUTES", 15),
		JWTRefreshExpiryDays:   getEnvAsInt("JWT_REFRESH_EXPIRY_DAYS", 7),

		LogLevel:      getEnv("LOG_LEVEL", "info"),
		LogPath:       getEnv("LOG_PATH", "logs/app.log"),
		LogMaxSizeMB:  getEnvAsInt("LOG_MAX_SIZE_MB", 10),
		LogMaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 5),
		LogMaxAgeDays: getEnvAsInt("LOG_MAX_AGE_DAYS", 30),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi di .env")
	}

	return cfg, nil
}

// DBConnString merakit connection string Postgres dari field-field DB di atas.
func (c *Config) DBConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("[config] %s bukan angka valid (%q), pakai default %d", key, v, fallback)
		return fallback
	}
	return i
}
