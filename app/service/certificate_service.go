package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"eventku/app/model"
	"eventku/app/repository"
	"eventku/helper"
)

type CertificateService struct {
	events  *repository.EventRepository
	certs   *repository.CertificateRepository
	members OrganizationMemberChecker
	tx      *repository.TxManager
	now     func() time.Time
}

func NewCertificateService(
	events *repository.EventRepository,
	certs *repository.CertificateRepository,
	members OrganizationMemberChecker,
	tx *repository.TxManager,
) *CertificateService {
	return &CertificateService{events: events, certs: certs, members: members, tx: tx, now: time.Now}
}

// IssueForEvent menerbitkan sertifikat untuk SEMUA peserta yang sudah
// check-in dan belum punya sertifikat. Aman dipanggil berulang: yang
// sudah punya dilewati, nomor tidak pernah ganda.
func (s *CertificateService) IssueForEvent(ctx context.Context, actor Actor, eventID uuid.UUID) (*model.IssueCertificatesResult, error) {
	peek, err := s.events.FindByID(ctx, eventID)
	if err != nil {
		return nil, mapEventErr(err)
	}
	isMember, err := s.members.IsMember(ctx, peek.OrganizationID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("cek keanggotaan organisasi: %w", err)
	}
	if err := CanIssueCertificates(*peek, isMember, s.now()); err != nil {
		return nil, err
	}

	result := &model.IssueCertificatesResult{Certificates: []model.Certificate{}}

	err = s.tx.WithTx(ctx, func(db repository.DBTX) error {
		// Kunci baris event: dua panitia yang menekan "terbitkan" bersamaan
		// akan antre, sehingga nomor urut tidak bentrok.
		ev, err := s.events.WithTx(db).LockByID(ctx, eventID)
		if err != nil {
			return mapEventErr(err)
		}

		certs := s.certs.WithTx(db)
		candidates, err := certs.ListCandidates(ctx, eventID)
		if err != nil {
			return err
		}
		seq, err := certs.CountByEvent(ctx, eventID)
		if err != nil {
			return err
		}

		// Reset hasil kalau transaction diulang (retry TxManager).
		result.Issued, result.Skipped = 0, 0
		result.Certificates = result.Certificates[:0]

		year := s.now().Year()
		for _, cand := range candidates {
			if err := EligibleForCertificate(cand); err != nil {
				result.Skipped++
				continue
			}
			seq++
			cert := &model.Certificate{
				RegistrationID: cand.RegistrationID,
				CertificateNo:  FormatCertificateNo(ev.EventNo, year, seq),
			}
			if err := certs.Create(ctx, cert); err != nil {
				return err
			}
			result.Issued++
			result.Certificates = append(result.Certificates, *cert)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, repository.ErrCertificateExists) {
			return nil, helper.NewConflict("sertifikat sedang diterbitkan oleh proses lain, coba lagi")
		}
		return nil, err
	}
	return result, nil
}

// ListMine untuk GET /me/certificates.
func (s *CertificateService) ListMine(ctx context.Context, actor Actor) ([]model.CertificateDetail, error) {
	return s.certs.ListByUser(ctx, actor.UserID)
}