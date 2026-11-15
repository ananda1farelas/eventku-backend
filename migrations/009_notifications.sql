-- 009_notifications.sql
-- Notifikasi dalam aplikasi (mis. pengingat H-1). Satu jenis notifikasi per
-- (user, event) hanya dibuat sekali, jadi pengingat tidak ganda walaupun
-- server restart atau scheduler jalan berulang.

CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    event_id   UUID REFERENCES events(id),
    type       VARCHAR(30) NOT NULL,
    title      VARCHAR(200) NOT NULL,
    message    TEXT NOT NULL,
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX notifications_dedupe
    ON notifications (user_id, event_id, type)
    WHERE event_id IS NOT NULL;

CREATE INDEX idx_notifications_user_created ON notifications (user_id, created_at DESC);