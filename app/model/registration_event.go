package model

import "time"

type EventRegistrationInfo struct {
	Status               string
	Quota                int
	RegistrationDeadline time.Time
	RegisteredCount      int // jumlah registrasi berstatus registered saat ini
}