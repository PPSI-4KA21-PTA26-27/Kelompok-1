package handlers

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"

	"hii-stranger1/backend/response"
	"hii-stranger1/backend/services"
)

// Alur: routes (/api/v1/uploads/{name}) -> UploadHandler -> services.ImageStore -> folder uploads/.
// Folder uploads/ tidak dibuka langsung ke publik. Gambar hanya keluar lewat handler ini,
// dan hanya untuk nama file yang berbentuk nama buatan ImageStore.

type UploadHandler struct {
	store *services.ImageStore
}

func NewUploadHandler(store *services.ImageStore) *UploadHandler {
	return &UploadHandler{store: store}
}

// Serve: GET /api/v1/uploads/{name}
func (h *UploadHandler) Serve(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	path, ok := h.store.Path(name)
	if !ok {
		response.Error(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}

	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		response.Error(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	if err != nil {
		log.Printf("upload: buka file gagal: %v", err)
		response.Error(w, http.StatusInternalServerError, "terjadi kesalahan pada server")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		log.Printf("upload: baca info file gagal: %v", err)
		response.Error(w, http.StatusInternalServerError, "terjadi kesalahan pada server")
		return
	}

	// nosniff: browser tidak boleh menebak tipe sendiri. Nama file acak dan isinya tidak pernah
	// berubah, jadi aman di-cache lama.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, name, info.ModTime(), f)
}
