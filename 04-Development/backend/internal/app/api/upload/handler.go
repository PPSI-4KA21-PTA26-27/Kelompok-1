// Package upload menerima unggahan gambar dari admin dan menyimpannya di
// Cloudinary. URL hasilnya dipakai sebagai image_url pada gambar produk dan
// banner (endpoint tersebut tetap menerima URL sesuai openapi.yaml).
package upload

import (
	"log/slog"
	"net/http"

	"hii-stranger1/backend/internal/pkg/cloudinary"
	"hii-stranger1/backend/internal/pkg/fileupload"
	"hii-stranger1/backend/internal/pkg/response"
)

// maxImageBytes adalah batas ukuran satu gambar yang diunggah (5 MB).
const maxImageBytes = 5 << 20

// imageTypes adalah tipe gambar yang diterima, dicek dari isi berkas.
var imageTypes = []string{"image/jpeg", "image/png", "image/webp"}

// Handler melayani endpoint /admin/uploads.
type Handler struct {
	cdn *cloudinary.Client
}

// NewHandler membuat Handler.
func NewHandler(cdn *cloudinary.Client) *Handler {
	return &Handler{cdn: cdn}
}

// Result adalah isi response unggahan.
type Result struct {
	URL string `json:"url"`
}

// Image melayani POST /admin/uploads/images (admin). Body multipart/form-data
// dengan field "file"; response berisi URL https dari Cloudinary.
func (h *Handler) Image(w http.ResponseWriter, r *http.Request) {
	if !h.cdn.Enabled() {
		response.Error(w, http.StatusServiceUnavailable, "service_unavailable",
			"Layanan unggah gambar belum dikonfigurasi.")
		return
	}

	data, _, ok := fileupload.Read(w, r, "file", maxImageBytes, imageTypes)
	if !ok {
		return
	}

	url, err := h.cdn.UploadImage(r.Context(), "image", data)
	if err != nil {
		slog.ErrorContext(r.Context(), "upload image failed", "error", err)
		response.Error(w, http.StatusBadGateway, "upload_failed", "Gagal mengunggah gambar. Coba lagi.")
		return
	}

	response.JSON(w, http.StatusCreated, Result{URL: url})
}
