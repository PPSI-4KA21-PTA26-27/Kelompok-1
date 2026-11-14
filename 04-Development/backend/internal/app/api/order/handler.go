package order

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"hii-stranger1/backend/internal/app/api/ongkir"
	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// maxCartItemIDs membatasi jumlah cart_item_ids dalam satu checkout.
const maxCartItemIDs = 100

// Handler melayani endpoint /orders dan /admin/orders.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListMine melayani GET /orders: order milik user yang sedang login.
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	uid, _ := user.Principal(r)
	if uid == "" {
		response.LoginRequired(w)
		return
	}
	f, ok := parseFilter(w, r)
	if !ok {
		return
	}
	f.UserID = uid // user biasa selalu hanya melihat ordernya sendiri
	h.list(w, r, f)
}

// AdminList melayani GET /admin/orders (admin), dengan filter user_id dan
// order_number.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	f, ok := parseFilter(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	f.UserID = q.Get("user_id")
	f.Q = strings.TrimSpace(q.Get("q"))
	if f.Q == "" {
		f.Q = strings.TrimSpace(q.Get("order_number"))
	}
	if len([]rune(f.Q)) > 100 {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter q maksimal 100 karakter.")
		return
	}
	if f.UserID != "" && !validator.UUID(f.UserID) {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter user_id tidak valid.")
		return
	}
	h.list(w, r, f)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, f Filter) {
	page, limit, ok := response.ParsePage(w, r)
	if !ok {
		return
	}
	orders, total, err := h.svc.List(r.Context(), f, limit, (page-1)*limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "list orders failed", "error", err)
		response.Internal(w)
		return
	}
	response.List(w, http.StatusOK, orders, response.Meta{Page: page, Limit: limit, Total: total})
}

// Create melayani POST /orders: checkout dari keranjang.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := user.Principal(r)
	if uid == "" {
		response.LoginRequired(w)
		return
	}

	var req createRequest
	if !response.Decode(w, r, &req) {
		return
	}
	in, errs := req.input()
	if len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	d, err := h.svc.Checkout(r.Context(), uid, in)
	if h.writeError(w, r, "create", err) {
		return
	}
	response.JSON(w, http.StatusCreated, d)
}

// Get melayani GET /orders/{orderId} (pemilik order atau admin).
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("orderId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}
	uid, admin := user.Principal(r)
	if uid == "" {
		response.LoginRequired(w)
		return
	}

	d, err := h.svc.Get(r.Context(), id, uid, admin)
	if h.writeError(w, r, "get", err) {
		return
	}
	response.JSON(w, http.StatusOK, d)
}

// AdminUpdateStatus melayani PUT /admin/orders/{orderId}/status (admin).
// Saat mengubah ke shipped, admin wajib mengisi tracking_number (no. resi);
// kurir mengikuti pilihan saat checkout. Pelanggan lalu menerima email berisi
// status, kurir, dan no. resi.
func (h *Handler) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("orderId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	var req statusRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.OrderStatus = strings.TrimSpace(req.OrderStatus)

	var c validator.Collector
	c.Check(validStatus(req.OrderStatus), "order_status", "order_status tidak dikenal.")
	var reason *string
	if req.Reason != nil {
		if v := strings.TrimSpace(*req.Reason); v != "" {
			reason = &v
		}
	}
	c.Check(reason == nil || validator.RuneLenBetween(*reason, 1, 300), "reason", "Alasan maksimal 300 karakter.")
	var ship *ShipInfo
	if req.OrderStatus == StatusShipped {
		info := ShipInfo{}
		if req.TrackingNumber != nil {
			info.TrackingNumber = strings.TrimSpace(*req.TrackingNumber)
		}
		c.Check(validator.RuneLenBetween(info.TrackingNumber, 1, 100), "tracking_number",
			"No. resi wajib diisi saat order dikirim, maksimal 100 karakter.")
		ship = &info
	}
	if errs := c.Errors(); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	o, err := h.svc.UpdateStatus(r.Context(), id, req.OrderStatus, ship, reason)
	if h.writeError(w, r, "update status", err) {
		return
	}
	response.JSON(w, http.StatusOK, o)
}

// writeError menerjemahkan error service menjadi response. Mengembalikan true
// bila err tidak nil (response sudah ditulis).
// Cancel melayani POST /orders/{orderId}/cancel: pelanggan membatalkan
// ordernya sendiri selama masih menunggu pembayaran.
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("orderId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}
	uid, _ := user.Principal(r)
	if uid == "" {
		response.LoginRequired(w)
		return
	}
	var req cancelRequest
	if r.ContentLength != 0 && !response.Decode(w, r, &req) {
		return
	}
	var reason *string
	if req.Reason != nil {
		if v := strings.TrimSpace(*req.Reason); v != "" {
			if !validator.RuneLenBetween(v, 1, 300) {
				response.ValidationFailed(w, response.FieldError{Field: "reason", Message: "Alasan maksimal 300 karakter."})
				return
			}
			reason = &v
		}
	}
	o, err := h.svc.Cancel(r.Context(), id, uid, reason)
	if h.writeError(w, r, "cancel", err) {
		return
	}
	response.JSON(w, http.StatusOK, o)
}

