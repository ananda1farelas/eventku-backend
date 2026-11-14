CREATE TABLE registrations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    UUID NOT NULL REFERENCES events(id),
    user_id     UUID NOT NULL REFERENCES users(id),
    status      VARCHAR(20) NOT NULL
                CHECK (status IN ('registered', 'waitlist', 'cancelled')),
    queued_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT registrations_event_user_key UNIQUE (event_id, user_id)
);

-- Dipakai buat hitung peserta terisi dan ambil antrean waitlist teratas.
CREATE INDEX idx_registrations_event_status_queue
    ON registrations (event_id, status, queued_at);

CREATE INDEX idx_registrations_user_id ON registrations (user_id);