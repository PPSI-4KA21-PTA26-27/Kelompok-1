package banner

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// errSchedule: jadwal selesai tidak setelah jadwal mulai.
var errSchedule = errors.New("banner: ends_at must be after starts_at")

// Handler melayani /banners.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List melayani GET /banners?placement=hero|main|catalog. Publik hanya
// melihat banner yang sedang tayang; admin dengan include_inactive=true
// melihat semua (termasuk terjadwal & berakhir).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := ListFilter{Placement: strings.TrimSpace(q.Get("placement"))}
	if f.Placement != "" && !validPlacement(f.Placement) {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter placement harus hero, main, atau catalog.")
		return
	}
	if v := q.Get("include_inactive"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "bad_request", "Parameter include_inactive harus true atau false.")
			return
		}
		f.All = b
	}
	if f.All && !user.IsAdminRequest(r) {
		response.Forbidden(w)
		return
	}
	banners, err := h.svc.List(r.Context(), f)
	if err != nil {
		slog.ErrorContext(r.Context(), "list banners failed", "error", err)
		response.Internal(w)
		return
	}
	response.JSON(w, http.StatusOK, banners)
}

// Create melayani POST /banners (admin).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req request
	if !response.Decode(w, r, &req) {
		return
	}
	ch, errs := req.changes(true)
	if len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}
	b := Banner{Placement: PlacementHero, IsActive: true}
	ch.apply(&b)
	if err := checkSchedule(b); err != nil {
		response.ValidationFailed(w, scheduleError())
		return
	}
	created, err := h.svc.Create(r.Context(), b)
	if h.writeError(w, r, "create", err) {
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

// Update melayani PATCH /banners/{bannerId} (admin). Field yang tidak
// dikirim tidak berubah; null menghapus nilai opsional (mis. jadwal).
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("bannerId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}
	var req request
	if !response.Decode(w, r, &req) {
		return
	}
	ch, errs := req.changes(false)
	if len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}
	b, err := h.svc.Update(r.Context(), id, ch, checkSchedule)
	if errors.Is(err, errSchedule) {
		response.ValidationFailed(w, scheduleError())
		return
	}
	if h.writeError(w, r, "update", err) {
		return
	}
	response.JSON(w, http.StatusOK, b)
}

// Delete melayani DELETE /banners/{bannerId} (admin).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("bannerId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}
	err := h.svc.Delete(r.Context(), id)
	if h.writeError(w, r, "delete", err) {
		return
	}
	response.NoContent(w)
}

// changes memvalidasi & menormalkan body. create mewajibkan image_url.
func (req request) changes(create bool) (Changes, []response.FieldError) {
	var c validator.Collector
	ch := Changes{SortOrder: req.SortOrder, IsActive: req.IsActive, StartsAt: req.StartsAt, EndsAt: req.EndsAt}

	if req.Placement != nil {
		p := strings.TrimSpace(*req.Placement)
		c.Check(validPlacement(p), "placement", "Penempatan harus hero, main, atau catalog.")
		ch.Placement = &p
	}
	if req.ImageURL != nil {
		u := strings.TrimSpace(*req.ImageURL)
		c.Check(validator.HTTPURL(u), "image_url", "Gambar banner wajib berupa URL http/https.")
		ch.ImageURL = &u
	} else if create {
		c.Check(false, "image_url", "Gambar banner wajib diunggah.")
	}
	text := func(o optText, field string, max int, msg string) optText {
		if !o.Set || o.Value == nil {
			return o
		}
		v := strings.TrimSpace(*o.Value)
		if v == "" {
			return optText{Set: true}
		}
		c.Check(validator.RuneLenBetween(v, 1, max), field, msg)
		return optText{Set: true, Value: &v}
	}
	ch.Title = text(req.Title, "title", 120, "Judul maksimal 120 karakter.")
	ch.Subtitle = text(req.Subtitle, "subtitle", 240, "Subjudul maksimal 240 karakter.")
	ch.CTALabel = text(req.CTALabel, "cta_label", 40, "Label tombol maksimal 40 karakter.")
	ch.MobileImageURL = text(req.MobileImageURL, "mobile_image_url", 1000, "URL gambar mobile terlalu panjang.")
	if ch.MobileImageURL.Value != nil {
		c.Check(validator.HTTPURL(*ch.MobileImageURL.Value), "mobile_image_url", "Gambar mobile harus URL http/https.")
	}
	ch.LinkURL = text(req.LinkURL, "link_url", 1000, "Tautan terlalu panjang.")
	if ch.LinkURL.Value != nil {
		c.Check(validLink(*ch.LinkURL.Value), "link_url",
			"Tautan harus alamat dalam situs (mis. /katalog) atau URL http/https.")
	}
	if ch.SortOrder != nil {
		c.Check(*ch.SortOrder >= 0 && *ch.SortOrder <= 999, "sort_order", "Urutan 0 sampai 999.")
	}
	c.Check(create || ch.hasAny(), "body", "Isi minimal satu field yang ingin diubah.")
	return ch, c.Errors()
}

func (ch Changes) hasAny() bool {
	return ch.Placement != nil || ch.ImageURL != nil || ch.SortOrder != nil || ch.IsActive != nil ||
		ch.Title.Set || ch.Subtitle.Set || ch.CTALabel.Set || ch.MobileImageURL.Set || ch.LinkURL.Set ||
		ch.StartsAt.Set || ch.EndsAt.Set
}

func checkSchedule(b Banner) error {
	if b.StartsAt != nil && b.EndsAt != nil && !b.EndsAt.After(*b.StartsAt) {
		return errSchedule
	}
	return nil
}

func scheduleError() response.FieldError {
	return response.FieldError{Field: "ends_at", Message: "Waktu turun tayang harus setelah waktu mulai tayang."}
}

func validPlacement(p string) bool {
	return p == PlacementHero || p == PlacementMain || p == PlacementCatalog
}

// validLink menerima path internal ("/shop", bukan "//host") atau URL http(s).
func validLink(s string) bool {
	if strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "//") && !strings.ContainsAny(s, " \\") {
		return true
	}
	return validator.HTTPURL(s)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	default:
		slog.ErrorContext(r.Context(), op+" banner failed", "error", err)
		response.Internal(w)
	}
	return true
}
