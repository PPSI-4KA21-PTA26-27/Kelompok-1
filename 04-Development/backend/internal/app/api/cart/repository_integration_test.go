//go:build integration

// Tes integrasi keranjang terhadap PostgreSQL yang sudah dimigrasi.
// Jalankan: TEST_DATABASE_URL=postgres://... go test -tags integration ./internal/app/api/cart/
// Tanpa TEST_DATABASE_URL tes dilewati. Data uji dibuat dengan email/SKU unik
// dan dihapus di akhir tes.
package cart

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type fixture struct {
	db        *pgxpool.Pool
	svc       *Service
	userID    string
	otherID   string
	variantID string
}

func setup(t *testing.T, stock int) fixture {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tidak diset; tes integrasi dilewati")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	f := fixture{db: db, svc: NewService(NewRepository(db))}
	newUser := func(tag string) string {
		var id string
		err := db.QueryRow(ctx,
			`INSERT INTO users (name, email, password_hash, role_id, email_verified_at)
			 SELECT 'Uji Cart', $1, 'x', id, now() FROM roles WHERE code = 'customer'
			 RETURNING id`, "cart-"+tag+"-"+suffix+"@test.local").Scan(&id)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
		return id
	}
	f.userID, f.otherID = newUser("a"), newUser("b")

	var productID string
	if err := db.QueryRow(ctx,
		`INSERT INTO products (name, price, weight_gram, availability_type, status)
		 VALUES ('Produk Uji Cart', 100000, 200, 'ready', 'active') RETURNING id`).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if err := db.QueryRow(ctx,
		`INSERT INTO product_variants (product_id, sku, size, stock) VALUES ($1, $2, 'M', $3) RETURNING id`,
		productID, "CART-"+suffix, stock).Scan(&f.variantID); err != nil {
		t.Fatalf("insert variant: %v", err)
	}

	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = db.Exec(ctx, `DELETE FROM cart_items WHERE user_id IN ($1, $2)`, f.userID, f.otherID)
		_, _ = db.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1, $2)`, f.userID, f.otherID)
	})
	return f
}

func TestCartFlowIntegration(t *testing.T) {
	f := setup(t, 5)
	ctx := context.Background()

	item, err := f.svc.Add(ctx, f.userID, CreateRequest{VariantID: f.variantID, Quantity: 2})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if item.Quantity != 2 {
		t.Fatalf("quantity = %d, want 2", item.Quantity)
	}

	// Varian yang sama digabung, bukan baris baru.
	merged, err := f.svc.Add(ctx, f.userID, CreateRequest{VariantID: f.variantID, Quantity: 3})
	if err != nil {
		t.Fatalf("Add() gabung error = %v", err)
	}
	if merged.ID != item.ID || merged.Quantity != 5 {
		t.Fatalf("gabung = %s/%d, want %s/5", merged.ID, merged.Quantity, item.ID)
	}

	if _, err := f.svc.Add(ctx, f.userID, CreateRequest{VariantID: f.variantID, Quantity: 1}); !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("melebihi stok: err = %v, want ErrInsufficientStock", err)
	}
	if _, err := f.svc.UpdateQuantity(ctx, f.userID, item.ID, 6); !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("update melebihi stok: err = %v, want ErrInsufficientStock", err)
	}
	if got, err := f.svc.UpdateQuantity(ctx, f.userID, item.ID, 1); err != nil || got.Quantity != 1 {
		t.Errorf("UpdateQuantity(1) = %d, %v", got.Quantity, err)
	}

	// Item milik user lain tidak bisa diubah atau dihapus.
	if _, err := f.svc.UpdateQuantity(ctx, f.otherID, item.ID, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("update milik user lain: err = %v, want ErrNotFound", err)
	}
	if err := f.svc.Delete(ctx, f.otherID, item.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("hapus milik user lain: err = %v, want ErrNotFound", err)
	}

	if err := f.svc.Delete(ctx, f.userID, item.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := f.svc.Delete(ctx, f.userID, item.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("hapus dua kali: err = %v, want ErrNotFound", err)
	}

	if _, err := f.svc.Add(ctx, f.userID, CreateRequest{VariantID: f.variantID, Quantity: 1}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := f.svc.Clear(ctx, f.userID); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	c, err := f.svc.Get(ctx, f.userID)
	if err != nil || len(c.Items) != 0 {
		t.Errorf("setelah Clear: %d item, err %v", len(c.Items), err)
	}
}

func TestCartUnknownVariantIntegration(t *testing.T) {
	f := setup(t, 1)
	_, err := f.svc.Add(context.Background(), f.userID,
		CreateRequest{VariantID: "33333333-3333-4333-8333-333333333333", Quantity: 1})
	if !errors.Is(err, ErrVariantNotFound) {
		t.Errorf("err = %v, want ErrVariantNotFound", err)
	}
}
