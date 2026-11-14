package post

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

const (
	maxContent  = 60000
	maxProducts = 12
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// errInvalid membawa error validasi dari tahap merge PATCH.
type errInvalid struct{ fields []response.FieldError }

func (e errInvalid) Error() string { return "post: invalid input" }

// Handler melayani /posts (publik) dan /admin/posts (admin & staff).
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// PublicList melayani GET /posts?tag=&page=&limit=.
func (h *Handler) PublicList(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, ListFilter{Public: true, Tag: strings.TrimSpace(r.URL.Query().Get("tag"))})
}

// PublicGet melayani GET /posts/{slug}.
func (h *Handler) PublicGet(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !slugPattern.MatchString(slug) || len(slug) > 160 {
		notFound(w)
		return
	}
	d, err := h.svc.GetPublic(r.Context(), slug)
	if h.writeError(w, r, "get", err) {
		return
	}
	response.JSON(w, http.StatusOK, d)
}

// AdminList melayani GET /admin/posts?status=&q=.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := ListFilter{Status: q.Get("status"), Q: strings.TrimSpace(q.Get("q"))}
	if f.Status != "" && !validStatus(f.Status) {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter status tidak valid.")
		return
	}
	if utf8.RuneCountInString(f.Q) > 100 {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter q maksimal 100 karakter.")
		return
	}
	h.list(w, r, f)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, f ListFilter) {
	page, limit, ok := response.ParsePage(w, r)
	if !ok {
		return
	}
	posts, total, err := h.svc.List(r.Context(), f, limit, (page-1)*limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "list posts failed", "error", err)
		response.Internal(w)
		return
	}
	response.List(w, http.StatusOK, posts, response.Meta{Page: page, Limit: limit, Total: total})
}

// AdminGet melayani GET /admin/posts/{postId}.
func (h *Handler) AdminGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("postId")
	if !validator.UUID(id) {
		notFound(w)
		return
	}
	d, err := h.svc.Get(r.Context(), id)
	if h.writeError(w, r, "get", err) {
		return
	}
	response.JSON(w, http.StatusOK, d)
}

// Create melayani POST /admin/posts.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "Autentikasi diperlukan.")
		return
	}
	var req request
	if !response.Decode(w, r, &req) {
		return
	}
	in := Input{Status: StatusDraft}
	req.apply(&in)
	if in.Slug == "" {
		in.Slug = Slugify(in.Title)
	}
	if errs := validate(in); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}
	d, err := h.svc.Create(r.Context(), p.UserID, in)
	if h.writeError(w, r, "create", err) {
		return
	}
	response.JSON(w, http.StatusCreated, d)
}

// Update melayani PATCH /admin/posts/{postId}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("postId")
	if !validator.UUID(id) {
		notFound(w)
		return
	}
	var req request
	if !response.Decode(w, r, &req) {
		return
	}
	d, err := h.svc.Update(r.Context(), id, req.apply, func(in Input) error {
		if errs := validate(in); len(errs) > 0 {
			return errInvalid{errs}
		}
		return nil
	})
	var inv errInvalid
	if errors.As(err, &inv) {
		response.ValidationFailed(w, inv.fields...)
		return
	}
	if h.writeError(w, r, "update", err) {
		return
	}
	response.JSON(w, http.StatusOK, d)
}

// Delete melayani DELETE /admin/posts/{postId}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("postId")
	if !validator.UUID(id) {
		notFound(w)
		return
	}
	if h.writeError(w, r, "delete", h.svc.Delete(r.Context(), id)) {
		return
	}
	response.NoContent(w)
}

