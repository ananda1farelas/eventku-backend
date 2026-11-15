-- 004_events.sql
-- Event milik sebuah organisasi. Alur status:
-- draft -> pending_review -> open | rejected, lalu open -> closed -> done.
-- Aturan perpindahan status ada di service/event_rules.go (bukan di DB).

CREATE TABLE events (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- event_no: nomor urut ramah-manusia, dipakai untuk nomor sertifikat
    -- (contoh EVT-12-2026-0007). ID asli tetap UUID.
    event_no              INTEGER GENERATED ALWAYS AS IDENTITY UNIQUE,
    organization_id       UUID NOT NULL REFERENCES organizations(id),
    created_by            UUID NOT NULL REFERENCES users(id),
    title                 VARCHAR(200) NOT NULL,
    description           TEXT NOT NULL DEFAULT '',
    category              VARCHAR(50)  NOT NULL DEFAULT 'umum',
    location              VARCHAR(200) NOT NULL,
    start_at              TIMESTAMPTZ NOT NULL,
    end_at                TIMESTAMPTZ NOT NULL,
    quota                 INTEGER NOT NULL CHECK (quota > 0),
    price                 NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    registration_deadline TIMESTAMPTZ NOT NULL,
    status                VARCHAR(20) NOT NULL DEFAULT 'draft'
                          CHECK (status IN ('draft', 'pending_review', 'open', 'rejected', 'closed', 'done')),
    rejection_reason      TEXT,
    deleted_at            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT events_schedule_check CHECK (end_at > start_at),
    CONSTRAINT events_deadline_check CHECK (registration_deadline <= start_at)
);

CREATE INDEX idx_events_org_id ON events (organization_id);
CREATE INDEX idx_events_status_start ON events (status, start_at) WHERE deleted_at IS NULL;