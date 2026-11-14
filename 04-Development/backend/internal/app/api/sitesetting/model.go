// Package sitesetting menyimpan pengaturan situs (maintenance & QRIS toko).
package sitesetting

// Settings adalah pengaturan situs (satu baris). Data QRIS bersifat publik
// karena ditampilkan ke pelanggan di halaman pembayaran.
type Settings struct {
	MaintenanceMode    bool    `json:"maintenance_mode"`
	MaintenanceMessage string  `json:"maintenance_message"`
	QRISImageURL       *string `json:"qris_image_url"`
	QRISMerchantName   *string `json:"qris_merchant_name"`
	QRISNMID           *string `json:"qris_nmid"`
	// PaymentWindowHours adalah batas waktu bayar order baru (jam).
	PaymentWindowHours int `json:"payment_window_hours"`
}

// updateRequest: semua field opsional; field yang tidak dikirim tidak berubah.
// String kosong pada field QRIS menghapus nilainya.
type updateRequest struct {
	MaintenanceMode    *bool   `json:"maintenance_mode"`
	MaintenanceMessage *string `json:"maintenance_message"`
	QRISImageURL       *string `json:"qris_image_url"`
	QRISMerchantName   *string `json:"qris_merchant_name"`
	QRISNMID           *string `json:"qris_nmid"`
	PaymentWindowHours *int    `json:"payment_window_hours"`
}
