package repository

import (
	"context"
	"database/sql"
	"errors"

	"hii-stranger1/backend/models"
)

// ErrKatalogNotFound adalah error khusus yang menandakan
// bahwa data katalog yang dicari tidak ditemukan di database.
//
// Error ini bisa digunakan oleh GetByID, Update, dan Delete.
// Dengan error khusus, layer lain bisa membedakan kasus data
// tidak ditemukan dari error database biasa.
var ErrKatalogNotFound = errors.New("katalog tidak ditemukan")

// KatalogRepository bertanggung jawab untuk berkomunikasi
// dengan database terkait data katalog.
//
// Field db menyimpan koneksi database yang akan digunakan
// oleh semua method repository.
type KatalogRepository struct {
	db *sql.DB
}

// NewKatalogRepository adalah constructor untuk membuat
// instance KatalogRepository.
//
// Parameter:
//   - db: koneksi database yang sudah dibuat sebelumnya.
//
// Return:
//   - Pointer ke KatalogRepository yang menyimpan koneksi db.
//
// Constructor ini tidak membuka koneksi baru. Ia hanya menerima
// koneksi database yang diberikan dan menyimpannya.
func NewKatalogRepository(db *sql.DB) *KatalogRepository {
	return &KatalogRepository{
		db: db,
	}
}

