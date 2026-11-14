// Package rajaongkir adalah client RajaOngkir (Komerce) untuk data wilayah
// Indonesia (provinsi, kota, kecamatan, kelurahan) dan hitung ongkir sampai
// tingkat kelurahan. Dokumentasi: https://www.rajaongkir.com/docs
//
// Data wilayah jarang berubah, jadi disimpan di memori (cache) supaya tidak
// menghabiskan kuota API setiap kali frontend membuka dropdown alamat.
package rajaongkir

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultBaseURL adalah alamat API RajaOngkir versi Komerce.
	DefaultBaseURL = "https://rajaongkir.komerce.id/api/v1"

	requestTimeout = 20 * time.Second
	regionCacheTTL = 24 * time.Hour
	maxResponse    = 4 << 20
)

var (
	// ErrDisabled berarti API key RajaOngkir belum diisi.
	ErrDisabled = errors.New("rajaongkir: not configured")
	// ErrNotFound berarti RajaOngkir tidak menemukan data (mis. rute tidak dilayani).
	ErrNotFound = errors.New("rajaongkir: not found")
)

// Region adalah satu wilayah: provinsi, kota/kabupaten, kecamatan, atau
// kelurahan. ZipCode hanya terisi pada sebagian level.
type Region struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	ZipCode string `json:"zip_code,omitempty"`
}

// Cost adalah satu pilihan layanan kurir beserta tarifnya (rupiah).
type Cost struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        int64  `json:"cost"`
	ETD         string `json:"etd"`
}

// Client memanggil API RajaOngkir.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	regions []Region
	expires time.Time
}

// New membuat Client. baseURL kosong berarti DefaultBaseURL.
func New(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: requestTimeout},
		cache:   map[string]cacheEntry{},
	}
}

// Enabled bernilai true bila API key sudah diisi.
func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != ""
}

// Provinces mengambil semua provinsi.
func (c *Client) Provinces(ctx context.Context) ([]Region, error) {
	return c.regions(ctx, "/destination/province")
}

// Cities mengambil kota/kabupaten dalam provinsi.
func (c *Client) Cities(ctx context.Context, provinceID int) ([]Region, error) {
	return c.regions(ctx, "/destination/city/"+strconv.Itoa(provinceID))
}

// Districts mengambil kecamatan dalam kota/kabupaten.
func (c *Client) Districts(ctx context.Context, cityID int) ([]Region, error) {
	return c.regions(ctx, "/destination/district/"+strconv.Itoa(cityID))
}

// SubDistricts mengambil kelurahan/desa dalam kecamatan.
func (c *Client) SubDistricts(ctx context.Context, districtID int) ([]Region, error) {
	return c.regions(ctx, "/destination/sub-district/"+strconv.Itoa(districtID))
}

// DomesticCost menghitung ongkir dari kelurahan origin ke kelurahan
// destination (ID dari SubDistricts). couriers dipisah titik dua, contoh
// "jne:sicepat:jnt". weightGram minimal 1.
func (c *Client) DomesticCost(ctx context.Context, origin, destination, weightGram int, couriers string) ([]Cost, error) {
	form := url.Values{
		"origin":      {strconv.Itoa(origin)},
		"destination": {strconv.Itoa(destination)},
		"weight":      {strconv.Itoa(max(weightGram, 1))},
		"courier":     {couriers},
		"price":       {"lowest"},
	}

	var raw []struct {
		Name        string  `json:"name"`
		Code        string  `json:"code"`
		Service     string  `json:"service"`
		Description string  `json:"description"`
		Cost        float64 `json:"cost"`
		ETD         string  `json:"etd"`
	}
	if err := c.do(ctx, http.MethodPost, "/calculate/domestic-cost", form, &raw); err != nil {
		return nil, err
	}

	out := make([]Cost, 0, len(raw))
	for _, r := range raw {
		out = append(out, Cost{
			Name: r.Name, Code: r.Code, Service: r.Service, Description: r.Description,
			Cost: int64(math.Round(r.Cost)), ETD: r.ETD,
		})
	}
	return out, nil
}

// regions mengambil daftar wilayah dengan cache. Wilayah yang tidak ada
// dikembalikan sebagai daftar kosong, bukan error.
func (c *Client) regions(ctx context.Context, path string) ([]Region, error) {
	c.mu.Lock()
	e, ok := c.cache[path]
	c.mu.Unlock()
	if ok && time.Now().Before(e.expires) {
		return e.regions, nil
	}

	var raw []struct {
		ID      int        `json:"id"`
		Name    string     `json:"name"`
		ZipCode flexString `json:"zip_code"`
	}
	err := c.do(ctx, http.MethodGet, path, nil, &raw)
	if errors.Is(err, ErrNotFound) {
		raw = nil
	} else if err != nil {
		return nil, err
	}

	out := make([]Region, 0, len(raw))
	for _, r := range raw {
		zip := string(r.ZipCode)
		if zip == "0" {
			zip = ""
		}
		out = append(out, Region{ID: r.ID, Name: r.Name, ZipCode: zip})
	}

	c.mu.Lock()
	c.cache[path] = cacheEntry{regions: out, expires: time.Now().Add(regionCacheTTL)}
	c.mu.Unlock()
	return out, nil
}

// do mengirim request dan membaca format {"meta": {...}, "data": ...}.
// form nil berarti GET tanpa body.
func (c *Client) do(ctx context.Context, method, path string, form url.Values, data any) error {
	if !c.Enabled() {
		return ErrDisabled
	}

	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("rajaongkir: new request: %w", err)
	}
	req.Header.Set("key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("rajaongkir: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var env struct {
		Meta struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
			Status  string `json:"status"`
		} `json:"meta"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponse)).Decode(&env); err != nil {
		return fmt.Errorf("rajaongkir: decode %s (status %d): %w", path, resp.StatusCode, err)
	}

	if resp.StatusCode == http.StatusNotFound || env.Meta.Code == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK || env.Meta.Status != "success" {
		return fmt.Errorf("rajaongkir: %s status %d: %s", path, resp.StatusCode, env.Meta.Message)
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return ErrNotFound
	}

	if err := json.Unmarshal(env.Data, data); err != nil {
		return fmt.Errorf("rajaongkir: decode data %s: %w", path, err)
	}
	return nil
}

// flexString menerima nilai JSON berupa teks maupun angka (zip_code kadang
// dikirim sebagai angka), dan null menjadi string kosong.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = flexString(n.String())
	return nil
}
