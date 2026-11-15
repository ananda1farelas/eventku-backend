package helper

import (
	"errors"
	"sync"
)

// Registry ini menjembatani error "biasa" (errors.New) buatan modul lain
// supaya tetap jadi respons 4xx yang benar, tanpa mengubah file modul itu.
// Pendaftaran dilakukan sekali di awal lewat service.RegisterErrorMappings().

type errorMapping struct {
	target error
	code   int
}

var (
	errorMapMu sync.RWMutex
	errorMaps  []errorMapping
)

// RegisterError mendaftarkan: kalau sebuah error cocok (errors.Is) dengan
// target, ubah jadi AppError dengan kode HTTP tersebut.
func RegisterError(target error, code int) {
	errorMapMu.Lock()
	defer errorMapMu.Unlock()
	errorMaps = append(errorMaps, errorMapping{target: target, code: code})
}

// ResolveError mengembalikan AppError untuk err, atau nil kalau err
// tidak dikenali (artinya: perlakukan sebagai 500).
func ResolveError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}

	errorMapMu.RLock()
	defer errorMapMu.RUnlock()
	for _, m := range errorMaps {
		if errors.Is(err, m.target) {
			return &AppError{Code: m.code, Message: m.target.Error()}
		}
	}
	return nil
}