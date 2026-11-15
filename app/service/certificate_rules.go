package service

import (
	"fmt"
	"time"

	"eventku/app/model"
	"eventku/helper"
)

var (
	ErrCertificateForbidden       = helper.NewForbidden("hanya panitia event ini yang boleh menerbitkan sertifikat")
	ErrCertificateEventNotStarted = helper.NewConflict("sertifikat baru bisa diterbitkan setelah event dimulai")
	ErrCertificateEventStatus     = helper.NewConflict("sertifikat hanya bisa diterbitkan untuk event yang berstatus open, closed, atau done")
	ErrCertificateNotAttended     = helper.NewConflict("sertifikat hanya untuk peserta yang sudah check-in")
	ErrCertificateAlreadyIssued   = helper.NewConflict("sertifikat untuk peserta ini sudah diterbitkan")
)

// CanIssueCertificates: hanya panitia event, event harus sudah berjalan.
func CanIssueCertificates(ev model.Event, isMember bool, now time.Time) error {
	if !isMember {
		return ErrCertificateForbidden
	}
	switch ev.Status {
	case model.EventStatusOpen, model.EventStatusClosed, model.EventStatusDone:
	default:
		return ErrCertificateEventStatus
	}
	if now.Before(ev.StartAt) {
		return ErrCertificateEventNotStarted
	}
	return nil
}

// EligibleForCertificate: hanya yang punya check-in, dan belum punya sertifikat.
func EligibleForCertificate(c model.CertificateCandidate) error {
	if !c.Attended {
		return ErrCertificateNotAttended
	}
	if c.HasCertificate {
		return ErrCertificateAlreadyIssued
	}
	return nil
}

// FormatCertificateNo membentuk nomor sertifikat, contoh EVT-12-2026-0007.
// TODO: setelah helper.FormatCertificateNo milik Zidan di-merge, pakai itu
// dan hapus fungsi ini (format dan isinya harus sama).
func FormatCertificateNo(eventNo, year, seq int) string {
	return fmt.Sprintf("EVT-%d-%d-%04d", eventNo, year, seq)
}