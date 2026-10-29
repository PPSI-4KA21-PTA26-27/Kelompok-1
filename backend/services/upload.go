package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// Alur penyimpanan: handler upload admin (tahap 6) -> ImageStore.Save -> folder uploads/.
// Alur penyajian:   handlers.UploadHandler -> ImageStore.Path -> file di folder uploads/ -> klien.
// File tidak pernah disajikan langsung dari disk publik; selalu lewat handler, dan nama file
// yang diminta klien dicocokkan dulu dengan bentuk nama yang dibuat Save (mencegah path traversal).

// maxImageBytes: batas ukuran gambar, sama dengan "maks 2MB" di form admin banner.
const maxImageBytes = 2 << 20

var (
	ErrImageTooLarge = errors.New("ukuran gambar maksimal 2 MB")
	ErrImageType     = errors.New("format gambar harus JPG, PNG, atau WebP")
)

// imageExt memetakan tipe isi file (hasil deteksi dari isinya, bukan dari nama atau header klien)
// ke ekstensi yang disimpan. Tipe lain ditolak, termasuk SVG karena bisa memuat script.
var imageExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// storedImageName adalah bentuk nama file yang dibuat Save: 32 karakter hex acak + ekstensi.
var storedImageName = regexp.MustCompile(`^[0-9a-f]{32}\.(jpg|png|webp)$`)

// ImageStore menyimpan gambar unggahan (produk, banner) di satu folder.
type ImageStore struct {
	dir string
}

// NewImageStore menyiapkan folder penyimpanan; folder dibuat kalau belum ada.
func NewImageStore(dir string) (*ImageStore, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("buat folder gambar: %w", err)
	}
	return &ImageStore{dir: dir}, nil
}

// Save memeriksa gambar lalu menyimpannya dengan nama acak. Mengembalikan nama file yang
// disimpan (bukan path), contoh "3f9a...c1.webp". Nama dari klien tidak pernah dipakai.
func (s *ImageStore) Save(src io.Reader) (string, error) {
	// Baca satu byte lebih banyak dari batas, supaya file yang kebesaran ketahuan
	// tanpa harus membaca seluruh kiriman.
	data, err := io.ReadAll(io.LimitReader(src, maxImageBytes+1))
	if err != nil {
		return "", fmt.Errorf("baca gambar: %w", err)
	}
	if len(data) > maxImageBytes {
		return "", ErrImageTooLarge
	}

	ext, ok := imageExt[http.DetectContentType(data)]
	if !ok {
		return "", ErrImageType
	}

	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("buat nama file: %w", err)
	}
	name := hex.EncodeToString(raw) + ext
	path := filepath.Join(s.dir, name)

	// O_EXCL membuat penyimpanan gagal kalau nama sudah ada, jadi file lama tidak pernah tertimpa.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", fmt.Errorf("buat file gambar: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		// Yang dilaporkan adalah error tulis; membersihkan file setengah jadi cukup dicoba saja.
		_ = f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("tulis gambar: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("tutup file gambar: %w", err)
	}
	return name, nil
}

// Path mengembalikan lokasi file di disk untuk nama yang diminta klien. ok bernilai false kalau
// nama tidak berbentuk nama buatan Save, misalnya "../.env" atau "a/b.png".
func (s *ImageStore) Path(name string) (string, bool) {
	if !storedImageName.MatchString(name) {
		return "", false
	}
	return filepath.Join(s.dir, name), true
}
