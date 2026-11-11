package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadEnv membaca file .env di folder kerja.
// Jika file tidak ada, aplikasi memakai environment sistem
// (misalnya saat berjalan di Docker).
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: berkas .env tidak ditemukan, memakai environment sistem")
	}
}

// GetEnv mengambil nilai environment berdasarkan key.
// Jika key tidak ada atau kosong, fallback yang dikembalikan.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// GetEnvInt sama seperti GetEnv, tetapi hasilnya berupa angka.
// Jika nilainya bukan angka yang valid, fallback yang dikembalikan.
func GetEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("peringatan: %s bukan angka (%q), memakai bawaan %d", key, value, fallback)
		return fallback
	}
	return parsed
}
