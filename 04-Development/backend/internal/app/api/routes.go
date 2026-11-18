package api

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"hii-stranger1/backend/internal/app/api/auth"
	"hii-stranger1/backend/internal/app/api/cart"
	"hii-stranger1/backend/internal/app/api/category"
	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/app/api/ongkir"
	"hii-stranger1/backend/internal/app/api/order"
	"hii-stranger1/backend/internal/app/api/payment"
	"hii-stranger1/backend/internal/app/api/product"
	"hii-stranger1/backend/internal/app/api/shipment"
	"hii-stranger1/backend/internal/app/api/upload"
	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/config"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/token"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// apiPrefix adalah base path kanonik (versi mayor). Versi lengkap dilayani
	// juga lewat alias /api/v<versi> (lihat Config.AppVersion dan LegacyAPIVersions).
	apiPrefix = "/api/v1"

	// Batas percobaan login, register, dan ganti password per IP, untuk mempersulit tebak password.
	authRateLimit  = 10
	authRateWindow = time.Minute

	// Batas per IP untuk endpoint publik yang sering dipanggil (pengaturan
	// situs, data wilayah, cek ongkir).
	publicRateLimit  = 120
	publicRateWindow = time.Minute

	// Batas per user (fallback per IP) untuk rute terautentikasi:
	// - userWrite: mutasi pelanggan (cart, order, profil, pembayaran); 1/detik
	//   cukup untuk klik cepat di keranjang.
	// - userRead: GET terautentikasi (profil, cart, order, detail pembayaran).
	// - upload: gambar (admin) dan bukti transfer; 20, bukan 10, karena admin
	//   biasa mengunggah beberapa foto produk sekaligus.
	// - admin: dashboard backoffice memanggil banyak endpoint per halaman.
	userWriteRateLimit = 60
	userReadRateLimit  = 120
	uploadRateLimit    = 20
	adminRateLimit     = 120
	userRateWindow     = time.Minute

	// Microcache di Redis untuk GET publik (produk, kategori,
	// pengaturan situs). Dikosongkan otomatis untuk semua instance setiap ada
	// perubahan data lewat API, jadi ttl hanya batas aman bila data diubah di
	// luar API.
	publicCacheTTL = 30 * time.Second
)

type handlers struct {
	auth     *auth.Handler
	cart     *cart.Handler
	category *category.Handler
	product  *product.Handler
	user     *user.Handler

	ongkir   *ongkir.Handler
	order    *order.Handler
	payment  *payment.Handler
	shipment *shipment.Handler
	upload   *upload.Handler

	// account membaca peran dan status akun dari DB untuk setiap request ber-token.
	account middleware.Account
}

