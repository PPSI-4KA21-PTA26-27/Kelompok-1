package product

// File ini berisi fitur gambar produk: model, akses data, logika bisnis, dan
// handler endpoint /products/{productId}/images. Gambar disimpan sebagai URL
// (misalnya hasil POST /admin/uploads/images ke Cloudinary).

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

const (
	imageColumns = `id, product_id, image_url, sort_order, is_primary`

	imageURLMessage = "image_url harus URL http atau https yang valid."
)

// ErrImageNotFound berarti gambar tidak ada pada produk tersebut.
var ErrImageNotFound = errors.New("product: image not found")

// ---------------------------------------------------------------- model

// Image adalah satu baris product_images (schema Image).
type Image struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	ImageURL  string `json:"image_url"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

// ImageChanges adalah perubahan gambar yang sudah divalidasi; nil berarti
// tidak diubah.
type ImageChanges struct {
	ImageURL  *string
	SortOrder *int
	IsPrimary *bool
}

type imageCreateRequest struct {
	ImageURL  string `json:"image_url"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

type imageUpdateRequest struct {
	ImageURL  *string `json:"image_url"`
	SortOrder *int    `json:"sort_order"`
	IsPrimary *bool   `json:"is_primary"`
}

// ----------------------------------------------------------- repository

