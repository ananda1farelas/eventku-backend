package service

import (
	"fmt"

	"eventku/app/model"
)

// BuildReminderMessage menyusun judul dan isi pengingat H-1 untuk satu penerima.
func BuildReminderMessage(to model.Recipient, ev model.Event) (title, message string) {
	title = fmt.Sprintf("Pengingat: %s dimulai besok", ev.Title)
	message = fmt.Sprintf(
		"Halo %s, event \"%s\" dimulai besok pukul %s WIB di %s. Jangan lupa tunjukkan tiket QR-mu saat check-in.",
		to.FullName, ev.Title, ev.StartAt.In(WIB()).Format("15:04"), ev.Location)
	return title, message
}