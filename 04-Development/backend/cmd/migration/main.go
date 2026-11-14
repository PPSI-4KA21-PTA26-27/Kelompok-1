// Command migration menjalankan file SQL di folder migrations/ ke database
// memakai golang-migrate (SDD D-10): berkas up/down, satu baris catatan versi
// di tabel schema_migrations (version, dirty).
//
// Cara pakai (dari folder backend):
//
//	go run ./cmd/migration up             jalankan semua migrasi yang belum jalan
//	go run ./cmd/migration down           batalkan 1 migrasi terakhir
//	go run ./cmd/migration down 3         batalkan 3 migrasi terakhir
//	go run ./cmd/migration status         lihat mana yang sudah / belum jalan
//	go run ./cmd/migration force <versi>  tandai versi tanpa menjalankan SQL
//	                                      (untuk memulihkan status dirty setelah gagal)
//	go run ./cmd/migration fresh          HAPUS semua tabel lalu jalankan ulang semua migrasi
//	                                      (hanya boleh saat APP_ENV=dev, data ikut hilang)
//	go run ./cmd/migration create nama    buat pasangan file .up.sql dan .down.sql baru
//
// Opsi: -dir <folder> untuk memakai folder migrasi lain (default: migrations).
//
// Format nama file: <versi>_<nama>.up.sql dan <versi>_<nama>.down.sql.
// Versi berupa angka dan diurutkan sebagai angka, jadi 000001 jalan lebih dulu
// daripada 20261029004406.
//
// Alamat database dibaca dari DB_URL (sama seperti server). Kalau kosong,
// DATABASE_URL dipakai sebagai cadangan. File .env ikut dibaca bila ada.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const defaultDir = "migrations"

// fileName mencocokkan nama file migrasi, contoh: 000001_create_users_table.up.sql
var fileName = regexp.MustCompile(`^(\d+)_([a-zA-Z0-9_]+)\.(up|down)\.sql$`)

// migration adalah satu pasangan file up/down dengan versi yang sama.
type migration struct {
	version  int64
	name     string
	upPath   string
	downPath string
}

// label menghasilkan teks seperti "000001_create_users_table" untuk dicetak.
func (m migration) label() string {
	return filepath.Base(strings.TrimSuffix(m.upPath, ".up.sql"))
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("migration", flag.ContinueOnError)
	dir := flags.String("dir", defaultDir, "folder berisi file migrasi")
	flags.Usage = func() { printUsage(flags) }

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	cmd := flags.Arg(0)
	rest := flags.Args()
	if len(rest) > 0 {
		rest = rest[1:]
	}

	// "create" hanya membuat file, tidak butuh koneksi database.
	if cmd == "create" {
		return create(*dir, rest)
	}

	switch cmd {
	case "up", "down", "status", "fresh", "force":
	case "":
		printUsage(flags)
		return errors.New("perintah belum diisi")
	default:
		printUsage(flags)
		return fmt.Errorf("perintah %q tidak dikenal", cmd)
	}

	// Baca dan cek semua file migrasi dulu, sebelum menyentuh database.
	migrations, err := loadMigrations(*dir)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	url, err := databaseURL()
	if err != nil {
		return err
	}

	// Koneksi langsung hanya untuk "fresh" dan konversi catatan lama.
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("koneksi database: %w", err)
	}
	defer func() { _ = conn.Close(context.Background()) }()

	if cmd == "fresh" {
		if err := dropAllTables(ctx, conn); err != nil {
			return err
		}
		cmd = "up"
	}
	if err := convertLegacyTable(ctx, conn, migrations); err != nil {
		return err
	}

	m, err := newMigrate(url, *dir)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	// Ctrl+C: golang-migrate berhenti setelah migrasi yang sedang berjalan.
	go func() {
		<-ctx.Done()
		select {
		case m.GracefulStop <- true:
		default:
		}
	}()

	switch cmd {
	case "up":
		return up(m, migrations)
	case "down":
		steps := 1
		if len(rest) > 0 {
			steps, err = strconv.Atoi(rest[0])
			if err != nil || steps < 1 {
				return fmt.Errorf("jumlah langkah %q harus angka minimal 1", rest[0])
			}
		}
		return down(m, migrations, steps)
	case "force":
		if len(rest) == 0 {
			return errors.New("versi belum diisi, contoh: force 20261113100200")
		}
		v, err := strconv.Atoi(rest[0])
		if err != nil {
			return fmt.Errorf("versi %q harus angka", rest[0])
		}
		if err := m.Force(v); err != nil {
			return fmt.Errorf("force %d: %w", v, err)
		}
		fmt.Printf("Versi database ditandai %d (tidak dirty).\n", v)
		return nil
	default: // status
		return status(m, migrations)
	}
}

// databaseURL membaca DB_URL / DATABASE_URL (file .env ikut dibaca bila ada).
func databaseURL() (string, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("baca .env: %w", err)
	}
	url := strings.TrimSpace(os.Getenv("DB_URL"))
	if url == "" {
		url = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if url == "" {
		return "", errors.New("DB_URL (atau DATABASE_URL) belum di-set")
	}
	return url, nil
}

