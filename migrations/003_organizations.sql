-- 003_organizations.sql
-- organizations = organisasi mahasiswa pemilik event.
-- organization_members = siapa aja yang jadi "panitia" di organisasi itu;
-- posisi owner/staff dicek di service (service/organization_rules.go),
-- bukan di DB constraint lain selain CHECK dasar.

CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(150) NOT NULL,
    description TEXT,
    owner_id    UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organization_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position        VARCHAR(20) NOT NULL CHECK (position IN ('owner', 'staff')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, user_id)
);

-- Dipakai berat oleh OrganizationRepository.IsMember(ctx, orgID, userID),
-- yang dicek Orang 2 & Orang 3 di hampir setiap request panitia.
CREATE INDEX idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX idx_org_members_org_id ON organization_members(organization_id);