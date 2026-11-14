package ongkir

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	"hii-stranger1/backend/internal/pkg/rajaongkir"
)

// Service berisi logika wilayah dan ongkir.
type Service struct {
	client   *rajaongkir.Client
	repo     *Repository
	originID int    // ID kelurahan asal toko (RAJAONGKIR_ORIGIN_ID)
	couriers string // kurir yang dipakai toko, contoh "pos" (bisa "jne:pos")
}

// NewService membuat Service.
func NewService(client *rajaongkir.Client, repo *Repository, originID int, couriers string) *Service {
	return &Service{client: client, repo: repo, originID: originID, couriers: couriers}
}

// Provinces mengambil daftar provinsi.
func (s *Service) Provinces(ctx context.Context) ([]rajaongkir.Region, error) {
	return mapErr(s.client.Provinces(ctx))
}

// Cities mengambil kota/kabupaten dalam provinsi.
func (s *Service) Cities(ctx context.Context, provinceID int) ([]rajaongkir.Region, error) {
	return mapErr(s.client.Cities(ctx, provinceID))
}

// Districts mengambil kecamatan dalam kota/kabupaten.
func (s *Service) Districts(ctx context.Context, cityID int) ([]rajaongkir.Region, error) {
	return mapErr(s.client.Districts(ctx, cityID))
}

// SubDistricts mengambil kelurahan/desa dalam kecamatan.
func (s *Service) SubDistricts(ctx context.Context, districtID int) ([]rajaongkir.Region, error) {
	return mapErr(s.client.SubDistricts(ctx, districtID))
}

// Resolve memeriksa bahwa provinsi -> kota -> kecamatan -> kelurahan benar-benar
// berjenjang (kota ada di provinsi itu, dst.) lalu mengembalikan namanya.
// Data wilayah di-cache oleh client, jadi pemeriksaan ini tidak mahal.
func (s *Service) Resolve(ctx context.Context, d Destination) (Address, error) {
	province, err := pick(s.Provinces(ctx))(d.ProvinceID)
	if err != nil {
		return Address{}, err
	}
	city, err := pick(s.Cities(ctx, d.ProvinceID))(d.CityID)
	if err != nil {
		return Address{}, err
	}
	district, err := pick(s.Districts(ctx, d.CityID))(d.DistrictID)
	if err != nil {
		return Address{}, err
	}
	sub, err := pick(s.SubDistricts(ctx, d.DistrictID))(d.SubdistrictID)
	if err != nil {
		return Address{}, err
	}
	return Address{
		Province:      province.Name,
		City:          city.Name,
		District:      district.Name,
		Subdistrict:   sub.Name,
		SubdistrictID: sub.ID,
		ZipCode:       sub.ZipCode,
	}, nil
}

// pick mengembalikan fungsi pencari wilayah ber-ID tertentu dari hasil
// pemanggilan daftar wilayah. ErrInvalidDestination bila ID tidak ada.
func pick(regions []rajaongkir.Region, err error) func(id int) (rajaongkir.Region, error) {
	return func(id int) (rajaongkir.Region, error) {
		if err != nil {
			return rajaongkir.Region{}, err
		}
		for _, r := range regions {
			if r.ID == id {
				return r, nil
			}
		}
		return rajaongkir.Region{}, ErrInvalidDestination
	}
}

// options mengambil semua layanan kurir toko dari kelurahan toko ke kelurahan
// tujuan untuk berat tertentu, urut dari yang termurah.
func (s *Service) options(ctx context.Context, subdistrictID, weightGram int) ([]Option, error) {
	if !s.client.Enabled() || s.originID <= 0 {
		return nil, ErrDisabled
	}
	costs, err := s.client.DomesticCost(ctx, s.originID, subdistrictID, weightGram, s.couriers)
	if errors.Is(err, rajaongkir.ErrNotFound) {
		return nil, ErrNoRoute
	}
	if err != nil {
		return nil, err
	}

	out := make([]Option, 0, len(costs))
	for _, c := range costs {
		out = append(out, Option{
			CourierCode: strings.ToLower(c.Code),
			CourierName: c.Name,
			Service:     c.Service,
			Description: c.Description,
			Cost:        strconv.FormatInt(c.Cost, 10),
			ETD:         c.ETD,
		})
	}
	if len(out) == 0 {
		return nil, ErrNoRoute
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, _ := strconv.ParseInt(out[i].Cost, 10, 64)
		b, _ := strconv.ParseInt(out[j].Cost, 10, 64)
		return a < b
	})
	return out, nil
}

// Cheapest mengambil layanan termurah dari kurir toko (RAJAONGKIR_COURIERS)
// ke kelurahan tujuan. Pelanggan tidak memilih kurir; toko yang menentukan.
// Dipakai cek ongkir dan checkout, jadi harganya selalu dihitung server.
func (s *Service) Cheapest(ctx context.Context, subdistrictID, weightGram int) (Option, error) {
	opts, err := s.options(ctx, subdistrictID, weightGram)
	if err != nil {
		return Option{}, err
	}
	return opts[0], nil
}

// CartWeight menghitung berat keranjang user (gram). ErrCartEmpty bila tidak
// ada item yang dihitung.
func (s *Service) CartWeight(ctx context.Context, userID string, ids []string) (int, error) {
	weight, count, err := s.repo.CartWeight(ctx, userID, ids)
	if err != nil {
		return 0, err
	}
	if count == 0 || (len(ids) > 0 && count != len(ids)) {
		return 0, ErrCartEmpty
	}
	return weight, nil
}

// QuoteCart menghitung ongkir isi keranjang user ke kelurahan tujuan.
func (s *Service) QuoteCart(ctx context.Context, userID string, ids []string, subdistrictID int) (Quote, error) {
	weight, err := s.CartWeight(ctx, userID, ids)
	if err != nil {
		return Quote{}, err
	}
	opt, err := s.Cheapest(ctx, subdistrictID, weight)
	if err != nil {
		return Quote{}, err
	}
	return Quote{WeightGram: weight, Option: opt}, nil
}

// mapErr menyamakan error "belum dikonfigurasi" dari client.
func mapErr(r []rajaongkir.Region, err error) ([]rajaongkir.Region, error) {
	if errors.Is(err, rajaongkir.ErrDisabled) {
		return nil, ErrDisabled
	}
	return r, err
}
