-- 001_users_roles.sql
-- Tabel identitas dan RBAC: users, roles, permissions, role_permissions.
-- Role di sini cuma level sistem (admin vs user/mahasiswa). Status "panitia"
-- TIDAK disimpan sebagai role, tapi ditentukan per-organisasi lewat tabel
-- organization_members (lihat 003_organizations.sql) dan dicek lewat
-- OrganizationRepository.IsMember.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) UNIQUE NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name     VARCHAR(150) NOT NULL,
    email         VARCHAR(150) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role_id       UUID NOT NULL REFERENCES roles(id),
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_role_id ON users(role_id);

-- === Seed (bagian RBAC di checklist) ===
-- Permission di bawah cuma starting point buat aksi yang sifatnya admin-only
-- atau lintas-organisasi. Aksi yang scope-nya "milik organisasi sendiri"
-- (create event, issue certificate, checkin, dst) dicek lewat IsMember, bukan
-- lewat permission ini. Tambah baris baru di sini kalau nanti ternyata ada
-- aksi global lain yang perlu digate oleh Can().

INSERT INTO roles (name) VALUES
    ('admin'),
    ('user');

INSERT INTO permissions (name, description) VALUES
    ('events.moderate', 'Approve/reject event yang diajukan panitia'),
    ('events.view_all', 'Lihat semua event lintas organisasi, termasuk yang belum di-approve'),
    ('organizations.create', 'Membuat organisasi baru'),
    ('organizations.manage_members', 'Menambah/mengubah anggota organisasi sendiri');

-- admin: bisa moderasi event dan lihat semua data
-- user: bisa bikin organisasi dan kelola anggota organisasi miliknya
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE (r.name = 'admin' AND p.name IN ('events.moderate', 'events.view_all'))
   OR (r.name = 'user'  AND p.name IN ('organizations.create', 'organizations.manage_members'));