// Package response menyeragamkan bentuk response JSON dan pembacaan body request
// sesuai kontrak di api/openapi.yaml.
package response

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 100

	// maxPage menjaga (page-1)*limit tidak overflow dan OFFSET tetap wajar.
	maxPage = 100000
)

// Meta adalah informasi pagination pada response daftar.
type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// FieldError menjelaskan satu kesalahan validasi pada satu field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type envelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// JSON menulis response sukses {"data": ...}.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{Data: data})
}

// List menulis response daftar {"data": [...], "meta": {...}}.
func List(w http.ResponseWriter, status int, data any, meta Meta) {
	write(w, status, envelope{Data: data, Meta: &meta})
}

// NoContent menulis response 204 tanpa body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error menulis response error {"error": {...}}. Pesan harus aman ditampilkan
// ke client: jangan masukkan detail internal.
func Error(w http.ResponseWriter, status int, code, message string, details ...FieldError) {
	write(w, status, errorBody{Error: errorDetail{Code: code, Message: message, Details: details}})
}

// Internal menulis response 500 standar. Penyebab sebenarnya harus dicatat
// oleh pemanggil, bukan dikirim ke client.
func Internal(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal_error", "Terjadi kesalahan pada server.")
}

// Decode membaca satu objek JSON dari body ke dst. Jika gagal, response error
// (400 atau 413) sudah ditulis dan Decode mengembalikan false.
func Decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			Error(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Ukuran request melebihi batas.")
			return false
		}
		Error(w, http.StatusBadRequest, "bad_request", "Format request tidak valid.")
		return false
	}
	// Body harus berisi tepat satu objek JSON.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		Error(w, http.StatusBadRequest, "bad_request", "Format request tidak valid.")
		return false
	}
	return true
}

// ParsePage membaca query page (mulai 1, default 1) dan limit (1 sampai 100,
// default 20). Jika salah satunya tidak valid, response 400 sudah ditulis dan
// ParsePage mengembalikan ok false.
func ParsePage(w http.ResponseWriter, r *http.Request) (page, limit int, ok bool) {
	page, limit = 1, defaultLimit
	q := r.URL.Query()

	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxPage {
			badPagination(w)
			return 0, 0, false
		}
		page = n
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxLimit {
			badPagination(w)
			return 0, 0, false
		}
		limit = n
	}
	return page, limit, true
}

func badPagination(w http.ResponseWriter) {
	Error(w, http.StatusBadRequest, "bad_request", "Parameter page atau limit tidak valid.")
}

func write(w http.ResponseWriter, status int, body any) {
	payload, err := json.Marshal(body)
	if err != nil {
		slog.Error("response: marshal body", "error", err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		// Client mungkin sudah terputus; tidak ada yang bisa dilakukan bila gagal.
		_, _ = w.Write([]byte(`{"error":{"code":"internal_error","message":"Terjadi kesalahan pada server."}}`))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(payload); err != nil {
		slog.Debug("response: write body", "error", err)
	}
}

// Helper error umum agar handler tidak menduplikasi.
func ValidationFailed(w http.ResponseWriter, errs ...FieldError) {
	Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.", errs...)
}

func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "not_found", "Data tidak ditemukan.")
}

func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "forbidden", "Anda tidak memiliki akses.")
}

func LoginRequired(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "unauthorized", "Anda harus login terlebih dahulu.")
}

func Conflict(w http.ResponseWriter, msg string) {
	Error(w, http.StatusConflict, "conflict", msg)
}
