-- 008_certificates.sql
-- Satu sertifikat per registrasi (peserta yang hadir). Nomor sertifikat unik.
-- file_url boleh kosong: sertifikat saat ini berupa data, file PDF menyusul.

CREATE TABLE certificates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES registrations(id),
    certificate_no  VARCHAR(50) NOT NULL,
    file_url        TEXT,
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT certificates_registration_key UNIQUE (registration_id),
    CONSTRAINT certificates_no_key UNIQUE (certificate_no)
);