// Complete melayani POST /orders/{orderId}/complete: pelanggan konfirmasi
// paket sudah diterima.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("orderId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}
	uid, _ := user.Principal(r)
	if uid == "" {
		response.LoginRequired(w)
		return
	}
	o, err := h.svc.Complete(r.Context(), id, uid)
	if h.writeError(w, r, "complete", err) {
		return
	}
	response.JSON(w, http.StatusOK, o)
}

// AdminSummary melayani GET /admin/summary (dashboard admin).
func (h *Handler) AdminSummary(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Summary(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "order summary failed", "error", err)
		response.Internal(w)
		return
	}
	response.JSON(w, http.StatusOK, s)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case isOngkirError(err):
		ongkir.WriteError(w, r, op+" order", err)
	case errors.Is(err, ErrCartChanged):
		response.Conflict(w, "Isi keranjang berubah. Silakan cek ongkir ulang lalu checkout lagi.")
	case errors.Is(err, ErrCartEmpty):
		response.ValidationFailed(w, response.FieldError{Field: "cart_item_ids", Message: "Keranjang masih kosong."})
	case errors.Is(err, ErrCartItemNotFound):
		response.ValidationFailed(w, response.FieldError{Field: "cart_item_ids",
			Message: "Ada item keranjang yang tidak ditemukan."})
	case errors.Is(err, ErrProductUnavailable):
		response.Conflict(w, "Ada produk di keranjang yang sudah tidak tersedia.")
	case errors.Is(err, ErrInsufficientStock):
		response.Conflict(w, "Stok tidak mencukupi.")
	case errors.Is(err, ErrTrackingRequired):
		response.ValidationFailed(w, response.FieldError{Field: "tracking_number", Message: "No. resi wajib diisi."})
	case errors.Is(err, ErrInvalidTransition):
		response.Conflict(w, "Perubahan status order tidak diizinkan.")
	case errors.Is(err, ErrEmailNotVerified):
		response.Error(w, http.StatusForbidden, "email_not_verified",
			"Verifikasi email kamu terlebih dahulu sebelum membuat pesanan.")
	case errors.Is(err, ErrNotCustomer):
		response.Error(w, http.StatusForbidden, "forbidden", "Hanya akun pelanggan yang dapat membuat pesanan.")
	case errors.Is(err, ErrProofUnderReview):
		response.Conflict(w, "Bukti pembayaran sedang diverifikasi admin, pesanan tidak dapat dibatalkan.")
	default:
		slog.ErrorContext(r.Context(), op+" order failed", "error", err)
		response.Internal(w)
	}
	return true
}

// input memvalidasi dan merapikan body checkout.
func (req createRequest) input() (CheckoutInput, []response.FieldError) {
	in := CheckoutInput{
		RecipientName:   strings.TrimSpace(req.RecipientName),
		RecipientPhone:  strings.TrimSpace(req.RecipientPhone),
		ShippingAddress: strings.TrimSpace(req.ShippingAddress),
		Destination: ongkir.Destination{
			ProvinceID:    req.ProvinceID,
			CityID:        req.CityID,
			DistrictID:    req.DistrictID,
			SubdistrictID: req.SubdistrictID,
		},
	}

	var c validator.Collector
	c.Check(validator.RuneLenBetween(in.RecipientName, 1, 100), "recipient_name",
		"Nama penerima wajib diisi, maksimal 100 karakter.")
	c.Check(len(in.RecipientPhone) >= 6 && validator.Phone(in.RecipientPhone), "recipient_phone",
		"Nomor telepon tidak valid (6 sampai 20 karakter).")
	c.Check(validator.RuneLenBetween(in.ShippingAddress, 1, 500), "shipping_address",
		"Alamat wajib diisi, maksimal 500 karakter.")
	c.Check(req.ProvinceID > 0, "province_id", "Provinsi wajib dipilih.")
	c.Check(req.CityID > 0, "city_id", "Kota/kabupaten wajib dipilih.")
	c.Check(req.DistrictID > 0, "district_id", "Kecamatan wajib dipilih.")
	c.Check(req.SubdistrictID > 0, "subdistrict_id", "Kelurahan wajib dipilih.")

	// cart_item_ids opsional: kosong berarti seluruh isi keranjang.
	seen := map[string]bool{}
	valid := len(req.CartItemIDs) <= maxCartItemIDs
	for _, id := range req.CartItemIDs {
		if !validator.UUID(id) || seen[strings.ToLower(id)] {
			valid = false
			break
		}
		seen[strings.ToLower(id)] = true
	}
	c.Check(valid, "cart_item_ids", "cart_item_ids harus berisi UUID unik (maksimal 100).")
	in.CartItemIDs = req.CartItemIDs

	return in, c.Errors()
}

// parseFilter membaca query order_status yang dipakai kedua endpoint daftar.
func parseFilter(w http.ResponseWriter, r *http.Request) (Filter, bool) {
	f := Filter{Status: r.URL.Query().Get("order_status")}
	if f.Status != "" && !validStatus(f.Status) {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter order_status tidak valid.")
		return Filter{}, false
	}
	return f, true
}

// isOngkirError melaporkan apakah err berasal dari perhitungan ongkir.
func isOngkirError(err error) bool {
	return errors.Is(err, ongkir.ErrDisabled) || errors.Is(err, ongkir.ErrInvalidDestination) ||
		errors.Is(err, ongkir.ErrNoRoute) ||
		strings.HasPrefix(err.Error(), "rajaongkir:")
}