// CreateImage menyimpan gambar baru. Bila IsPrimary true, gambar utama lama
// dijadikan non-utama dalam transaksi yang sama, sehingga tiap produk hanya
// punya satu gambar utama. ErrNotFound bila produk tidak ada.
func (r *Repository) CreateImage(ctx context.Context, img Image) (Image, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Image{}, fmt.Errorf("product: create image begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if img.IsPrimary {
		if err := unsetPrimary(ctx, tx, img.ProductID); err != nil {
			return Image{}, err
		}
	}

	created, err := scanImage(tx.QueryRow(ctx,
		`INSERT INTO product_images (product_id, image_url, sort_order, is_primary)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+imageColumns,
		img.ProductID, img.ImageURL, img.SortOrder, img.IsPrimary))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation {
			return Image{}, ErrNotFound
		}
		return Image{}, fmt.Errorf("product: create image: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Image{}, fmt.Errorf("product: create image commit: %w", err)
	}
	return created, nil
}

// UpdateImage mengubah field yang tidak nil. Menjadikan gambar ini utama akan
// melepas status utama gambar lain milik produk yang sama.
func (r *Repository) UpdateImage(ctx context.Context, productID, id string, ch ImageChanges) (Image, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Image{}, fmt.Errorf("product: update image begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Pastikan gambar memang milik produk ini sebelum menyentuh gambar lain.
	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM product_images WHERE id = $1 AND product_id = $2)`,
		id, productID).Scan(&exists)
	if err != nil {
		return Image{}, fmt.Errorf("product: update image check: %w", err)
	}
	if !exists {
		return Image{}, ErrImageNotFound
	}

	if ch.IsPrimary != nil && *ch.IsPrimary {
		if err := unsetPrimary(ctx, tx, productID); err != nil {
			return Image{}, err
		}
	}

	args := []any{id, productID}
	var sets []string
	set := func(col string, v any) {
		args = append(args, v)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	if ch.ImageURL != nil {
		set("image_url", *ch.ImageURL)
	}
	if ch.SortOrder != nil {
		set("sort_order", *ch.SortOrder)
	}
	if ch.IsPrimary != nil {
		set("is_primary", *ch.IsPrimary)
	}

	query := "SELECT " + imageColumns + " FROM product_images WHERE id = $1 AND product_id = $2"
	if len(sets) > 0 {
		query = "UPDATE product_images SET " + strings.Join(sets, ", ") +
			" WHERE id = $1 AND product_id = $2 RETURNING " + imageColumns
	}
	img, err := scanImage(tx.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, ErrImageNotFound) {
			return Image{}, err
		}
		return Image{}, fmt.Errorf("product: update image: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Image{}, fmt.Errorf("product: update image commit: %w", err)
	}
	return img, nil
}

// DeleteImage menghapus gambar milik produk.
func (r *Repository) DeleteImage(ctx context.Context, productID, id string) error {
	tag, err := r.db.Exec(ctx,
		"DELETE FROM product_images WHERE id = $1 AND product_id = $2", id, productID)
	if err != nil {
		return fmt.Errorf("product: delete image: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrImageNotFound
	}
	return nil
}

// unsetPrimary melepas status utama semua gambar milik produk.
func unsetPrimary(ctx context.Context, tx pgx.Tx, productID string) error {
	_, err := tx.Exec(ctx,
		"UPDATE product_images SET is_primary = FALSE WHERE product_id = $1 AND is_primary", productID)
	if err != nil {
		return fmt.Errorf("product: unset primary image: %w", err)
	}
	return nil
}

func scanImage(row pgx.Row) (Image, error) {
	var img Image
	err := row.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.SortOrder, &img.IsPrimary)
	if errors.Is(err, pgx.ErrNoRows) {
		return Image{}, ErrImageNotFound
	}
	return img, err
}

// -------------------------------------------------------------- service

// CreateImage menambah gambar ke produk.
func (s *Service) CreateImage(ctx context.Context, img Image) (Image, error) {
	return s.repo.CreateImage(ctx, img)
}

// UpdateImage mengubah gambar milik produk.
func (s *Service) UpdateImage(ctx context.Context, productID, id string, ch ImageChanges) (Image, error) {
	return s.repo.UpdateImage(ctx, productID, id, ch)
}

// DeleteImage menghapus gambar milik produk.
func (s *Service) DeleteImage(ctx context.Context, productID, id string) error {
	return s.repo.DeleteImage(ctx, productID, id)
}

// -------------------------------------------------------------- handler

// CreateImage melayani POST /products/{productId}/images (admin).
func (h *Handler) CreateImage(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productId")
	if !validator.UUID(productID) {
		response.NotFound(w)
		return
	}

	var req imageCreateRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.ImageURL = strings.TrimSpace(req.ImageURL)
	if !validator.HTTPURL(req.ImageURL) {
		response.ValidationFailed(w, response.FieldError{Field: "image_url", Message: imageURLMessage})
		return
	}

	img, err := h.svc.CreateImage(r.Context(), Image{
		ProductID: productID,
		ImageURL:  req.ImageURL,
		SortOrder: req.SortOrder,
		IsPrimary: req.IsPrimary,
	})
	if h.writeImageError(w, r, "create", err) {
		return
	}
	response.JSON(w, http.StatusCreated, img)
}

// UpdateImage melayani PATCH /products/{productId}/images/{imageId} (admin).
func (h *Handler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	productID, id := r.PathValue("productId"), r.PathValue("imageId")
	if !validator.UUID(productID) || !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	var req imageUpdateRequest
	if !response.Decode(w, r, &req) {
		return
	}

	var c validator.Collector
	ch := ImageChanges{SortOrder: req.SortOrder, IsPrimary: req.IsPrimary}
	if req.ImageURL != nil {
		u := strings.TrimSpace(*req.ImageURL)
		c.Check(validator.HTTPURL(u), "image_url", imageURLMessage)
		ch.ImageURL = &u
	}
	c.Check(ch != (ImageChanges{}), "body", "Isi minimal satu field yang ingin diubah.")
	if errs := c.Errors(); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	img, err := h.svc.UpdateImage(r.Context(), productID, id, ch)
	if h.writeImageError(w, r, "update", err) {
		return
	}
	response.JSON(w, http.StatusOK, img)
}

// DeleteImage melayani DELETE /products/{productId}/images/{imageId} (admin).
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	productID, id := r.PathValue("productId"), r.PathValue("imageId")
	if !validator.UUID(productID) || !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	err := h.svc.DeleteImage(r.Context(), productID, id)
	if h.writeImageError(w, r, "delete", err) {
		return
	}
	response.NoContent(w)
}

// writeImageError menjawab error fitur gambar. Mengembalikan true bila err
// tidak nil (response sudah ditulis).
func (h *Handler) writeImageError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrImageNotFound):
		response.NotFound(w)
	default:
		slog.ErrorContext(r.Context(), op+" image failed", "error", err)
		response.Internal(w)
	}
	return true
}