// newMigrate menyiapkan golang-migrate: sumber berkas dari folder dir dan
// database PostgreSQL lewat driver pgx v5.
func newMigrate(url, dir string) (*migrate.Migrate, error) {
	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("baca DB_URL: %w", err)
	}
	db := stdlib.OpenDB(*cfg)
	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("siapkan driver migrasi: %w", err)
	}
	src, err := iofs.New(os.DirFS(dir), ".")
	if err != nil {
		return nil, fmt.Errorf("baca folder %s: %w", dir, err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "pgx5", driver)
	if err != nil {
		return nil, fmt.Errorf("siapkan golang-migrate: %w", err)
	}
	return m, nil
}

// convertLegacyTable memindahkan catatan dari format runner lama (satu baris
// per versi, kolom name) ke format golang-migrate (satu baris: versi terakhir
// dan dirty). Hanya dilakukan bila versi yang tercatat sama persis dengan
// urutan berkas migrasi dari awal; selain itu ditolak supaya tidak salah catat.
func convertLegacyTable(ctx context.Context, conn *pgx.Conn, migrations []migration) error {
	var legacy bool
	err := conn.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM information_schema.columns
		                WHERE table_schema = current_schema()
		                  AND table_name = 'schema_migrations' AND column_name = 'name')`).Scan(&legacy)
	if err != nil {
		return fmt.Errorf("cek schema_migrations: %w", err)
	}
	if !legacy {
		return nil
	}

	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("baca schema_migrations lama: %w", err)
	}
	versions, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return fmt.Errorf("baca schema_migrations lama: %w", err)
	}
	for i, v := range versions {
		if i >= len(migrations) || migrations[i].version != v {
			return fmt.Errorf("catatan migrasi lama tidak berurutan sesuai berkas (versi %d); "+
				"periksa schema_migrations secara manual sebelum memakai golang-migrate", v)
		}
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `DROP TABLE schema_migrations`); err != nil {
		return fmt.Errorf("konversi schema_migrations: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`CREATE TABLE schema_migrations (version BIGINT NOT NULL PRIMARY KEY, dirty BOOLEAN NOT NULL)`); err != nil {
		return fmt.Errorf("konversi schema_migrations: %w", err)
	}
	if len(versions) > 0 {
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, dirty) VALUES ($1, false)`,
			versions[len(versions)-1]); err != nil {
			return fmt.Errorf("konversi schema_migrations: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	fmt.Printf("Catatan migrasi lama (%d versi) dipindah ke format golang-migrate.\n", len(versions))
	return nil
}

// current mengembalikan versi database (0 bila belum ada migrasi).
func current(m *migrate.Migrate) (int64, bool, error) {
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("baca versi database: %w", err)
	}
	return int64(v), dirty, nil
}

func dirtyHint(m *migrate.Migrate, migrations []migration, err error) error {
	v, dirty, verr := current(m)
	if verr != nil || !dirty {
		return err
	}
	prev := int64(-1)
	for _, mg := range migrations {
		if mg.version < v {
			prev = mg.version
		}
	}
	return fmt.Errorf("%w\npetunjuk: database ditandai dirty di versi %d. Perbaiki penyebabnya, lalu jalankan:\n"+
		"  go run ./cmd/migration force %d\nkemudian ulangi perintah tadi", err, v, prev)
}

// up menjalankan semua migrasi yang belum jalan.
func up(m *migrate.Migrate, migrations []migration) error {
	before, _, err := current(m)
	if err != nil {
		return err
	}
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("Database sudah versi terbaru, tidak ada migrasi baru.")
		return nil
	}
	after, dirty, verr := current(m)
	if verr != nil {
		return verr
	}
	count := 0
	for _, mg := range migrations {
		if mg.version > before && mg.version <= after && (!dirty || mg.version != after) {
			fmt.Println("up  ", mg.label())
			count++
		}
	}
	if err != nil {
		return dirtyHint(m, migrations, err)
	}
	fmt.Printf("Selesai: %d migrasi dijalankan.\n", count)
	return nil
}

// down membatalkan migrasi terakhir sebanyak steps.
func down(m *migrate.Migrate, migrations []migration, steps int) error {
	before, _, err := current(m)
	if err != nil {
		return err
	}
	var done []migration
	for _, i := range slices.Backward(migrations) {
		if i.version <= before {
			done = append(done, i)
		}
	}
	if len(done) == 0 {
		fmt.Println("Belum ada migrasi yang dijalankan, tidak ada yang dibatalkan.")
		return nil
	}
	if steps > len(done) {
		steps = len(done)
	}
	err = m.Steps(-steps)
	after, _, verr := current(m)
	if verr != nil {
		return verr
	}
	count := 0
	for _, mg := range done {
		if mg.version > after {
			fmt.Println("down", mg.label())
			count++
		}
	}
	if err != nil {
		return dirtyHint(m, migrations, err)
	}
	fmt.Printf("Selesai: %d migrasi dibatalkan.\n", count)
	return nil
}

