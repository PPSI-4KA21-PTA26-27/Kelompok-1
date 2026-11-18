package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// role dibaca sebagai kode dari tabel roles (users.role_id), sehingga
	// struct User dan token tetap memakai kode role (customer, staff, owner).
	userColumns = `id, name, email, password_hash, phone,
		(SELECT r.code FROM roles r WHERE r.id = users.role_id) AS role,
		status, email_verified_at, created_at`
	uniqueViolation = "23505"
	emailConstraint = "uq_users_email"
)

// Repository mengakses tabel users.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository di atas pool koneksi.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create menyimpan user baru. ID dibuat database. Mengembalikan ErrEmailTaken
// bila email sudah ada.
func (r *Repository) Create(ctx context.Context, u User) (User, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, phone, role_id, status)
		 VALUES ($1, $2, $3, $4, (SELECT id FROM roles WHERE code = $5), $6)
		 RETURNING `+userColumns,
		u.Name, u.Email, u.PasswordHash, u.Phone, u.Role, u.Status)

	created, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == emailConstraint {
			return User{}, ErrEmailTaken
		}
		return User{}, wrap("create", err)
	}
	return created, nil
}

// GetByEmail mencari user berdasarkan email (sudah dinormalisasi huruf kecil).
func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email))
	if err != nil {
		return User{}, wrap("get by email", err)
	}
	return u, nil
}

// GetByID mencari user berdasarkan ID.
func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	if err != nil {
		return User{}, wrap("get by id", err)
	}
	return u, nil
}

// UpdateProfile mengubah nama dan/atau nomor telepon user yang berstatus
// StatusActive dalam satu query. Field yang tidak diubah dibiarkan apa adanya.
// Mengembalikan ErrNotFound bila user tidak ada atau tidak aktif.
func (r *Repository) UpdateProfile(ctx context.Context, id string, c ProfileChanges) (User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`UPDATE users
		 SET name = COALESCE($2, name),
		     phone = CASE WHEN $3 THEN $4 ELSE phone END
		 WHERE id = $1 AND status = $5
		 RETURNING `+userColumns,
		id, c.Name, c.Phone.Set, c.Phone.Value, StatusActive))
	if err != nil {
		return User{}, wrap("update profile", err)
	}
	return u, nil
}

// UpdatePasswordHash mengganti hash password user. Mengembalikan ErrNotFound
// bila user tidak ada.
func (r *Repository) UpdatePasswordHash(ctx context.Context, id, hash string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $2 WHERE id = $1`, id, hash)
	if err != nil {
		return wrap("update password", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.Role, &u.Status, &u.EmailVerifiedAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

// wrap menambahkan konteks operasi, tetapi membiarkan error domain tetap utuh
// agar pemanggil bisa memeriksanya dengan errors.Is.
func wrap(op string, err error) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrEmailTaken) || errors.Is(err, ErrInvalidVerification) {
		return err
	}
	return fmt.Errorf("user: %s: %w", op, err)
}