// newRouter menyusun seluruh route dan middleware global. Route baru
// didaftarkan di sini; rute yang butuh login dibungkus authenticate.
func newRouter(cfg config.Config, tokens *token.Manager, h handlers, rdb *redis.Client) http.Handler {
	verify, revalidate := middleware.Authenticate(tokens), middleware.Revalidate(h.account)
	authenticate := func(next http.Handler) http.Handler { return verify(revalidate(next)) }
	// Semua rate limit memakai sumber IP yang sama (TRUST_PROXY_HEADERS).
	rateLimit := func(limit int, window time.Duration) middleware.Middleware {
		return middleware.RateLimit(limit, window, cfg.TrustProxyHeaders)
	}
	// Limiter per kelompok rute terautentikasi, dihitung per user ID (fallback
	// IP). Satu nilai dipakai bersama oleh semua rute di kelompoknya.
	userWriteLimit := middleware.RateLimitByUser(userWriteRateLimit, userRateWindow, cfg.TrustProxyHeaders)
	userReadLimit := middleware.RateLimitByUser(userReadRateLimit, userRateWindow, cfg.TrustProxyHeaders)
	uploadLimit := middleware.RateLimitByUser(uploadRateLimit, userRateWindow, cfg.TrustProxyHeaders)
	adminLimit := middleware.RateLimitByUser(adminRateLimit, userRateWindow, cfg.TrustProxyHeaders)
	// Belanja (keranjang, checkout, pembayaran) hanya untuk akun pelanggan;
	// admin dan staff ditolak di server, bukan hanya disembunyikan di UI.
	requireCustomer := middleware.RequireRole(user.RoleCustomer)
	// Pengelola katalog, pesanan, dan pembayaran.
	requireBackoffice := middleware.RequireRole(user.RoleOwner, user.RoleStaff)
	optionalAuthenticate := middleware.AuthenticateOptional(tokens)
	cache := middleware.NewResponseCache(rdb, publicCacheTTL)
	cached := cache.Cached()

	// Routing memakai Gin (lihat router.go); pola rute ditulis gaya net/http.
	mux := newGinMux()
	mux.version = cfg.AppVersion
	for _, v := range append([]string{cfg.AppVersion}, cfg.LegacyAPIVersions...) {
		if v != "" {
			mux.aliases = append(mux.aliases, "/api/v"+v)
		}
	}
	mux.HandleFunc("GET "+apiPrefix+"/health", health)
	// Path health persis (kanonik + alias versi), dikecualikan dari MaxConcurrent.
	healthPaths := []string{apiPrefix + "/health"}
	for _, alias := range mux.aliases {
		healthPaths = append(healthPaths, alias+"/health")
	}

	mux.Handle("POST "+apiPrefix+"/auth/register",
		middleware.Chain(http.HandlerFunc(h.auth.Register), rateLimit(authRateLimit, authRateWindow)))
	mux.Handle("POST "+apiPrefix+"/auth/login",
		middleware.Chain(http.HandlerFunc(h.auth.Login), rateLimit(authRateLimit, authRateWindow)))

	mux.Handle("GET "+apiPrefix+"/users/me",
		middleware.Chain(http.HandlerFunc(h.user.Me), authenticate, userReadLimit))
	mux.Handle("PATCH "+apiPrefix+"/users/me",
		middleware.Chain(http.HandlerFunc(h.user.UpdateMe), authenticate, userWriteLimit))
	mux.Handle("PUT "+apiPrefix+"/users/me/password",
		middleware.Chain(http.HandlerFunc(h.user.ChangePassword), authenticate, rateLimit(authRateLimit, authRateWindow)))

	// Publik, tetapi admin dikenali lewat token opsional (lihat include_inactive).
	// GET publik tanpa token dilayani dari cache (lihat middleware.ResponseCache).
	mux.Handle("GET "+apiPrefix+"/categories",
		middleware.Chain(http.HandlerFunc(h.category.List), cached, optionalAuthenticate))
	mux.Handle("GET "+apiPrefix+"/categories/{categoryId}",
		middleware.Chain(http.HandlerFunc(h.category.Get), cached, optionalAuthenticate))
	mux.Handle("POST "+apiPrefix+"/categories",
		middleware.Chain(http.HandlerFunc(h.category.Create), authenticate, requireBackoffice, adminLimit))
	mux.Handle("PATCH "+apiPrefix+"/categories/{categoryId}",
		middleware.Chain(http.HandlerFunc(h.category.Update), authenticate, requireBackoffice, adminLimit))
	mux.Handle("DELETE "+apiPrefix+"/categories/{categoryId}",
		middleware.Chain(http.HandlerFunc(h.category.Delete), authenticate, requireBackoffice, adminLimit))

	// Publik; admin dikenali lewat token opsional (lihat parameter status).
	mux.Handle("GET "+apiPrefix+"/products",
		middleware.Chain(http.HandlerFunc(h.product.List), cached, optionalAuthenticate))
	mux.Handle("GET "+apiPrefix+"/products/{productId}",
		middleware.Chain(http.HandlerFunc(h.product.Get), cached, optionalAuthenticate))
	mux.Handle("POST "+apiPrefix+"/products",
		middleware.Chain(http.HandlerFunc(h.product.Create), authenticate, requireBackoffice, adminLimit))
	mux.Handle("PATCH "+apiPrefix+"/products/{productId}",
		middleware.Chain(http.HandlerFunc(h.product.Update), authenticate, requireBackoffice, adminLimit))
	mux.Handle("DELETE "+apiPrefix+"/products/{productId}",
		middleware.Chain(http.HandlerFunc(h.product.Delete), authenticate, requireBackoffice, adminLimit))

	// Keranjang selalu milik user yang sedang login (lihat token).
	mux.Handle("GET "+apiPrefix+"/cart",
		middleware.Chain(http.HandlerFunc(h.cart.Get), authenticate, requireCustomer, userReadLimit))
	mux.Handle("DELETE "+apiPrefix+"/cart",
		middleware.Chain(http.HandlerFunc(h.cart.Clear), authenticate, requireCustomer, userWriteLimit))
	mux.Handle("POST "+apiPrefix+"/cart/items",
		middleware.Chain(http.HandlerFunc(h.cart.Create), authenticate, requireCustomer, userWriteLimit))
	mux.Handle("PATCH "+apiPrefix+"/cart/items/{cartItemId}",
		middleware.Chain(http.HandlerFunc(h.cart.Update), authenticate, requireCustomer, userWriteLimit))
	mux.Handle("DELETE "+apiPrefix+"/cart/items/{cartItemId}",
		middleware.Chain(http.HandlerFunc(h.cart.Delete), authenticate, requireCustomer, userWriteLimit))

	// Varian dan gambar produk (admin).
	mux.Handle("POST "+apiPrefix+"/products/{productId}/variants",
		middleware.Chain(http.HandlerFunc(h.product.CreateVariant), authenticate, requireBackoffice, adminLimit))
	mux.Handle("PATCH "+apiPrefix+"/products/{productId}/variants/{variantId}",
		middleware.Chain(http.HandlerFunc(h.product.UpdateVariant), authenticate, requireBackoffice, adminLimit))
	mux.Handle("DELETE "+apiPrefix+"/products/{productId}/variants/{variantId}",
		middleware.Chain(http.HandlerFunc(h.product.DeleteVariant), authenticate, requireBackoffice, adminLimit))
	mux.Handle("POST "+apiPrefix+"/products/{productId}/images",
		middleware.Chain(http.HandlerFunc(h.product.CreateImage), authenticate, requireBackoffice, adminLimit))
	mux.Handle("PATCH "+apiPrefix+"/products/{productId}/images/{imageId}",
		middleware.Chain(http.HandlerFunc(h.product.UpdateImage), authenticate, requireBackoffice, adminLimit))
	mux.Handle("DELETE "+apiPrefix+"/products/{productId}/images/{imageId}",
		middleware.Chain(http.HandlerFunc(h.product.DeleteImage), authenticate, requireBackoffice, adminLimit))

	// Unggah gambar ke Cloudinary; URL hasilnya dipakai untuk image_url (admin).
	mux.Handle("POST "+apiPrefix+"/admin/uploads/images",
		middleware.Chain(http.HandlerFunc(h.upload.Image), authenticate, requireBackoffice, uploadLimit))

	// Wilayah berjenjang untuk dropdown alamat (publik, di-cache) dan cek
	// ongkir sampai kelurahan berdasarkan berat keranjang (login).
	// Satu limiter dipakai bersama supaya kuota RajaOngkir tidak habis oleh
	// permintaan ID acak (hasil yang sama tetap dilayani dari cache).
	regionLimit := rateLimit(publicRateLimit, publicRateWindow)
	mux.Handle("GET "+apiPrefix+"/regions/provinces",
		middleware.Chain(http.HandlerFunc(h.ongkir.Provinces), regionLimit))
	mux.Handle("GET "+apiPrefix+"/regions/provinces/{provinceId}/cities",
		middleware.Chain(http.HandlerFunc(h.ongkir.Cities), regionLimit))
	mux.Handle("GET "+apiPrefix+"/regions/cities/{cityId}/districts",
		middleware.Chain(http.HandlerFunc(h.ongkir.Districts), regionLimit))
	mux.Handle("GET "+apiPrefix+"/regions/districts/{districtId}/subdistricts",
		middleware.Chain(http.HandlerFunc(h.ongkir.SubDistricts), regionLimit))
	mux.Handle("GET "+apiPrefix+"/shipping/costs",
		middleware.Chain(http.HandlerFunc(h.ongkir.Costs), authenticate, requireCustomer,
			rateLimit(publicRateLimit, publicRateWindow)))

	// Order: user melihat ordernya sendiri; admin lewat /admin/orders.
	mux.Handle("GET "+apiPrefix+"/orders",
		middleware.Chain(http.HandlerFunc(h.order.ListMine), authenticate, requireCustomer, userReadLimit))
	mux.Handle("POST "+apiPrefix+"/orders",
		middleware.Chain(http.HandlerFunc(h.order.Create), authenticate, requireCustomer, userWriteLimit))
	mux.Handle("POST "+apiPrefix+"/orders/{orderId}/cancel",
		middleware.Chain(http.HandlerFunc(h.order.Cancel), authenticate, requireCustomer, userWriteLimit))
	mux.Handle("POST "+apiPrefix+"/orders/{orderId}/complete",
		middleware.Chain(http.HandlerFunc(h.order.Complete), authenticate, requireCustomer, userWriteLimit))
	mux.Handle("GET "+apiPrefix+"/orders/{orderId}",
		middleware.Chain(http.HandlerFunc(h.order.Get), authenticate, userReadLimit))
	mux.Handle("GET "+apiPrefix+"/admin/orders",
		middleware.Chain(http.HandlerFunc(h.order.AdminList), authenticate, requireBackoffice, adminLimit))
	mux.Handle("PUT "+apiPrefix+"/admin/orders/{orderId}/status",
		middleware.Chain(http.HandlerFunc(h.order.AdminUpdateStatus), authenticate, requireBackoffice, adminLimit))

	// Pembayaran: pemilik order membuat tagihan & unggah bukti; admin memverifikasi.
	mux.Handle("GET "+apiPrefix+"/orders/{orderId}/payment",
		middleware.Chain(http.HandlerFunc(h.payment.Get), authenticate, userReadLimit))
	mux.Handle("POST "+apiPrefix+"/orders/{orderId}/payment",
		middleware.Chain(http.HandlerFunc(h.payment.Create), authenticate, requireCustomer, userWriteLimit))
	mux.Handle("GET "+apiPrefix+"/orders/{orderId}/payment/proof",
		middleware.Chain(http.HandlerFunc(h.payment.GetProof), authenticate, userReadLimit))
	mux.Handle("PUT "+apiPrefix+"/orders/{orderId}/payment/proof",
		middleware.Chain(http.HandlerFunc(h.payment.UploadProof), authenticate, requireCustomer, uploadLimit))
	mux.Handle("GET "+apiPrefix+"/admin/payments",
		middleware.Chain(http.HandlerFunc(h.payment.AdminList), authenticate, requireBackoffice, adminLimit))
	mux.Handle("POST "+apiPrefix+"/admin/payments/{paymentId}/review",
		middleware.Chain(http.HandlerFunc(h.payment.AdminReview), authenticate, requireBackoffice, adminLimit))

	// Pengiriman: dibaca pemilik order/admin, diisi admin.
	mux.Handle("GET "+apiPrefix+"/orders/{orderId}/shipment",
		middleware.Chain(http.HandlerFunc(h.shipment.Get), authenticate, userReadLimit))
	mux.Handle("PUT "+apiPrefix+"/orders/{orderId}/shipment",
		middleware.Chain(http.HandlerFunc(h.shipment.Upsert), authenticate, requireBackoffice, adminLimit))

	// Semua yang tidak cocok dijawab dengan format error yang sama.
	mux.HandleFunc("/", notFound)

	api := middleware.Chain(mux,
		middleware.Recover(),
		// Header keamanan dipasang paling awal supaya ikut di respons error,
		// dan preflight CORS.
		middleware.SecurityHeaders(cfg.TrustProxyHeaders),
		middleware.Gzip(),
		// Batas request bersamaan (MAX_CONCURRENT_REQUESTS); penuh = 503.
		middleware.MaxConcurrent(cfg.MaxConcurrentRequests, healthPaths...),
		middleware.Logger(),
		middleware.CORS(cfg.FrontendURL),
		// Login, profil, dan keranjang tidak mengubah produk/kategori.
		// Checkout & pembatalan (/orders) mengubah stok, jadi tetap mengosongkan cache.
		cache.InvalidateOnWrite(apiPrefix+"/auth/", apiPrefix+"/users/me", apiPrefix+"/cart"),
		middleware.MaxBody(cfg.MaxBodyBytes),
		middleware.Timeout(cfg.HTTPRequestTimeout),
		// Paling dalam supaya pola rute dari router Gin (r.Pattern) terbaca sebagai label.
		middleware.Metrics(),
	)

	// /metrics dibaca Prometheus dan sengaja di luar middleware API (tidak
	// dicatat di log, tidak kena timeout, tidak ikut dihitung sebagai request API).
	root := http.NewServeMux()
	root.Handle("GET /metrics", promhttp.Handler())
	root.Handle("/", api)
	return root
}

func health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotFound, "not_found", "Endpoint tidak ditemukan.")
}