// status mencetak daftar migrasi beserta tanda sudah/belum dijalankan.
func status(m *migrate.Migrate, migrations []migration) error {
	v, dirty, err := current(m)
	if err != nil {
		return err
	}
	pending := 0
	for _, mg := range migrations {
		mark := "[x]"
		switch {
		case dirty && mg.version == v:
			mark = "[!]"
		case mg.version > v:
			mark = "[ ]"
			pending++
		}
		fmt.Println(mark, mg.label())
	}
	fmt.Printf("\n%d sudah jalan, %d belum. Versi database: %d\n", len(migrations)-pending, pending, v)
	if dirty {
		fmt.Printf("PERHATIAN: versi %d dirty (migrasi terakhir gagal). Lihat: go run ./cmd/migration force\n", v)
	}
	return nil
}

// dropAllTables menghapus SEMUA tabel di schema aktif (biasanya public).
// Hanya diizinkan saat APP_ENV kosong atau "dev" supaya tidak terjadi di server.
func dropAllTables(ctx context.Context, conn *pgx.Conn) error {
	if env := strings.TrimSpace(os.Getenv("APP_ENV")); env != "" && env != "dev" {
		return fmt.Errorf("fresh ditolak: APP_ENV=%q, hanya boleh saat dev", env)
	}

	rows, err := conn.Query(ctx,
		"SELECT quote_ident(tablename) FROM pg_tables WHERE schemaname = current_schema()")
	if err != nil {
		return fmt.Errorf("daftar tabel: %w", err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("daftar tabel: %w", err)
	}

	if len(tables) == 0 {
		fmt.Println("Database sudah kosong.")
		return nil
	}

	// CASCADE ikut menghapus foreign key dan sequence milik tabel tersebut.
	if _, err := conn.Exec(ctx, "DROP TABLE "+strings.Join(tables, ", ")+" CASCADE"); err != nil {
		return fmt.Errorf("hapus tabel: %w", err)
	}

	fmt.Printf("Dihapus %d tabel: %s\n", len(tables), strings.Join(tables, ", "))
	return nil
}

// loadMigrations membaca folder migrasi, memasangkan file up/down per versi,
// lalu mengurutkannya dari versi terkecil. golang-migrate butuh file down
// untuk setiap versi, jadi versi tanpa .down.sql ditolak di sini.
func loadMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("baca folder %s: %w", dir, err)
	}

	byVersion := map[int64]*migration{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := fileName.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		version, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("versi pada %s terlalu besar", e.Name())
		}
		name, direction := m[2], m[3]
		path := filepath.Join(dir, e.Name())

		mig, ok := byVersion[version]
		if !ok {
			mig = &migration{version: version, name: name}
			byVersion[version] = mig
		}
		if mig.name != name {
			return nil, fmt.Errorf("versi %d dipakai dua nama berbeda: %s dan %s", version, mig.name, name)
		}
		target := &mig.upPath
		if direction == "down" {
			target = &mig.downPath
		}
		if *target != "" {
			return nil, fmt.Errorf("file %s ganda untuk versi %d", direction, version)
		}
		*target = path
	}

	list := make([]migration, 0, len(byVersion))
	for _, m := range byVersion {
		if m.upPath == "" {
			return nil, fmt.Errorf("versi %d (%s) tidak punya file .up.sql", m.version, m.name)
		}
		if m.downPath == "" {
			return nil, fmt.Errorf("versi %d (%s) tidak punya file .down.sql", m.version, m.name)
		}
		list = append(list, *m)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].version < list[j].version })
	return list, nil
}

// create membuat pasangan file migrasi kosong dengan versi berupa waktu sekarang (UTC).
func create(dir string, args []string) error {
	if len(args) == 0 {
		return errors.New("nama migrasi belum diisi, contoh: create add_phone_to_users")
	}

	// Ubah nama jadi huruf kecil dengan garis bawah, misal "Add Phone" -> "add_phone".
	name := strings.ToLower(strings.Join(args, "_"))
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		return errors.New("nama migrasi hanya boleh huruf dan angka")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	base := time.Now().UTC().Format("20060102150405") + "_" + name
	for _, direction := range []string{"up", "down"} {
		path := filepath.Join(dir, base+"."+direction+".sql")

		// O_EXCL: gagal bila file sudah ada, jadi file lama tidak tertimpa.
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(f, "-- %s: tulis perintah SQL di sini.\n", direction)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
		fmt.Println("dibuat", path)
	}
	return nil
}

func printUsage(flags *flag.FlagSet) {
	fmt.Fprintln(os.Stderr, `Pemakaian:
  go run ./cmd/migration [opsi] up
  go run ./cmd/migration [opsi] down [jumlah]
  go run ./cmd/migration [opsi] status
  go run ./cmd/migration [opsi] force <versi>
  go run ./cmd/migration [opsi] fresh      (dev saja: hapus semua tabel lalu up)
  go run ./cmd/migration [opsi] create <nama>

Opsi:`)
	flags.PrintDefaults()
}
