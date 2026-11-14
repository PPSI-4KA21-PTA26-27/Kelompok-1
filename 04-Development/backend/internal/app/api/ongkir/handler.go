package ongkir

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// Handler melayani endpoint /regions dan /shipping/costs.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Provinces melayani GET /regions/provinces.
func (h *Handler) Provinces(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Provinces(r.Context())
	h.writeList(w, r, list, err)
}

// Cities melayani GET /regions/provinces/{provinceId}/cities.
func (h *Handler) Cities(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "provinceId")
	if !ok {
		return
	}
	list, err := h.svc.Cities(r.Context(), id)
	h.writeList(w, r, list, err)
}

// Districts melayani GET /regions/cities/{cityId}/districts (kecamatan).
func (h *Handler) Districts(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "cityId")
	if !ok {
		return
	}
	list, err := h.svc.Districts(r.Context(), id)
	h.writeList(w, r, list, err)
}

// SubDistricts melayani GET /regions/districts/{districtId}/subdistricts (kelurahan).
func (h *Handler) SubDistricts(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "districtId")
	if !ok {
		return
	}
	list, err := h.svc.SubDistricts(r.Context(), id)
	h.writeList(w, r, list, err)
}

// Costs melayani GET /shipping/costs?subdistrict_id=...&cart_item_ids=a,b
// (login). Berat dihitung dari keranjang user, lalu dikembalikan ongkir
// layanan termurah dari kurir toko ke kelurahan tujuan.
func (h *Handler) Costs(w http.ResponseWriter, r *http.Request) {
	p, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		response.LoginRequired(w)
		return
	}

	q := r.URL.Query()
	sub, err := strconv.Atoi(q.Get("subdistrict_id"))
	if err != nil || sub < 1 {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter subdistrict_id wajib berupa angka.")
		return
	}
	var ids []string
	if raw := strings.TrimSpace(q.Get("cart_item_ids")); raw != "" {
		for _, id := range strings.Split(raw, ",") {
			id = strings.TrimSpace(id)
			if !validator.UUID(id) {
				response.Error(w, http.StatusBadRequest, "bad_request", "Parameter cart_item_ids tidak valid.")
				return
			}
			ids = append(ids, id)
		}
	}

	quote, err := h.svc.QuoteCart(r.Context(), p.UserID, ids, sub)
	if WriteError(w, r, "quote", err) {
		return
	}
	response.JSON(w, http.StatusOK, quote)
}

func (h *Handler) writeList(w http.ResponseWriter, r *http.Request, list any, err error) {
	if WriteError(w, r, "list regions", err) {
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// WriteError menerjemahkan error ongkir menjadi response. Diekspor agar
// checkout di package order memakai pesan yang sama. Mengembalikan true bila
// err tidak nil (response sudah ditulis).
func WriteError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrDisabled):
		response.Error(w, http.StatusServiceUnavailable, "service_unavailable",
			"Layanan cek ongkir belum dikonfigurasi.")
	case errors.Is(err, ErrInvalidDestination):
		response.Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.",
			response.FieldError{Field: "subdistrict_id",
				Message: "Provinsi, kota, kecamatan, dan kelurahan tidak saling cocok."})
	case errors.Is(err, ErrNoRoute):
		response.Error(w, http.StatusNotFound, "not_found", "Tujuan belum terjangkau pengiriman.")
	case errors.Is(err, ErrCartEmpty):
		response.Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.",
			response.FieldError{Field: "cart_item_ids", Message: "Keranjang kosong atau ada item yang tidak ditemukan."})
	case strings.HasPrefix(err.Error(), "rajaongkir:"):
		// Gagal menghubungi RajaOngkir: jangan bocorkan detailnya ke client.
		slog.ErrorContext(r.Context(), op+" ongkir failed", "error", err)
		response.Error(w, http.StatusBadGateway, "upstream_error", "Gagal mengambil data ongkir. Coba lagi.")
	default:
		slog.ErrorContext(r.Context(), op+" ongkir failed", "error", err)
		response.Internal(w)
	}
	return true
}

// pathID membaca ID wilayah (angka positif) dari path.
func pathID(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue(name))
	if err != nil || id < 1 {
		response.NotFound(w)
		return 0, false
	}
	return id, true
}
