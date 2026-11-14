// Package fileupload membaca satu berkas dari request multipart/form-data
// dengan batas ukuran dan pemeriksaan tipe dari isi berkas.
package fileupload

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"hii-stranger1/backend/internal/pkg/response"
)

// multipartOverhead adalah ruang tambahan untuk boundary dan header part,
// di luar ukuran berkas itu sendiri.
const multipartOverhead = 64 << 10

// Read membaca part bernama field dari body multipart. Tipe berkas ditentukan
// dari isinya (http.DetectContentType), bukan dari nama file atau header, dan
// harus ada di allowed. Bila gagal, response error (400, 413, atau 415) sudah
// ditulis dan ok bernilai false.
//
// Global middleware.MaxBody melewati request multipart, jadi batas ukuran
// dipasang di sini sesuai kebutuhan masing-masing endpoint.
func Read(w http.ResponseWriter, r *http.Request, field string, maxBytes int64, allowed []string) (data []byte, contentType string, ok bool) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		response.Error(w, http.StatusBadRequest, "bad_request",
			"Request harus multipart/form-data dengan field "+field+".")
		return nil, "", false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+multipartOverhead)
	mr, err := r.MultipartReader()
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "Format multipart tidak valid.")
		return nil, "", false
	}

	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			response.Error(w, http.StatusBadRequest, "bad_request", "Field "+field+" wajib diisi.")
			return nil, "", false
		}
		if err != nil {
			writeReadError(w, err)
			return nil, "", false
		}
		if part.FormName() != field {
			_ = part.Close()
			continue
		}

		// Baca satu byte lebih dari batas untuk tahu apakah berkas kebesaran.
		data, err = io.ReadAll(io.LimitReader(part, maxBytes+1))
		_ = part.Close()
		if err != nil {
			writeReadError(w, err)
			return nil, "", false
		}
		if int64(len(data)) > maxBytes {
			tooLarge(w)
			return nil, "", false
		}
		if len(data) == 0 {
			response.Error(w, http.StatusBadRequest, "bad_request", "Berkas tidak boleh kosong.")
			return nil, "", false
		}

		// DetectContentType bisa menambahkan parameter, mis. "; charset=utf-8".
		contentType, _, _ = strings.Cut(http.DetectContentType(data), ";")
		for _, a := range allowed {
			if contentType == a {
				return data, contentType, true
			}
		}
		response.Error(w, http.StatusUnsupportedMediaType, "unsupported_media_type",
			"Tipe berkas tidak didukung. Gunakan "+strings.Join(allowed, ", ")+".")
		return nil, "", false
	}
}

func writeReadError(w http.ResponseWriter, err error) {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		tooLarge(w)
		return
	}
	response.Error(w, http.StatusBadRequest, "bad_request", "Berkas gagal dibaca.")
}

func tooLarge(w http.ResponseWriter) {
	response.Error(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Ukuran berkas melebihi batas.")
}
