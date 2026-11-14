// Package ongkir menyediakan data wilayah berjenjang (provinsi -> kota ->
// kecamatan -> kelurahan) dan menghitung ongkos kirim sampai tingkat
// kelurahan memakai API RajaOngkir. Berat paket dihitung dari keranjang.
package ongkir

import "errors"

var (
	// ErrDisabled berarti API key atau kelurahan asal toko belum diisi.
	ErrDisabled = errors.New("ongkir: not configured")
	// ErrInvalidDestination berarti susunan provinsi/kota/kecamatan/kelurahan
	// tidak cocok satu sama lain (mis. kelurahan bukan bagian kecamatan itu).
	ErrInvalidDestination = errors.New("ongkir: invalid destination")
	// ErrNoRoute berarti tidak ada kurir yang melayani tujuan.
	ErrNoRoute = errors.New("ongkir: destination not covered")
	// ErrCartEmpty berarti keranjang kosong sehingga berat tidak bisa dihitung.
	ErrCartEmpty = errors.New("ongkir: cart is empty")
)

// Destination adalah ID wilayah tujuan yang dipilih pelanggan dari dropdown.
type Destination struct {
	ProvinceID    int `json:"province_id"`
	CityID        int `json:"city_id"`
	DistrictID    int `json:"district_id"`
	SubdistrictID int `json:"subdistrict_id"`
}

// Address adalah nama wilayah tujuan yang sudah divalidasi ke RajaOngkir.
// Nama inilah yang disimpan di order, bukan teks ketikan pelanggan.
type Address struct {
	Province      string
	City          string
	District      string // kecamatan
	Subdistrict   string // kelurahan/desa
	SubdistrictID int
	ZipCode       string
}

// Option adalah layanan kurir beserta tarifnya. Cost berupa string desimal
// agar seragam dengan nilai uang lain di API.
type Option struct {
	CourierCode string `json:"courier_code"`
	CourierName string `json:"courier_name"`
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        string `json:"cost"`
	ETD         string `json:"etd"`
}

// Quote adalah hasil cek ongkir untuk isi keranjang (schema ShippingCostQuote):
// berat paket dan layanan kurir yang akan dipakai toko.
type Quote struct {
	WeightGram int `json:"weight_gram"`
	Option
}
