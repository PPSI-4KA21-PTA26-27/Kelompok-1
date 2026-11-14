package fileupload

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	imageMax = 5 << 20 // batas gambar produk/banner
	proofMax = 2 << 20 // batas bukti pembayaran
)

var imageTypes = []string{"image/jpeg", "image/png", "image/webp"}

// pngHeader cukup untuk dikenali http.DetectContentType sebagai image/png.
var pngHeader = []byte("\x89PNG\r\n\x1a\n")

func pngOf(size int) []byte {
	b := make([]byte, size)
	copy(b, pngHeader)
	return b
}

func multipartRequest(t *testing.T, field string, data []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if field != "" {
		fw, err := mw.CreateFormFile(field, "berkas.bin")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestReadAccepts(t *testing.T) {
	rec := httptest.NewRecorder()
	data, ct, ok := Read(rec, multipartRequest(t, "image", pngOf(1024)), "image", imageMax, imageTypes)
	if !ok {
		t.Fatalf("Read() ok = false, status %d body %s", rec.Code, rec.Body.String())
	}
	if ct != "image/png" || len(data) != 1024 {
		t.Errorf("Read() = %d byte, %q; want 1024 byte image/png", len(data), ct)
	}
}

func TestReadExactlyAtLimit(t *testing.T) {
	rec := httptest.NewRecorder()
	if _, _, ok := Read(rec, multipartRequest(t, "proof", pngOf(proofMax)), "proof", proofMax, imageTypes); !ok {
		t.Fatalf("berkas tepat 2 MB harus diterima, status %d", rec.Code)
	}
}

func TestReadRejects(t *testing.T) {
	tests := []struct {
		name     string
		req      func(t *testing.T) *http.Request
		field    string
		max      int64
		want     int
		wantCode string
	}{
		{"bukan multipart", func(t *testing.T) *http.Request {
			r := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("{}"))
			r.Header.Set("Content-Type", "application/json")
			return r
		}, "image", imageMax, http.StatusBadRequest, "bad_request"},
		{"field tidak ada", func(t *testing.T) *http.Request {
			return multipartRequest(t, "", nil)
		}, "image", imageMax, http.StatusBadRequest, "bad_request"},
		{"field lain saja", func(t *testing.T) *http.Request {
			return multipartRequest(t, "file", pngOf(10))
		}, "image", imageMax, http.StatusBadRequest, "bad_request"},
		{"berkas kosong", func(t *testing.T) *http.Request {
			return multipartRequest(t, "image", nil)
		}, "image", imageMax, http.StatusBadRequest, "bad_request"},
		{"gambar lebih dari 5 MB", func(t *testing.T) *http.Request {
			return multipartRequest(t, "image", pngOf(imageMax+1))
		}, "image", imageMax, http.StatusRequestEntityTooLarge, "payload_too_large"},
		{"bukti lebih dari 2 MB", func(t *testing.T) *http.Request {
			return multipartRequest(t, "proof", pngOf(proofMax+1))
		}, "proof", proofMax, http.StatusRequestEntityTooLarge, "payload_too_large"},
		{"isi teks walau bernama gambar", func(t *testing.T) *http.Request {
			return multipartRequest(t, "image", []byte("ini bukan gambar sama sekali"))
		}, "image", imageMax, http.StatusUnsupportedMediaType, "unsupported_media_type"},
		{"isi PDF", func(t *testing.T) *http.Request {
			return multipartRequest(t, "image", []byte("%PDF-1.4\n%...."))
		}, "image", imageMax, http.StatusUnsupportedMediaType, "unsupported_media_type"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if _, _, ok := Read(rec, tc.req(t), tc.field, tc.max, imageTypes); ok {
				t.Fatal("Read() ok = true, want false")
			}
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
			if !strings.Contains(rec.Body.String(), tc.wantCode) {
				t.Errorf("body = %s, want kode %q", rec.Body.String(), tc.wantCode)
			}
		})
	}
}
