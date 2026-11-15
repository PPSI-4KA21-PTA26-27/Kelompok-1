package shipment

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// Handler melayani endpoint /orders/{orderId}/shipment.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Get melayani GET /orders/{orderId}/shipment (pemilik order atau admin).
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderId")
	if !validator.UUID(orderID) {
		response.NotFound(w)
		return
	}
	p, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		response.LoginRequired(w)
		return
	}

	s, err := h.svc.Get(r.Context(), orderID, p.UserID, user.IsBackoffice(p.Role))
	if h.writeError(w, r, "get", err) {
		return
	}
	response.JSON(w, http.StatusOK, s)
}

// Upsert melayani PUT /orders/{orderId}/shipment (admin). Menjawab 201 bila
// pengiriman baru dibuat dan 200 bila diperbarui.
func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderId")
	if !validator.UUID(orderID) {
		response.NotFound(w)
		return
	}

	var req upsertRequest
	if !response.Decode(w, r, &req) {
		return
	}
	s := Shipment{
		OrderID:        orderID,
		CoverageStatus: strings.TrimSpace(req.CoverageStatus),
		ShippingStatus: strings.TrimSpace(req.ShippingStatus),
	}
	if req.TrackingNumber != nil {
		if t := strings.TrimSpace(*req.TrackingNumber); t != "" {
			s.TrackingNumber = &t
		}
	}

	var c validator.Collector
	c.Check(s.TrackingNumber == nil || validator.RuneLenBetween(*s.TrackingNumber, 1, 100),
		"tracking_number", "tracking_number maksimal 100 karakter.")
	c.Check(s.CoverageStatus == CoverageCovered || s.CoverageStatus == CoverageNotCovered,
		"coverage_status", "coverage_status harus covered atau not_covered.")
	c.Check(validShipping(s.ShippingStatus), "shipping_status",
		"shipping_status harus pending, shipped, delivered, atau returned.")
	c.Check(s.ShippingStatus != ShippingShipped || (s.TrackingNumber != nil && *s.TrackingNumber != ""),
		"tracking_number", "tracking_number wajib diisi bila shipping_status shipped.")
	if errs := c.Errors(); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	out, created, err := h.svc.Upsert(r.Context(), s)
	if h.writeError(w, r, "upsert", err) {
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	response.JSON(w, status, out)
}

// writeError menerjemahkan error service menjadi response. Mengembalikan true
// bila err tidak nil (response sudah ditulis).
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrOrderNotFound):
		response.NotFound(w)
	case errors.Is(err, ErrOrderNotShippable):
		response.Conflict(w, "Pesanan belum dibayar atau sudah dibatalkan.")
	default:
		slog.ErrorContext(r.Context(), op+" shipment failed", "error", err)
		response.Internal(w)
	}
	return true
}

func validShipping(s string) bool {
	switch s {
	case ShippingPending, ShippingShipped, ShippingDelivered, ShippingReturned:
		return true
	}
	return false
}
