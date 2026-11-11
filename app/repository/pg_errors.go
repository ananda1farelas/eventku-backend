package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation ngecek apakah error dari pgx itu pelanggaran unique
// constraint (kode "23505"), dan opsional cocokin nama constraint-nya.
// Dipakai supaya repo bisa nerjemahin error driver mentah jadi error
// domain yang jelas (ErrDuplicateEmail, dst) tanpa string-matching pesan.
func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	if constraintName == "" {
		return true
	}
	return pgErr.ConstraintName == constraintName
}
