-- 007_payments.sql
-- Pembayaran MOCK untuk event berbayar (tanpa payment gateway sungguhan).
-- Satu registrasi boleh punya banyak percobaan bayar, tapi hanya SATU yang
-- aktif (pending atau paid) pada satu waktu. Yang failed boleh diulang.

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES registrations(id),
    amount          NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    method          VARCHAR(20) NOT NULL
                    CHECK (method IN ('transfer', 'ewallet', 'qris')),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'paid', 'failed')),
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payments_paid_at_check CHECK ((status = 'paid') = (paid_at IS NOT NULL))
);

CREATE UNIQUE INDEX payments_one_active_per_registration
    ON payments (registration_id)
    WHERE status IN ('pending', 'paid');

CREATE INDEX idx_payments_registration_id ON payments (registration_id);