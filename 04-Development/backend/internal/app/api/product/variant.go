package product

// File ini berisi fitur varian produk (ukuran/warna/stok):
// model, akses data, logika bisnis, dan handler endpoint
// /products/{productId}/variants.

import (
	"context"
	"errors"
	"fmt"
	"hii-stranger1/backend/internal/pkg/database"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

const (
	uniqueViolation = "23505"

	variantColumns = `id, product_id, sku, size, color, stock`

	skuMessage   = "SKU wajib diisi, maksimal 64 karakter."
	sizeMessage  = "Ukuran wajib diisi, maksimal 32 karakter."
	colorMessage = "Warna maksimal 32 karakter."
	stockMessage = "Stok tidak boleh negatif."
)

var (
	// ErrVariantNotFound berarti varian tidak ada pada produk tersebut.
	ErrVariantNotFound = errors.New("product: variant not found")
	// ErrSKUTaken berarti SKU sudah dipakai varian lain.
	ErrSKUTaken = errors.New("product: sku already used")
	// ErrVariantInUse berarti varian sudah pernah dipesan sehingga tidak bisa dihapus.
	ErrVariantInUse = errors.New("product: variant still referenced by orders")
)

// ---------------------------------------------------------------- model

// Variant adalah satu baris product_variants (schema Variant).
type Variant struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku"`
	Size      string  `json:"size"`
	Color     *string `json:"color"`
	Stock     int     `json:"stock"`
}

// VariantChanges adalah perubahan varian yang sudah divalidasi; nil berarti
// tidak diubah. Color berisi string kosong berarti warna dikosongkan (NULL).
type VariantChanges struct {
	SKU   *string
	Size  *string
	Color *string
	Stock *int
}

type variantCreateRequest struct {
	SKU   string  `json:"sku"`
	Size  string  `json:"size"`
	Color *string `json:"color"`
	Stock *int    `json:"stock"`
}

type variantUpdateRequest struct {
	SKU   *string `json:"sku"`
	Size  *string `json:"size"`
	Color *string `json:"color"`
	Stock *int    `json:"stock"`
}

// ----------------------------------------------------------- repository

