package services

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

// Header minimal supaya isi file dikenali sebagai gambar; sisanya diisi nol.
var (
	jpegHeader = []byte{0xFF, 0xD8, 0xFF, 0xE0}
	pngHeader  = []byte("\x89PNG\r\n\x1a\n")
	webpHeader = []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
)

func fakeImage(header []byte, size int) []byte {
	b := make([]byte, size)
	copy(b, header)
	return b
}

func TestImageStoreSaveAcceptsSupportedImages(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
		ext    string
	}{
		{"jpeg", jpegHeader, ".jpg"},
		{"png", pngHeader, ".png"},
		{"webp", webpHeader, ".webp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			store, err := NewImageStore(dir)
			if err != nil {
				t.Fatal(err)
			}
			data := fakeImage(tt.header, 1024)

			name, err := store.Save(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}
			if !storedImageName.MatchString(name) || name[len(name)-len(tt.ext):] != tt.ext {
				t.Errorf("Save() nama = %q, want 32 hex + %q", name, tt.ext)
			}
			saved, err := os.ReadFile(dir + string(os.PathSeparator) + name)
			if err != nil {
				t.Fatalf("file hasil Save tidak bisa dibaca: %v", err)
			}
			if !bytes.Equal(saved, data) {
				t.Error("isi file tersimpan berbeda dari yang dikirim")
			}
		})
	}
}

func TestImageStoreSaveRejects(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{"teks biasa", []byte("bukan gambar"), ErrImageType},
		{"svg (bisa memuat script)", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), ErrImageType},
		{"html yang diberi nama gambar", []byte("<html><body>x</body></html>"), ErrImageType},
		{"file kosong", nil, ErrImageType},
		{"satu byte melewati batas 2 MB", fakeImage(jpegHeader, maxImageBytes+1), ErrImageTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			store, err := NewImageStore(dir)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := store.Save(bytes.NewReader(tt.data)); !errors.Is(err, tt.want) {
				t.Errorf("Save() error = %v, want %v", err, tt.want)
			}
			// Gambar yang ditolak tidak boleh meninggalkan file di folder.
			if entries, _ := os.ReadDir(dir); len(entries) != 0 {
				t.Errorf("folder berisi %d file setelah Save ditolak, want 0", len(entries))
			}
		})
	}
}

func TestImageStoreSaveAcceptsExactLimit(t *testing.T) {
	store, err := NewImageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(bytes.NewReader(fakeImage(jpegHeader, maxImageBytes))); err != nil {
		t.Errorf("Save() file tepat 2 MB error = %v, want nil", err)
	}
}

func TestImageStorePath(t *testing.T) {
	store, err := NewImageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	valid := "0123456789abcdef0123456789abcdef"

	tests := []struct {
		name   string
		file   string
		wantOK bool
	}{
		{"nama buatan Save", valid + ".webp", true},
		{"jpg", valid + ".jpg", true},
		{"path traversal", "../.env", false},
		{"path traversal dengan ekstensi valid", "../" + valid + ".jpg", false},
		{"subfolder", "a/" + valid + ".jpg", false},
		{"backslash Windows", `..\` + valid + ".jpg", false},
		{"huruf besar", "0123456789ABCDEF0123456789ABCDEF.jpg", false},
		{"terlalu pendek", "abc.jpg", false},
		{"ekstensi tidak diizinkan", valid + ".svg", false},
		{"tanpa ekstensi", valid, false},
		{"kosong", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := store.Path(tt.file); ok != tt.wantOK {
				t.Errorf("Path(%q) ok = %v, want %v", tt.file, ok, tt.wantOK)
			}
		})
	}
}
