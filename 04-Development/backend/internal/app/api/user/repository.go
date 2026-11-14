package user

import (
	"context"
	"errors"
	"fmt"
	"hii-stranger1/backend/internal/pkg/database"
	"strings"
	"time"

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

// List mengambil satu halaman user sesuai filter beserta total seluruh data
// yang cocok. Urutan nama lalu id dipilih karena tabel tidak punya kolom waktu
// dan pagination butuh urutan yang stabil.
func (r *Repository) List(ctx context.Context, f ListFilter) ([]User, int, error) {
	where, args := f.where()

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`+where, args...).Scan(&total); err != nil {
		return nil, 0, wrap("count", err)
	}

	args = append(args, f.Limit, f.Offset)
	rows, err := r.db.Query(ctx,
		fmt.Sprintf(`SELECT %s FROM users%s ORDER BY created_at DESC, id LIMIT $%d OFFSET $%d`,
			userColumns, where, len(args)-1, len(args)),
		args...)
	if err != nil {
		return nil, 0, wrap("list", err)
	}
	defer rows.Close()

	users := make([]User, 0, f.Limit)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, wrap("list scan", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, wrap("list rows", err)
	}
	return users, total, nil
}

// AdminUpdate mengubah role dan/atau status user mana pun. Mengembalikan
// ErrNotFound bila user tidak ada.
func (r *Repository) AdminUpdate(ctx context.Context, id string, c AdminChanges) (User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`UPDATE users
		 SET role_id = COALESCE((SELECT id FROM roles WHERE code = $2), role_id),
		     status = COALESCE($3, status)
		 WHERE id = $1
		 RETURNING `+userColumns,
		id, c.Role, c.Status))
	if err != nil {
		return User{}, wrap("admin update", err)
	}
	return u, nil
}

// SaveVerification menyimpan hash token verifikasi email baru untuk user dan
// menghapus token lamanya, sehingga hanya tautan terbaru yang berlaku.
func (r *Repository) SaveVerification(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	return r.replaceToken(ctx, "email_verifications", "save verification", userID, tokenHash, expiresAt)
}

// ConsumeVerification menandai email user terverifikasi bila token masih
// berlaku, lalu menghapus token tersebut (sekali pakai).
func (r *Repository) ConsumeVerification(ctx context.Context, tokenHash string) (User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return User{}, wrap("consume verification begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	err = tx.QueryRow(ctx,
		`DELETE FROM email_verifications
		  WHERE token_hash = $1 AND expires_at > now()
		  RETURNING user_id`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidVerification
	}
	if err != nil {
		return User{}, wrap("consume verification", err)
	}
	u, err := scanUser(tx.QueryRow(ctx,
		`UPDATE users SET email_verified_at = COALESCE(email_verified_at, now())
		  WHERE id = $1
		  RETURNING `+userColumns, userID))
	if err != nil {
		return User{}, wrap("consume verification update", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM email_verifications WHERE user_id = $1`, userID); err != nil {
		return User{}, wrap("consume verification cleanup", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, wrap("consume verification commit", err)
	}
	return u, nil
}

// where menyusun klausa WHERE dari filter. Hanya potongan SQL konstan yang
// disambung; seluruh nilai dari client dikirim sebagai parameter.
func (f ListFilter) where() (string, []any) {
	var conds []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}

	if f.Q != "" {
		add(`(name ILIKE $%[1]d ESCAPE '\' OR email ILIKE $%[1]d ESCAPE '\')`, "%"+database.EscapeLike(f.Q)+"%")
	}
	if f.Role != "" {
		add(`role_id = (SELECT id FROM roles WHERE code = $%d)`, f.Role)
	}
	if f.Status != "" {
		add(`status = $%d`, f.Status)
	}

	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
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

// SavePasswordReset menyimpan hash token reset password baru dan menghapus
// token lama user, sehingga hanya tautan terbaru yang berlaku.
func (r *Repository) SavePasswordReset(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	return r.replaceToken(ctx, "password_resets", "save password reset", userID, tokenHash, expiresAt)
}

// replaceToken menghapus token lama user di table lalu menyimpan hash token
// baru dalam satu transaksi. table selalu konstanta dari pemanggil.
func (r *Repository) replaceToken(ctx context.Context, table, op, userID, tokenHash string, expiresAt time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return wrap(op+" begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM `+table+` WHERE user_id = $1`, userID); err != nil {
		return wrap(op+" clear", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO `+table+` (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		tokenHash, userID, expiresAt); err != nil {
		return wrap(op+" insert", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return wrap(op+" commit", err)
	}
	return nil
}

// ConsumePasswordReset mengganti password bila token masih berlaku lalu
// menghapus semua token reset user (sekali pakai). Karena tautan dikirim ke
// email, keberhasilan reset sekaligus membuktikan kepemilikan email.
func (r *Repository) ConsumePasswordReset(ctx context.Context, tokenHash, passwordHash string) (User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return User{}, wrap("consume password reset begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	err = tx.QueryRow(ctx,
		`DELETE FROM password_resets
		  WHERE token_hash = $1 AND expires_at > now()
		  RETURNING user_id`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidPasswordReset
	}
	if err != nil {
		return User{}, wrap("consume password reset", err)
	}
	u, err := scanUser(tx.QueryRow(ctx,
		`UPDATE users SET password_hash = $2, email_verified_at = COALESCE(email_verified_at, now())
		  WHERE id = $1
		  RETURNING `+userColumns, userID, passwordHash))
	if err != nil {
		return User{}, wrap("consume password reset update", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM password_resets WHERE user_id = $1`, userID); err != nil {
		return User{}, wrap("consume password reset cleanup", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, wrap("consume password reset commit", err)
	}
	return u, nil
}
