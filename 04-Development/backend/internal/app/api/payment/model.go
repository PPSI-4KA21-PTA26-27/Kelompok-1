// Package payment mengelola pembayaran order: pembuatan tagihan dengan kode
// unik, unggah bukti transfer, dan verifikasi oleh admin.
package payment

import (
	"errors"
	"time"
)

var (
	// ErrNotFound berarti pembayaran (atau bukti pembayaran) tidak ada.
	ErrNotFound = errors.New("payment: not found")
	// ErrOrderNotFound berarti order tidak ada atau bukan milik pemanggil.
	ErrOrderNotFound = errors.New("payment: order not found")
	// ErrAlreadyExists berarti order sudah punya pembayaran.
	ErrAlreadyExists = errors.New("payment: already exists")
	// ErrOrderNotPayable berarti status order tidak lagi menunggu pembayaran.
	ErrOrderNotPayable = errors.New("payment: order is not awaiting payment")
	// ErrAlreadyApproved berarti pembayaran sudah disetujui sehingga bukti
	// tidak boleh diganti.
	ErrAlreadyApproved = errors.New("payment: already approved")
	// ErrNoProof berarti pembayaran belum punya bukti sehingga belum bisa diverifikasi.
	ErrNoProof = errors.New("payment: proof not uploaded")
	// ErrAlreadyReviewed berarti pembayaran sudah diverifikasi sebelumnya.
	ErrAlreadyReviewed = errors.New("payment: already reviewed")
	// ErrStorageDisabled berarti penyimpanan berkas (Cloudinary) belum
	// dikonfigurasi sehingga bukti tidak bisa diunggah atau dibuka.
	ErrStorageDisabled = errors.New("payment: proof storage not configured")
)

// Status pembayaran. Alurnya: pending (menunggu bukti / verifikasi) lalu
// approved atau rejected. Bukti yang ditolak boleh diunggah ulang, dan
// statusnya kembali pending.
const (
	StatusPending  = "menunggu"
	StatusApproved = "valid"
	StatusRejected = "ditolak"
)

// Metode pembayaran yang diterima aplikasi.
const (
	// MethodQRIS adalah satu-satunya metode untuk order baru: QRIS statis toko,
	// nominal unik (total + kode unik) diverifikasi admin dari bukti bayar.
	MethodQRIS = "qris"
)

// Status order yang dibaca/diubah oleh package ini. Nilainya HARUS sama
// dengan konstanta di package order (package order mengimpor payment, jadi
// payment tidak bisa mengimpor order tanpa membuat import cycle).
const (
	orderStatusPendingPayment = "pending_payment"
	orderStatusPaid           = "paid"
)

// Payment adalah data pembayaran (schema Payment). Alamat bukti di Cloudinary
// (proof_url) tidak pernah ikut di JSON supaya bukti hanya bisa dibuka lewat
// GET /orders/{orderId}/payment/proof yang memeriksa hak akses; HasProof
// menandakan bukti sudah diunggah dan ProofType diturunkan dari alamatnya.
type Payment struct {
	ID         string     `json:"id"`
	OrderID    string     `json:"order_id"`
	Method     string     `json:"method"`
	Amount     string     `json:"amount"`
	UniqueCode int        `json:"unique_code"`
	ProofType  *string    `json:"proof_type"`
	HasProof   bool       `json:"has_proof"`
	Status     string     `json:"status"`
	PaidAt     *time.Time `json:"paid_at"`
	ReviewedBy *string    `json:"reviewed_by"`
	ReviewedAt *time.Time `json:"reviewed_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Proof adalah isi berkas bukti pembayaran beserta tipe MIME-nya.
// AdminItem adalah pembayaran di antrean admin beserta ringkasan ordernya.
type AdminItem struct {
	Payment
	OrderNumber   string     `json:"order_number"`
	OrderStatus   string     `json:"order_status"`
	OrderTotal    string     `json:"order_total"`
	PaymentDueAt  *time.Time `json:"payment_due_at"`
	CustomerName  string     `json:"customer_name"`
	CustomerEmail string     `json:"customer_email"`
}

type Proof struct {
	Data []byte
	Type string
}

// proofRef adalah alamat bukti di Cloudinary beserta tipe MIME-nya.
type proofRef struct {
	URL  string
	Type string
}

// Filter adalah syarat pencarian daftar pembayaran (admin).
type Filter struct {
	Status  string
	OrderID string
	Q       string // cari sebagian: nomor order, ID pembayaran, nama/email pelanggan
}

// contact berisi data untuk email notifikasi.
type contact struct {
	Email       string
	Name        string
	OrderNumber string
}

type createRequest struct {
	Method string `json:"method"`
}

type reviewRequest struct {
	Status string `json:"status"`
}
