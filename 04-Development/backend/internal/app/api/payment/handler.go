package payment

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/fileupload"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// maxProofBytes adalah batas ukuran bukti pembayaran (2 MB, sesuai openapi).
const maxProofBytes = 2 << 20

// proofTypes adalah tipe bukti yang diterima, dicek dari isi berkas.
var proofTypes = []string{"image/jpeg", "image/png", "image/webp", "application/pdf"}

// Handler melayani endpoint pembayaran.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Get melayani GET /orders/{orderId}/payment (pemilik order atau admin).
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	orderID, uid, admin, ok := orderRequest(w, r)
	if !ok {
		return
	}
	p, err := h.svc.Get(r.Context(), orderID, uid, admin)
	if h.writeError(w, r, "get", err) {
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// Create melayani POST /orders/{orderId}/payment (pemilik order).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	orderID, uid, _, ok := orderRequest(w, r)
	if !ok {
		return
	}

	var req createRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.Method = strings.TrimSpace(req.Method)
	if req.Method == "" {
		req.Method = MethodQRIS
	}
	if req.Method != MethodQRIS {
		response.ValidationFailed(w, response.FieldError{Field: "method",
			Message: "Pembayaran hanya melalui QRIS (method: " + MethodQRIS + ")."})
		return
	}

	p, err := h.svc.Create(r.Context(), orderID, uid, req.Method)
	if h.writeError(w, r, "create", err) {
		return
	}
	response.JSON(w, http.StatusCreated, p)
}

// UploadProof melayani PUT /orders/{orderId}/payment/proof (pemilik order).
// Body multipart/form-data dengan field "file".
func (h *Handler) UploadProof(w http.ResponseWriter, r *http.Request) {
	orderID, uid, _, ok := orderRequest(w, r)
	if !ok {
		return
	}

	data, typ, ok := fileupload.Read(w, r, "file", maxProofBytes, proofTypes)
	if !ok {
		return
	}

	p, err := h.svc.UploadProof(r.Context(), orderID, uid, Proof{Data: data, Type: typ})
	if h.writeError(w, r, "upload proof", err) {
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// GetProof melayani GET /orders/{orderId}/payment/proof: mengirim berkas apa
// adanya (diambil dari Cloudinary) dengan Content-Type sesuai proof_type.
func (h *Handler) GetProof(w http.ResponseWriter, r *http.Request) {
	orderID, uid, admin, ok := orderRequest(w, r)
	if !ok {
		return
	}

	pr, err := h.svc.Proof(r.Context(), orderID, uid, admin)
	if h.writeError(w, r, "get proof", err) {
		return
	}

	w.Header().Set("Content-Type", pr.Type)
	w.Header().Set("Content-Length", strconv.Itoa(len(pr.Data)))
	// Cegah browser menebak tipe lain dan jangan simpan berkas di cache bersama.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(pr.Data); err != nil {
		slog.DebugContext(r.Context(), "write payment proof", "error", err)
	}
}

// AdminList melayani GET /admin/payments (admin).
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := response.ParsePage(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	f := Filter{Status: q.Get("status"), OrderID: q.Get("order_id"), Q: strings.TrimSpace(q.Get("q"))}
	if f.Status != "" && !validStatus(f.Status) {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter status tidak valid.")
		return
	}
	if f.OrderID != "" && !validator.UUID(f.OrderID) {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter order_id tidak valid.")
		return
	}

	if len([]rune(f.Q)) > 100 {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter q maksimal 100 karakter.")
		return
	}

	items, total, err := h.svc.List(r.Context(), f, limit, (page-1)*limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "list payments failed", "error", err)
		response.Internal(w)
		return
	}
	response.List(w, http.StatusOK, items, response.Meta{Page: page, Limit: limit, Total: total})
}

// AdminReview melayani POST /admin/payments/{paymentId}/review (admin).
// reviewed_by diambil dari token admin, bukan dari body.
func (h *Handler) AdminReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("paymentId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}
	uid, _ := user.Principal(r)

	var req reviewRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.Status = strings.TrimSpace(req.Status)
	if req.Status != StatusApproved && req.Status != StatusRejected {
		response.ValidationFailed(w, response.FieldError{Field: "status",
			Message: "status harus " + StatusApproved + " atau " + StatusRejected + "."})
		return
	}

	p, err := h.svc.Review(r.Context(), id, uid, req.Status)
	if h.writeError(w, r, "review", err) {
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// writeError menerjemahkan error service menjadi response. Mengembalikan true
// bila err tidak nil (response sudah ditulis).
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrOrderNotFound):
		response.NotFound(w)
	case errors.Is(err, ErrAlreadyExists):
		response.Conflict(w, "Order ini sudah memiliki pembayaran.")
	case errors.Is(err, ErrOrderNotPayable):
		response.Conflict(w, "Order tidak sedang menunggu pembayaran.")
	case errors.Is(err, ErrAlreadyApproved):
		response.Conflict(w, "Pembayaran sudah disetujui sehingga bukti tidak bisa diganti.")
	case errors.Is(err, ErrNoProof):
		response.Conflict(w, "Bukti pembayaran belum diunggah.")
	case errors.Is(err, ErrAlreadyReviewed):
		response.Conflict(w, "Pembayaran sudah diverifikasi sebelumnya.")
	case errors.Is(err, ErrStorageDisabled):
		response.Error(w, http.StatusServiceUnavailable, "upload_unavailable",
			"Layanan penyimpanan bukti pembayaran belum dikonfigurasi.")
	default:
		slog.ErrorContext(r.Context(), op+" payment failed", "error", err)
		response.Internal(w)
	}
	return true
}

// orderRequest membaca orderId dari path dan identitas pemanggil dari token.
// ok false berarti response error sudah ditulis.
func orderRequest(w http.ResponseWriter, r *http.Request) (orderID, userID string, admin, ok bool) {
	orderID = r.PathValue("orderId")
	if !validator.UUID(orderID) {
		response.NotFound(w)
		return "", "", false, false
	}
	userID, admin = user.Principal(r)
	if userID == "" {
		response.LoginRequired(w)
		return "", "", false, false
	}
	return orderID, userID, admin, true
}

func validStatus(s string) bool {
	return s == StatusPending || s == StatusApproved || s == StatusRejected
}
