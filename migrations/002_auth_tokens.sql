-- 002_auth_tokens.sql
-- Refresh token disimpan dalam bentuk HASH (bukan plaintext), supaya kalau
-- tabel ini bocor, token yang ada gak langsung bisa dipakai login.
-- Alur: service hash token pakai SHA-256 sebelum INSERT/SELECT ke kolom
-- token_hash; yang dikirim ke client cuma token mentahnya.

CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);