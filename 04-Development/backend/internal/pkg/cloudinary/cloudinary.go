// Package cloudinary mengunggah gambar ke Cloudinary memakai Upload API
// bertanda tangan (signed upload). Hanya memakai standard library.
package cloudinary

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	apiBase = "https://api.cloudinary.com/v1_1/"

	// requestTimeout membatasi lama satu unggahan ke Cloudinary.
	requestTimeout = 30 * time.Second
)

// ErrDisabled berarti kredensial Cloudinary belum diisi di environment.
var ErrDisabled = errors.New("cloudinary: not configured")

// Client menyimpan kredensial Cloudinary dan HTTP client dengan timeout.
type Client struct {
	cloudName string
	apiKey    string
	apiSecret string
	folder    string
	http      *http.Client
}

// New membuat Client. folder boleh kosong (gambar masuk ke root akun).
func New(cloudName, apiKey, apiSecret, folder string) *Client {
	return &Client{
		cloudName: cloudName,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		folder:    folder,
		http:      &http.Client{Timeout: requestTimeout},
	}
}

// Enabled bernilai true bila semua kredensial sudah diisi.
func (c *Client) Enabled() bool {
	return c != nil && c.cloudName != "" && c.apiKey != "" && c.apiSecret != ""
}

// deliveryHost adalah host penyajian berkas Cloudinary; Fetch hanya mau
// mengambil berkas dari host ini.
const deliveryHost = "res.cloudinary.com"

// UploadImage mengirim data gambar ke Cloudinary dan mengembalikan secure_url
// (alamat https) yang bisa langsung disimpan sebagai image_url.
func (c *Client) UploadImage(ctx context.Context, filename string, data []byte) (string, error) {
	return c.upload(ctx, "image", "", nil, filename, data)
}

// UploadProof mengunggah bukti pembayaran ke subfolder payment-proofs dan
// mengembalikan secure_url-nya. Gambar diunggah sebagai resource image; PDF
// sebagai resource raw dengan public_id berakhiran .pdf, karena penyajian
// PDF sebagai image sering dinonaktifkan pada akun Cloudinary.
func (c *Client) UploadProof(ctx context.Context, contentType string, data []byte) (string, error) {
	if contentType == "application/pdf" {
		id, err := randomID()
		if err != nil {
			return "", err
		}
		return c.upload(ctx, "raw", "payment-proofs", map[string]string{"public_id": id + ".pdf"}, "proof.pdf", data)
	}
	return c.upload(ctx, "image", "payment-proofs", nil, "proof", data)
}

// Fetch mengambil berkas dari alamat Cloudinary (secure_url hasil unggah)
// dengan batas maxBytes. Alamat selain https://res.cloudinary.com ditolak
// supaya server tidak bisa dipakai mengambil alamat sembarang.
func (c *Client) Fetch(ctx context.Context, rawURL string, maxBytes int64) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Host != deliveryHost {
		return nil, fmt.Errorf("cloudinary: fetch: alamat bukan Cloudinary: %q", rawURL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: fetch request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: fetch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudinary: fetch status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cloudinary: fetch read: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("cloudinary: fetch: berkas melebihi %d byte", maxBytes)
	}
	return data, nil
}

// upload mengirim berkas ke Upload API. subfolder ditambahkan di bawah folder
// akun (boleh kosong); extra berisi parameter tambahan yang ikut ditandatangani.
func (c *Client) upload(ctx context.Context, resourceType, subfolder string, extra map[string]string, filename string, data []byte) (string, error) {
	if !c.Enabled() {
		return "", ErrDisabled
	}

	params := map[string]string{"timestamp": strconv.FormatInt(time.Now().Unix(), 10)}
	if folder := strings.Trim(c.folder+"/"+subfolder, "/"); folder != "" {
		params["folder"] = folder
	}
	for k, v := range extra {
		params[k] = v
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fields := map[string]string{
		"api_key":   c.apiKey,
		"signature": c.sign(params),
	}
	for k, v := range params {
		fields[k] = v
	}
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return "", fmt.Errorf("cloudinary: write field %s: %w", k, err)
		}
	}
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("cloudinary: create file part: %w", err)
	}
	if _, err := fw.Write(data); err != nil {
		return "", fmt.Errorf("cloudinary: write file part: %w", err)
	}
	if err := mw.Close(); err != nil {
		return "", fmt.Errorf("cloudinary: close multipart: %w", err)
	}

	endpoint := apiBase + url.PathEscape(c.cloudName) + "/" + resourceType + "/upload"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("cloudinary: new request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudinary: upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var out struct {
		SecureURL string `json:"secure_url"`
		Error     struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	// Batasi pembacaan response supaya tidak membaca body yang tidak wajar.
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		return "", fmt.Errorf("cloudinary: decode response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || out.SecureURL == "" {
		return "", fmt.Errorf("cloudinary: upload status %d: %s", resp.StatusCode, out.Error.Message)
	}
	return out.SecureURL, nil
}

// sign membuat signature sesuai aturan Cloudinary: parameter (selain file,
// api_key, dan signature) diurutkan menurut nama, digabung "a=1&b=2", lalu
// ditambah api secret dan di-hash SHA-1.
func (c *Client) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, len(keys))
	for i, k := range keys {
		pairs[i] = k + "=" + params[k]
	}
	sum := sha1.Sum([]byte(strings.Join(pairs, "&") + c.apiSecret))
	return hex.EncodeToString(sum[:])
}

// randomID membuat nama berkas acak (32 karakter hex) untuk public_id.
func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("cloudinary: random id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
