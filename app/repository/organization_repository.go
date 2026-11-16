package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventku/app/model"
)

// ErrAlreadyMember dibalikin kalau AddMember dipanggil buat user yang
// sudah jadi anggota organisasi itu (kena UNIQUE (organization_id, user_id)).
var ErrAlreadyMember = errors.New("user sudah menjadi anggota organisasi ini")

type OrganizationRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationRepository(db *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create menyimpan organisasi baru DAN langsung mendaftarkan pembuatnya
// sebagai anggota dengan posisi "owner", dalam satu transaksi — supaya
// gak pernah ada organisasi tanpa owner (misal request kedua gagal karena
// koneksi putus di tengah).
func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx create organization: %w", err)
	}
	defer tx.Rollback(ctx) // no-op kalau sudah Commit

	err = tx.QueryRow(ctx, `
		INSERT INTO organizations (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`,
		org.Name, org.Description, org.OwnerID,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert organization: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, position)
		VALUES ($1, $2, $3)`,
		org.ID, org.OwnerID, model.OrgPositionOwner,
	)
	if err != nil {
		return fmt.Errorf("insert owner sebagai member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx create organization: %w", err)
	}
	return nil
}

// FindByID dipakai sebelum Update/AddMember, buat mastiin organisasi
// yang dimaksud memang ada sebelum ngecek hak akses lebih jauh.
func (r *OrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	org := &model.Organization{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM organizations
		WHERE id = $1`,
		id,
	).Scan(&org.ID, &org.Name, &org.Description, &org.OwnerID, &org.CreatedAt, &org.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find organization by id: %w", err)
	}
	return org, nil
}

// List dipakai GET /organizations, urut dari yang terbaru.
func (r *OrganizationRepository) List(ctx context.Context, limit, offset int) ([]model.Organization, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM organizations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []model.Organization
	for rows.Next() {
		var o model.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Description, &o.OwnerID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		orgs = append(orgs, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizations: %w", err)
	}
	return orgs, nil
}

// Update cuma mengubah name & description. owner_id sengaja gak bisa
// diubah lewat method ini — pindah kepemilikan (kalau nanti dibutuhkan)
// harus jadi method terpisah yang lebih eksplisit, bukan nyelip di Update biasa.
func (r *OrganizationRepository) Update(ctx context.Context, id uuid.UUID, name, description string) error {
	cmd, err := r.db.Exec(ctx, `
		UPDATE organizations
		SET name = $1, description = $2, updated_at = now()
		WHERE id = $3`,
		name, description, id,
	)
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// AddMember dipanggil pas owner/staff nambah anggota baru.
func (r *OrganizationRepository) AddMember(ctx context.Context, orgID, userID uuid.UUID, position string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, position)
		VALUES ($1, $2, $3)`,
		orgID, userID, position,
	)
	if err != nil {
		if isUniqueViolation(err, "organization_members_organization_id_user_id_key") {
			return ErrAlreadyMember
		}
		return fmt.Errorf("add organization member: %w", err)
	}
	return nil
}

// ListMembers dipakai buat nampilin daftar anggota satu organisasi,
// JOIN ke users biar nama & email anggota ikut ke-load sekalian.
func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]model.OrganizationMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT om.id, om.organization_id, om.user_id, om.position, om.created_at,
		       u.full_name, u.email
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1
		ORDER BY om.created_at ASC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("list organization members: %w", err)
	}
	defer rows.Close()

	var members []model.OrganizationMember
	for rows.Next() {
		var m model.OrganizationMember
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.UserID, &m.Position, &m.CreatedAt,
			&m.UserFullName, &m.UserEmail,
		); err != nil {
			return nil, fmt.Errorf("scan organization member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organization members: %w", err)
	}
	return members, nil
}

// IsMember adalah KONTRAK PALING PENTING buat Orang 2 & Orang 3: dipanggil
// di service event/registrasi/checkin mereka buat ngecek "user ini panitia
// di organisasi X atau bukan" sebelum ngizinin create event / approve
// registrasi / checkin. Sengaja cuma balikin (bool, error) — gak peduli
// posisinya owner atau staff, soal itu beda cerita (lihat IsOwner kalau
// nanti butuh pembedaan lebih detail).
func (r *OrganizationRepository) IsMember(ctx context.Context, organizationID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM organization_members
			WHERE organization_id = $1 AND user_id = $2
		)`,
		organizationID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("cek organization member: %w", err)
	}
	return exists, nil
}

// IsOwner dipakai kalau suatu aksi cuma boleh dilakukan owner, bukan staff
// biasa (misal: hapus organisasi, ganti kepemilikan — belum ada endpoint-nya
// sekarang, disiapin aja).
func (r *OrganizationRepository) IsOwner(ctx context.Context, organizationID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM organization_members
			WHERE organization_id = $1 AND user_id = $2 AND position = $3
		)`,
		organizationID, userID, model.OrgPositionOwner,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("cek organization owner: %w", err)
	}
	return exists, nil
}