// CreateVariant menyimpan varian baru. ErrNotFound bila produk tidak ada,
// ErrSKUTaken bila SKU sudah dipakai.
func (r *Repository) CreateVariant(ctx context.Context, v Variant) (Variant, error) {
	created, err := scanVariant(r.db.QueryRow(ctx,
		`INSERT INTO product_variants (product_id, sku, size, color, stock)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+variantColumns,
		v.ProductID, v.SKU, v.Size, v.Color, v.Stock))
	if err != nil {
		return Variant{}, wrapVariant("create variant", err)
	}
	return created, nil
}

// UpdateVariant mengubah field yang tidak nil. ErrVariantNotFound bila varian
// tidak ada pada produk tersebut.
func (r *Repository) UpdateVariant(ctx context.Context, productID, id string, ch VariantChanges) (Variant, error) {
	args := []any{id, productID}
	var sets []string
	set := func(col string, v any) {
		args = append(args, v)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	if ch.SKU != nil {
		set("sku", *ch.SKU)
	}
	if ch.Size != nil {
		set("size", *ch.Size)
	}
	if ch.Color != nil {
		set("color", database.NullIfEmpty(*ch.Color))
	}
	if ch.Stock != nil {
		set("stock", *ch.Stock)
	}

	query := "SELECT " + variantColumns + " FROM product_variants WHERE id = $1 AND product_id = $2"
	if len(sets) > 0 {
		query = "UPDATE product_variants SET " + strings.Join(sets, ", ") +
			" WHERE id = $1 AND product_id = $2 RETURNING " + variantColumns
	}
	v, err := scanVariant(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		return Variant{}, wrapVariant("update variant", err)
	}
	return v, nil
}

// DeleteVariant menghapus varian. ErrVariantInUse bila sudah pernah dipesan.
func (r *Repository) DeleteVariant(ctx context.Context, productID, id string) error {
	tag, err := r.db.Exec(ctx,
		"DELETE FROM product_variants WHERE id = $1 AND product_id = $2", id, productID)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrVariantInUse
		}
		return fmt.Errorf("product: delete variant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrVariantNotFound
	}
	return nil
}

func scanVariant(row pgx.Row) (Variant, error) {
	var v Variant
	err := row.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Size, &v.Color, &v.Stock)
	if errors.Is(err, pgx.ErrNoRows) {
		return Variant{}, ErrVariantNotFound
	}
	return v, err
}

// wrapVariant menerjemahkan error database menjadi error domain varian.
func wrapVariant(op string, err error) error {
	if errors.Is(err, ErrVariantNotFound) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == uniqueViolation && pgErr.ConstraintName == "uq_product_variants_sku":
			return ErrSKUTaken
		case pgErr.Code == foreignKeyViolation && pgErr.ConstraintName == "fk_product_variants_product":
			return ErrNotFound
		}
	}
	return fmt.Errorf("product: %s: %w", op, err)
}

// -------------------------------------------------------------- service

// CreateVariant menambah varian ke produk.
func (s *Service) CreateVariant(ctx context.Context, v Variant) (Variant, error) {
	return s.repo.CreateVariant(ctx, v)
}

// UpdateVariant mengubah varian milik produk.
func (s *Service) UpdateVariant(ctx context.Context, productID, id string, ch VariantChanges) (Variant, error) {
	return s.repo.UpdateVariant(ctx, productID, id, ch)
}

// DeleteVariant menghapus varian milik produk.
func (s *Service) DeleteVariant(ctx context.Context, productID, id string) error {
	return s.repo.DeleteVariant(ctx, productID, id)
}

// -------------------------------------------------------------- handler

// CreateVariant melayani POST /products/{productId}/variants (admin).
func (h *Handler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productId")
	if !validator.UUID(productID) {
		response.NotFound(w)
		return
	}

	var req variantCreateRequest
	if !response.Decode(w, r, &req) {
		return
	}
	v := Variant{
		ProductID: productID,
		SKU:       strings.TrimSpace(req.SKU),
		Size:      strings.TrimSpace(req.Size),
	}
	if req.Color != nil {
		v.Color = database.NullIfEmpty(strings.TrimSpace(*req.Color))
	}
	if req.Stock != nil {
		v.Stock = *req.Stock
	}

	var c validator.Collector
	c.Check(validator.RuneLenBetween(v.SKU, 1, 64), "sku", skuMessage)
	c.Check(validator.RuneLenBetween(v.Size, 1, 32), "size", sizeMessage)
	c.Check(v.Color == nil || validator.RuneLenBetween(*v.Color, 0, 32), "color", colorMessage)
	c.Check(v.Stock >= 0, "stock", stockMessage)
	if errs := c.Errors(); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	created, err := h.svc.CreateVariant(r.Context(), v)
	if h.writeVariantError(w, r, "create", err) {
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

// UpdateVariant melayani PATCH /products/{productId}/variants/{variantId} (admin).
func (h *Handler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	productID, id := r.PathValue("productId"), r.PathValue("variantId")
	if !validator.UUID(productID) || !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	var req variantUpdateRequest
	if !response.Decode(w, r, &req) {
		return
	}

	var c validator.Collector
	ch := VariantChanges{Stock: req.Stock}
	if req.SKU != nil {
		sku := strings.TrimSpace(*req.SKU)
		c.Check(validator.RuneLenBetween(sku, 1, 64), "sku", skuMessage)
		ch.SKU = &sku
	}
	if req.Size != nil {
		size := strings.TrimSpace(*req.Size)
		c.Check(validator.RuneLenBetween(size, 1, 32), "size", sizeMessage)
		ch.Size = &size
	}
	if req.Color != nil {
		color := strings.TrimSpace(*req.Color)
		c.Check(validator.RuneLenBetween(color, 0, 32), "color", colorMessage)
		ch.Color = &color
	}
	c.Check(req.Stock == nil || *req.Stock >= 0, "stock", stockMessage)
	c.Check(ch != (VariantChanges{}), "body", "Isi minimal satu field yang ingin diubah.")
	if errs := c.Errors(); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	v, err := h.svc.UpdateVariant(r.Context(), productID, id, ch)
	if h.writeVariantError(w, r, "update", err) {
		return
	}
	response.JSON(w, http.StatusOK, v)
}

// DeleteVariant melayani DELETE /products/{productId}/variants/{variantId} (admin).
func (h *Handler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	productID, id := r.PathValue("productId"), r.PathValue("variantId")
	if !validator.UUID(productID) || !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	err := h.svc.DeleteVariant(r.Context(), productID, id)
	if h.writeVariantError(w, r, "delete", err) {
		return
	}
	response.NoContent(w)
}

// writeVariantError menjawab error fitur varian. Mengembalikan true bila err
// tidak nil (response sudah ditulis).
func (h *Handler) writeVariantError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrVariantNotFound):
		response.NotFound(w)
	case errors.Is(err, ErrSKUTaken):
		response.Conflict(w, "SKU sudah dipakai varian lain.")
	case errors.Is(err, ErrVariantInUse):
		response.Error(w, http.StatusConflict, "conflict",
			"Varian sudah pernah dipesan sehingga tidak bisa dihapus. Ubah stoknya menjadi 0.")
	default:
		slog.ErrorContext(r.Context(), op+" variant failed", "error", err)
		response.Internal(w)
	}
	return true
}