// Create menyimpan data katalog baru ke database.
//
// Alur:
//  1. Menyiapkan query INSERT.
//  2. Mengirim nilai dari struct models.Katalog ke query.
//  3. Menjalankan query menggunakan ExecContext.
//  4. Mengembalikan error jika proses insert gagal.
//
// ctx digunakan agar operasi database bisa mengikuti
// pembatalan atau deadline dari pemanggil.
func (r *KatalogRepository) Create(
	ctx context.Context,
	katalog *models.Katalog,
) error {
	// $1 sampai $6 adalah placeholder PostgreSQL.
	// Nilainya dikirim secara terpisah melalui ExecContext.
	//
	// Urutan nilai harus sesuai dengan urutan placeholder:
	// $1 = Name
	// $2 = Slug
	// $3 = Description
	// $4 = IsActive
	// $5 = CreatedAt
	// $6 = UpdatedAt
	query := `
		INSERT INTO katalog (
			name_katalog,
			slug,
			description,
			is_active,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// ExecContext digunakan untuk query yang tidak membutuhkan
	// hasil berupa kumpulan baris, seperti INSERT.
	//
	// Hasil pertama adalah sql.Result, tetapi tidak dibutuhkan
	// untuk operasi ini sehingga disimpan ke variabel _.
	//
	// Hasil kedua adalah error yang harus diperiksa.
	_, err := r.db.ExecContext(
		ctx,
		query,
		katalog.Name,
		katalog.Slug,
		katalog.Description,
		katalog.IsActive,
		katalog.CreatedAt,
		katalog.UpdatedAt,
	)
	if err != nil {
		// Contoh: koneksi database gagal, constraint dilanggar,
		// atau query tidak sesuai dengan skema database.
		return err
	}

	// nil berarti operasi INSERT berhasil tanpa error database.
	return nil
}

// GetAll mengambil seluruh data katalog dari database.
//
// Alur:
//  1. Menjalankan SELECT menggunakan QueryContext.
//  2. Menutup rows setelah method selesai.
//  3. Membaca setiap baris dengan rows.Next().
//  4. Memasukkan nilai kolom ke struct models.Katalog.
//  5. Memeriksa error yang mungkin muncul selama iterasi.
//  6. Mengembalikan slice katalog.
func (r *KatalogRepository) GetAll(
	ctx context.Context,
) ([]models.Katalog, error) {
	// Nama kolom ditulis secara eksplisit agar urutan hasil query
	// jelas dan sesuai dengan urutan rows.Scan().
	query := `
		SELECT
			id,
			name_katalog,
			slug,
			description,
			is_active,
			created_at,
			updated_at
		FROM katalog
	`

	// QueryContext digunakan karena SELECT ini dapat menghasilkan
	// banyak baris yang perlu dibaca satu per satu.
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// rows memegang hasil query database.
	// Close wajib dipanggil agar resource hasil query dilepaskan.
	// defer membuat Close dipanggil ketika method selesai.
	defer rows.Close()

	// Slice untuk menampung semua katalog yang berhasil dibaca.
	// Slice nil akan menjadi slice kosong setelah append pertama.
	var katalogs []models.Katalog

	// Next memindahkan posisi pembacaan ke baris berikutnya.
	// Loop berhenti ketika sudah tidak ada baris lagi.
	for rows.Next() {
		var katalog models.Katalog

		// Scan membaca kolom dari baris saat ini ke field struct.
		//
		// Urutan argumen harus sama dengan urutan kolom SELECT.
		if err := rows.Scan(
			&katalog.ID,
			&katalog.Name,
			&katalog.Slug,
			&katalog.Description,
			&katalog.IsActive,
			&katalog.CreatedAt,
			&katalog.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Tambahkan katalog yang sudah berhasil di-scan
		// ke dalam slice hasil.
		katalogs = append(katalogs, katalog)
	}

	// rows.Next() yang berhenti tidak selalu berarti semua proses
	// berhasil. Periksa error yang terjadi selama iterasi.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return katalogs, nil
}

// GetByID mengambil satu katalog berdasarkan ID.
//
// Jika ID tidak ditemukan, method mengembalikan
// ErrKatalogNotFound, bukan sql.ErrNoRows secara langsung.
func (r *KatalogRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.Katalog, error) {
	query := `
		SELECT
			id,
			name_katalog,
			slug,
			description,
			is_active,
			created_at,
			updated_at
		FROM katalog
		WHERE id = $1
	`

	// QueryRowContext digunakan ketika hasil yang diharapkan
	// adalah satu baris.
	//
	// QueryRowContext tidak langsung mengembalikan error.
	// Error pembacaan diperiksa ketika Scan dipanggil.
	row := r.db.QueryRowContext(ctx, query, id)

	var katalog models.Katalog

	err := row.Scan(
		&katalog.ID,
		&katalog.Name,
		&katalog.Slug,
		&katalog.Description,
		&katalog.IsActive,
		&katalog.CreatedAt,
		&katalog.UpdatedAt,
	)
	if err != nil {
		// sql.ErrNoRows muncul ketika query tidak menemukan baris.
		// errors.Is lebih fleksibel daripada membandingkan error
		// secara langsung dengan operator ==.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKatalogNotFound
		}

		// Error lainnya diteruskan ke pemanggil.
		return nil, err
	}

	// Mengembalikan pointer ke katalog yang sudah terisi.
	return &katalog, nil
}

// Update memperbarui data katalog berdasarkan ID.
//
// Alur:
//  1. Menjalankan UPDATE.
//  2. Memeriksa error query.
//  3. Memeriksa jumlah baris yang terdampak.
//  4. Mengembalikan ErrKatalogNotFound jika tidak ada baris
//     yang terdampak.
func (r *KatalogRepository) Update(
	ctx context.Context,
	id int64,
	katalog *models.Katalog,
) error {
	query := `
		UPDATE katalog
		SET
			name_katalog = $1,
			slug = $2,
			description = $3,
			is_active = $4,
			updated_at = $5
		WHERE id = $6
	`

	// UPDATE tidak mengembalikan kumpulan baris.
	// ExecContext mengembalikan sql.Result dan error.
	//
	// Perhatikan bahwa ID berada pada placeholder $6 karena
	// lima placeholder sebelumnya digunakan untuk nilai update.
	result, err := r.db.ExecContext(
		ctx,
		query,
		katalog.Name,
		katalog.Slug,
		katalog.Description,
		katalog.IsActive,
		katalog.UpdatedAt,
		id,
	)
	if err != nil {
		return err
	}

	// RowsAffected memberikan jumlah baris yang terdampak query.
	// Jangan menjalankan ExecContext kedua hanya untuk menghitung
	// baris karena itu akan menjalankan UPDATE sekali lagi.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Jika tidak ada baris yang terdampak, ID mungkin tidak ada.
	// Pada PostgreSQL, UPDATE yang menemukan baris biasanya
	// tetap menghitung baris tersebut meskipun nilainya sama.
	if rowsAffected == 0 {
		return ErrKatalogNotFound
	}

	return nil
}

// Delete menghapus satu katalog berdasarkan ID.
//
// Alur:
//  1. Menjalankan DELETE.
//  2. Memeriksa error query.
//  3. Memeriksa jumlah baris yang dihapus.
//  4. Mengembalikan ErrKatalogNotFound jika tidak ada
//     baris yang dihapus.
func (r *KatalogRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	query := `DELETE FROM katalog WHERE id = $1`

	// ExecContext menjalankan query DELETE dengan ID sebagai
	// parameter $1.
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Periksa berapa baris yang berhasil dihapus.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Nol berarti tidak ada baris yang cocok dengan ID tersebut.
	if rowsAffected == 0 {
		return ErrKatalogNotFound
	}

	return nil
}
