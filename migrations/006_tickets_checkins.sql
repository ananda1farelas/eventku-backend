CREATE TABLE tickets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES registrations(id),
    qr_code         VARCHAR(64) NOT NULL,
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tickets_registration_key UNIQUE (registration_id),
    CONSTRAINT tickets_qr_code_key UNIQUE (qr_code)
);

CREATE TABLE checkins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id     UUID NOT NULL REFERENCES tickets(id),
    checked_in_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    checked_in_by UUID NOT NULL REFERENCES users(id),
    CONSTRAINT checkins_ticket_key UNIQUE (ticket_id)
);