// apply menyalin field yang dikirim ke in (string kosong menghapus field opsional).
func (req request) apply(in *Input) {
	opt := func(p *string) *string {
		v := strings.TrimSpace(*p)
		if v == "" {
			return nil
		}
		return &v
	}
	if req.Title != nil {
		in.Title = strings.TrimSpace(*req.Title)
	}
	if req.Slug != nil {
		in.Slug = strings.ToLower(strings.TrimSpace(*req.Slug))
		if in.Slug == "" {
			in.Slug = Slugify(in.Title)
		}
	}
	if req.Excerpt != nil {
		in.Excerpt = opt(req.Excerpt)
	}
	if req.Content != nil {
		in.Content = strings.TrimSpace(*req.Content)
	}
	if req.CoverImageURL != nil {
		in.CoverImageURL = opt(req.CoverImageURL)
	}
	if req.Tag != nil {
		in.Tag = opt(req.Tag)
	}
	if req.Status != nil {
		in.Status = strings.TrimSpace(*req.Status)
	}
	if req.PublishedAt.Set {
		in.PublishedAt = req.PublishedAt.Value
	}
	if req.ProductIDs != nil {
		in.ProductIDs = dedupe(*req.ProductIDs)
	}
}

func validate(in Input) []response.FieldError {
	var c validator.Collector
	c.Check(validator.RuneLenBetween(in.Title, 3, 160), "title", "Judul wajib diisi, 3 sampai 160 karakter.")
	c.Check(len(in.Slug) <= 160 && slugPattern.MatchString(in.Slug), "slug",
		"Slug hanya huruf kecil, angka, dan tanda hubung (mis. cara-merawat-hoodie).")
	c.Check(in.Excerpt == nil || validator.RuneLenBetween(*in.Excerpt, 1, 300), "excerpt", "Ringkasan maksimal 300 karakter.")
	c.Check(utf8.RuneCountInString(in.Content) <= maxContent, "content", "Isi artikel terlalu panjang.")
	c.Check(in.Status != StatusPublished || utf8.RuneCountInString(in.Content) >= 50, "content",
		"Isi artikel minimal 50 karakter sebelum diterbitkan.")
	c.Check(in.CoverImageURL == nil || validator.HTTPURL(*in.CoverImageURL), "cover_image_url", "Cover harus URL http/https.")
	c.Check(in.Tag == nil || validator.RuneLenBetween(*in.Tag, 1, 40), "tag", "Tag maksimal 40 karakter.")
	c.Check(validStatus(in.Status), "status", "Status harus draft, published, atau archived.")
	valid := len(in.ProductIDs) <= maxProducts
	for _, id := range in.ProductIDs {
		valid = valid && validator.UUID(id)
	}
	c.Check(valid, "product_ids", "Produk terkait maksimal 12 dan harus id produk yang valid.")
	return c.Errors()
}

// accents mengganti huruf Latin beraksen yang umum dengan huruf dasarnya.
var accents = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ä", "a", "ã", "a", "å", "a",
	"è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i",
	"ò", "o", "ó", "o", "ô", "o", "ö", "o", "õ", "o",
	"ù", "u", "ú", "u", "û", "u", "ü", "u",
	"ñ", "n", "ç", "c", "&", " dan ",
)

// Slugify membuat slug dari judul: huruf kecil ASCII, angka, tanda hubung.
func Slugify(title string) string {
	var b strings.Builder
	dash := false
	for _, r := range accents.Replace(strings.ToLower(title)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	s := strings.TrimRight(b.String(), "-")
	if len(s) > 120 {
		s = strings.TrimRight(s[:120], "-")
	}
	return s
}

func dedupe(ids []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.ToLower(strings.TrimSpace(id))
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func validStatus(s string) bool {
	return s == StatusDraft || s == StatusPublished || s == StatusArchived
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound):
		notFound(w)
	case errors.Is(err, ErrSlugTaken):
		response.Conflict(w, "Slug sudah dipakai artikel lain. Ganti slug.")
	case errors.Is(err, ErrProductNotFound):
		response.ValidationFailed(w, response.FieldError{Field: "product_ids", Message: "Ada produk terkait yang tidak ditemukan."})
	default:
		slog.ErrorContext(r.Context(), op+" post failed", "error", err)
		response.Internal(w)
	}
	return true
}

func notFound(w http.ResponseWriter) {
	response.Error(w, http.StatusNotFound, "not_found", "Artikel tidak ditemukan.")
}
