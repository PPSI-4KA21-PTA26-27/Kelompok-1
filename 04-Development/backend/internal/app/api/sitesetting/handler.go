package sitesetting

import (
	"log/slog"
	"net/http"
	"strings"

	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// Handler melayani /site-settings.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Get melayani GET /site-settings (publik).
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Get(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "get site settings failed", "error", err)
		response.Internal(w)
		return
	}
	response.JSON(w, http.StatusOK, s)
}

// Update melayani PUT /site-settings (admin). Hanya field yang dikirim yang
// diubah.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateRequest
	if !response.Decode(w, r, &req) {
		return
	}
	var c validator.Collector
	trim := func(p *string) *string {
		if p == nil {
			return nil
		}
		v := strings.TrimSpace(*p)
		return &v
	}
	msg := trim(req.MaintenanceMessage)
	qris, merchant, nmid := trim(req.QRISImageURL), trim(req.QRISMerchantName), trim(req.QRISNMID)
	c.Check(msg == nil || validator.RuneLenBetween(*msg, 1, 500), "maintenance_message",
		"Pesan maintenance wajib diisi, maksimal 500 karakter.")
	c.Check(qris == nil || *qris == "" || validator.HTTPURL(*qris), "qris_image_url", "Gambar QRIS harus URL http/https.")
	c.Check(merchant == nil || validator.RuneLenBetween(*merchant, 0, 100), "qris_merchant_name", "Nama merchant maksimal 100 karakter.")
	c.Check(nmid == nil || validator.RuneLenBetween(*nmid, 0, 50), "qris_nmid", "NMID maksimal 50 karakter.")
	c.Check(req.PaymentWindowHours == nil || (*req.PaymentWindowHours >= 1 && *req.PaymentWindowHours <= 168),
		"payment_window_hours", "Batas waktu bayar 1 sampai 168 jam.")
	c.Check(req != (updateRequest{}), "body", "Isi minimal satu pengaturan.")
	if errs := c.Errors(); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	orNil := func(p *string) *string {
		if p == nil || *p == "" {
			return nil
		}
		return p
	}
	s, err := h.svc.Update(r.Context(), func(s *Settings) {
		if req.MaintenanceMode != nil {
			s.MaintenanceMode = *req.MaintenanceMode
		}
		if msg != nil {
			s.MaintenanceMessage = *msg
		}
		if qris != nil {
			s.QRISImageURL = orNil(qris)
		}
		if merchant != nil {
			s.QRISMerchantName = orNil(merchant)
		}
		if nmid != nil {
			s.QRISNMID = orNil(nmid)
		}
		if req.PaymentWindowHours != nil {
			s.PaymentWindowHours = *req.PaymentWindowHours
		}
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "update site settings failed", "error", err)
		response.Internal(w)
		return
	}
	response.JSON(w, http.StatusOK, s)
}
