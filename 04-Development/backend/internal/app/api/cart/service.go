package cart

import (
	"context"
	"math"
)

// Service berisi logika bisnis cart.
type Service struct {
	repo *Repository
}

// NewService membuat Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Get mengambil isi keranjang milik user.
func (s *Service) Get(ctx context.Context, userID string) (Cart, error) {
	if userID == "" {
		return Cart{}, ErrUserNotFound
	}

	return s.repo.GetCart(ctx, userID)
}

// Add menambahkan varian ke keranjang. Bila varian sudah ada, jumlahnya
// digabung. Total jumlah tidak boleh melebihi stok varian.
func (s *Service) Add(
	ctx context.Context,
	userID string,
	req CreateRequest,
) (ItemView, error) {
	if userID == "" {
		return ItemView{}, ErrUserNotFound
	}

	if req.VariantID == "" {
		return ItemView{}, ErrVariantNotFound
	}

	if req.Quantity < 1 {
		return ItemView{}, ErrInvalidQuantity
	}

	// Stok disimpan sebagai INT, jadi jumlah di atas batas ini pasti melebihi stok.
	if req.Quantity > math.MaxInt32 {
		return ItemView{}, ErrInsufficientStock
	}

	id, err := s.repo.AddItem(ctx, userID, req.VariantID, req.Quantity)
	if err != nil {
		return ItemView{}, err
	}

	return s.repo.GetItem(ctx, userID, id)
}

// UpdateQuantity mengubah jumlah item keranjang milik user.
func (s *Service) UpdateQuantity(
	ctx context.Context,
	userID, id string,
	quantity int,
) (ItemView, error) {
	if userID == "" {
		return ItemView{}, ErrUserNotFound
	}

	if id == "" {
		return ItemView{}, ErrNotFound
	}

	if quantity < 1 {
		return ItemView{}, ErrInvalidQuantity
	}

	if quantity > math.MaxInt32 {
		return ItemView{}, ErrInsufficientStock
	}

	if err := s.repo.SetQuantity(ctx, userID, id, quantity); err != nil {
		return ItemView{}, err
	}

	return s.repo.GetItem(ctx, userID, id)
}

// Delete menghapus satu item dari keranjang.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	if userID == "" {
		return ErrUserNotFound
	}

	if id == "" {
		return ErrNotFound
	}

	return s.repo.DeleteItem(ctx, userID, id)
}

// Clear menghapus seluruh item keranjang milik user.
func (s *Service) Clear(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrUserNotFound
	}

	return s.repo.ClearCart(ctx, userID)
}
