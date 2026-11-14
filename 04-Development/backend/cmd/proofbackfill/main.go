// Command proofbackfill memindahkan bukti pembayaran lama yang masih
// tersimpan di payments.proof_data ke Cloudinary dan mengisi proof_url (SDD
// D-07). Aman dijalankan berulang: hanya baris dengan proof_url kosong yang
// diproses. Setelah selesai, jalankan migrasi berikutnya yang menghapus kolom
// proof_data dan proof_type.
//
//	go run ./cmd/proofbackfill
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"hii-stranger1/backend/internal/pkg/cloudinary"
	"hii-stranger1/backend/internal/pkg/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "proofbackfill:", err)
		os.Exit(1)
	}
}

type legacyProof struct {
	id   string
	data []byte
	typ  string
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cdn := cloudinary.New(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret, cfg.CloudinaryFolder)
	if !cdn.Enabled() {
		return errors.New("kredensial Cloudinary belum diisi")
	}

	conn, err := pgx.Connect(ctx, cfg.DBURL)
	if err != nil {
		return fmt.Errorf("koneksi database: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	var hasColumn bool
	if err := conn.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns
		                 WHERE table_name = 'payments' AND column_name = 'proof_data')`).Scan(&hasColumn); err != nil {
		return fmt.Errorf("cek kolom proof_data: %w", err)
	}
	if !hasColumn {
		fmt.Println("kolom proof_data sudah tidak ada; tidak ada yang perlu dipindahkan")
		return nil
	}

	rows, err := conn.Query(ctx,
		`SELECT id, proof_data, COALESCE(proof_type, '') FROM payments
		  WHERE proof_data IS NOT NULL AND proof_url IS NULL
		  ORDER BY id`)
	if err != nil {
		return fmt.Errorf("baca bukti lama: %w", err)
	}
	proofs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (legacyProof, error) {
		var p legacyProof
		err := row.Scan(&p.id, &p.data, &p.typ)
		return p, err
	})
	if err != nil {
		return fmt.Errorf("baca bukti lama: %w", err)
	}

	moved := 0
	for _, p := range proofs {
		url, err := cdn.UploadProof(ctx, p.typ, p.data)
		if err != nil {
			return fmt.Errorf("unggah bukti payment %s (%d sudah dipindahkan): %w", p.id, moved, err)
		}
		if _, err := conn.Exec(ctx,
			`UPDATE payments SET proof_url = $2 WHERE id = $1 AND proof_url IS NULL`, p.id, url); err != nil {
			return fmt.Errorf("simpan proof_url payment %s (%d sudah dipindahkan): %w", p.id, moved, err)
		}
		moved++
	}
	fmt.Printf("%d bukti pembayaran dipindahkan ke Cloudinary\n", moved)
	return nil
}
