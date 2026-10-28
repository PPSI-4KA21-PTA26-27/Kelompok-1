package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"hii-stranger1/backend/models"
)

// Alur: services.AuthService -> UserRepository -> db.DB (PostgreSQL).
// Paket ini hanya berisi query SQL. Aturan bisnis (validasi, hash password) ada di services.

var (
	// ErrNotFound: data yang dicari tidak ada.
	ErrNotFound = errors.New("data tidak ditemukan")
	// ErrEmailTaken: email sudah dipakai akun lain.
	ErrEmailTaken = errors.New("email sudah terdaftar")
)

// pgUniqueViolation adalah kode error PostgreSQL untuk pelanggaran UNIQUE.
const pgUniqueViolation = "23505"

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Kolom yang dibaca untuk membentuk models.User. phone yang NULL dibaca sebagai "".
const userColumns = `id, full_name, email, COALESCE(phone, ''), role, created_at, password_hash`

func scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.FullName, &u.Email, &u.Phone, &u.Role, &u.CreatedAt, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Create menyimpan akun baru. ID dan CreatedAt diisi dari database.
func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (full_name, email, phone, password_hash, role)
		 VALUES ($1, $2, NULLIF($3, ''), $4, $5)
		 RETURNING id, created_at`,
		u.FullName, u.Email, u.Phone, u.PasswordHash, u.Role,
	).Scan(&u.ID, &u.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return ErrEmailTaken
	}
	if err != nil {
		return fmt.Errorf("simpan user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("cari user by email: %w", err)
	}
	return u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("cari user by id: %w", err)
	}
	return u, err
}

// UpdateProfile mengubah nama dan nomor HP.
func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, fullName, phone string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET full_name = $2, phone = NULLIF($3, '') WHERE id = $1`,
		id, fullName, phone)
	if err != nil {
		return fmt.Errorf("update profil: %w", err)
	}
	return expectOneRow(res)
}

// UpdatePassword mengganti hash kata sandi.
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $2 WHERE id = $1`, id, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return expectOneRow(res)
}

func expectOneRow(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("cek hasil update: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
