package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"hii-stranger1/backend/internal/pkg/cloudinary"
	"hii-stranger1/backend/internal/pkg/mailer"
)


// ProofStorage menyimpan dan mengambil berkas bukti pembayaran (Cloudinary).
type ProofStorage interface {
	Enabled() bool
	UploadProof(ctx context.Context, contentType string, data []byte) (string, error)
	Fetch(ctx context.Context, url string, maxBytes int64) ([]byte, error)
}

// Service berisi logika bisnis pembayaran.
type Service struct {
	repo  *Repository
	mail  *mailer.Mailer
	files ProofStorage
}

// NewService membuat Service. mail boleh nonaktif (SMTP belum diisi); files
// adalah penyimpanan bukti pembayaran di Cloudinary (SDD D-07).
func NewService(repo *Repository, mail *mailer.Mailer, files ProofStorage) *Service {
	return &Service{repo: repo, mail: mail, files: files}
}

// Get mengambil pembayaran order. Hanya pemilik order atau admin; selain itu
// dianggap tidak ada supaya keberadaan order tidak bocor.
func (s *Service) Get(ctx context.Context, orderID, userID string, admin bool) (Payment, error) {
	if err := s.checkAccess(ctx, orderID, userID, admin); err != nil {
		return Payment{}, err
	}
	return s.repo.GetByOrder(ctx, orderID)
}

// Create membuat tagihan pembayaran untuk order milik userID. Kode unik
// ditambahkan ke total agar transfer mudah dicocokkan admin.
func (s *Service) Create(ctx context.Context, orderID, userID, method string) (Payment, error) {
	p, err := s.repo.Create(ctx, orderID, userID, method)
	if errors.Is(err, ErrAlreadyExists) {
		// Tagihan sudah ada; membuka info pembayaran mengembalikannya (SDD D-06).
		return s.repo.GetByOrder(ctx, orderID)
	}
	return p, err
}

// UploadProof menyimpan bukti pembayaran milik order userID, lalu memberi
// tahu pemilik toko lewat email bahwa ada bukti yang perlu diverifikasi.
func (s *Service) UploadProof(ctx context.Context, orderID, userID string, proof Proof) (Payment, error) {
	if err := s.checkAccess(ctx, orderID, userID, false); err != nil {
		return Payment{}, err
	}
	if !s.files.Enabled() {
		return Payment{}, ErrStorageDisabled
	}
	if err := s.repo.CheckProofUpload(ctx, orderID); err != nil {
		return Payment{}, err
	}
	proofURL, err := s.files.UploadProof(ctx, proof.Type, proof.Data)
	if err != nil {
		if errors.Is(err, cloudinary.ErrDisabled) {
			return Payment{}, ErrStorageDisabled
		}
		return Payment{}, fmt.Errorf("payment: upload proof: %w", err)
	}
	p, err := s.repo.SaveProof(ctx, orderID, proofURL)
	if err != nil {
		return Payment{}, err
	}

	if c, err := s.repo.contactOf(ctx, orderID); err == nil {
		s.mail.NotifyOwnerEmail("Bukti pembayaran baru: "+c.OrderNumber, mailer.Email{
			Heading:    "Bukti pembayaran perlu diverifikasi",
			Paragraphs: []string{c.Name + " mengunggah bukti pembayaran QRIS. Cocokkan nominal di mutasi QRIS sebelum menyetujui."},
			Details: [][2]string{
				{"No. order", c.OrderNumber},
				{"Nominal harus diterima", "Rp " + p.Amount},
				{"Kode unik", fmt.Sprintf("%03d", p.UniqueCode)},
			},
			Note:       "Bukti pembayaran terlampir di email ini.",
			Attachment: &mailer.Attachment{Name: "bukti-" + c.OrderNumber + proofExt(proof.Type), Type: proof.Type, Data: proof.Data},
			Button: &mailer.Link{Label: "Buka antrean verifikasi", URL: s.mail.AppURL("/admin/pembayaran")},
		})
	} else {
		slog.WarnContext(ctx, "payment: load contact for email failed", "error", err)
	}
	return p, nil
}

// Proof mengambil berkas bukti pembayaran (pemilik order atau admin).
func (s *Service) Proof(ctx context.Context, orderID, userID string, admin bool) (Proof, error) {
	if err := s.checkAccess(ctx, orderID, userID, admin); err != nil {
		return Proof{}, err
	}
	ref, err := s.repo.GetProof(ctx, orderID)
	if err != nil {
		return Proof{}, err
	}
	// Berkas diambil server dari Cloudinary lalu diteruskan, sehingga alamat
	// Cloudinary tidak pernah sampai ke klien dan hak akses tetap diperiksa.
	data, err := s.files.Fetch(ctx, ref.URL, maxProofBytes)
	if err != nil {
		return Proof{}, fmt.Errorf("payment: fetch proof: %w", err)
	}
	return Proof{Data: data, Type: ref.Type}, nil
}

// List mengambil satu halaman pembayaran (admin).
func (s *Service) List(ctx context.Context, f Filter, limit, offset int) ([]AdminItem, int, error) {
	return s.repo.List(ctx, f, limit, offset)
}

// Review menyetujui atau menolak pembayaran, lalu mengirim email hasilnya ke
// pelanggan.
func (s *Service) Review(ctx context.Context, paymentID, adminID, status string) (Payment, error) {
	p, err := s.repo.Review(ctx, paymentID, adminID, status)
	if err != nil {
		return Payment{}, err
	}

	if c, err := s.repo.contactOf(ctx, p.OrderID); err == nil {
		orderURL := s.mail.AppURL("/pesanan/" + p.OrderID)
		e := mailer.Email{
			Heading: "Pembayaran diterima",
			Paragraphs: []string{
				"Halo " + c.Name + ", pembayaran untuk pesanan " + c.OrderNumber + " sudah kami verifikasi.",
				"Pesanan masuk antrean pengemasan. Nomor resi akan dikirim lewat email begitu paket diserahkan ke kurir.",
			},
			Details: [][2]string{{"No. order", c.OrderNumber}, {"Nominal", "Rp " + p.Amount}},
			Button:  &mailer.Link{Label: "Lihat pesanan", URL: orderURL},
		}
		subject := "Pembayaran diterima: " + c.OrderNumber
		if status == StatusRejected {
			subject = "Bukti pembayaran ditolak: " + c.OrderNumber
			e = mailer.Email{
				Heading: "Bukti pembayaran belum valid",
				Paragraphs: []string{
					"Halo " + c.Name + ", bukti pembayaran untuk pesanan " + c.OrderNumber + " belum dapat kami cocokkan dengan mutasi QRIS.",
					"Pastikan nominal yang dibayar sama persis (termasuk 3 digit kode unik), lalu unggah ulang bukti yang jelas sebelum batas waktu pembayaran.",
				},
				Details: [][2]string{{"No. order", c.OrderNumber}, {"Nominal yang benar", "Rp " + p.Amount}},
				Button:  &mailer.Link{Label: "Unggah ulang bukti", URL: orderURL},
			}
		}
		s.mail.NotifyEmail(c.Email, subject, e)
	} else {
		slog.WarnContext(ctx, "payment: load contact for email failed", "error", err)
	}
	return p, nil
}

// checkAccess memastikan order ada dan milik userID (kecuali admin).
func (s *Service) checkAccess(ctx context.Context, orderID, userID string, admin bool) error {
	owner, err := s.repo.OrderOwner(ctx, orderID)
	if err != nil {
		return err
	}
	if !admin && owner != userID {
		return ErrOrderNotFound
	}
	return nil
}

func proofExt(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	}
	return ".jpg"
}
