package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"hii-stranger1/backend/repository"
	"hii-stranger1/backend/response"
	"hii-stranger1/backend/services"
)

// Alur: dipakai semua handler di paket ini. decodeJSON membaca request,
// writeServiceError menerjemahkan error dari services/repository menjadi respons HTTP.

// maxBodyBytes membatasi isi request JSON. 16 KB sudah sangat cukup untuk form auth/akun.
const maxBodyBytes = 16 << 10

// decodeJSON membaca body JSON ke dst. Kalau gagal, respons error sudah dikirim dan hasilnya false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.Error(w, http.StatusRequestEntityTooLarge, "data yang dikirim terlalu besar")
			return false
		}
		response.Error(w, http.StatusBadRequest, "format data tidak valid")
		return false
	}
	return true
}

// writeServiceError memilih status HTTP yang sesuai. Error yang tidak dikenali dicatat di log
// dan klien hanya menerima pesan umum (tanpa detail SQL atau stack trace).
func writeServiceError(w http.ResponseWriter, err error) {
	var validation *services.ValidationError
	switch {
	case errors.As(err, &validation):
		response.ValidationError(w, validation.Fields)
	case errors.Is(err, services.ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, services.ErrInvalidResetToken):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrEmailTaken):
		response.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		response.Error(w, http.StatusNotFound, "data tidak ditemukan")
	default:
		log.Printf("error tidak terduga: %v", err)
		response.Error(w, http.StatusInternalServerError, "terjadi kesalahan pada server")
	}
}
