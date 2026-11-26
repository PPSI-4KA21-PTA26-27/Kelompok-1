package order

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"hii-stranger1/backend/internal/pkg/response"
)

// paidStatuses adalah status pesanan yang dihitung sebagai penjualan.
const paidStatuses = `('paid', 'processing', 'shipped', 'completed')`

// wib dipakai untuk batas hari laporan (toko beroperasi di WIB).
var wib = time.FixedZone("WIB", 7*60*60)

// Analytics adalah laporan penjualan untuk dashboard analitik admin.
type Analytics struct {
	Days        int                `json:"days"`
	Month       string             `json:"month,omitempty"`
	Current     AnalyticsTotals    `json:"current"`
	Previous    AnalyticsTotals    `json:"previous"`
	Series      []AnalyticsPoint   `json:"series"`
	TopProducts []AnalyticsProduct `json:"top_products"`
	Statuses    map[string]int     `json:"statuses"`
}

// AnalyticsTotals merangkum satu periode.
type AnalyticsTotals struct {
	Revenue    string `json:"revenue"`
	PaidOrders int    `json:"paid_orders"`
	AllOrders  int    `json:"all_orders"`
	Cancelled  int    `json:"cancelled"`
	Buyers     int    `json:"buyers"`
	ItemsSold  int    `json:"items_sold"`
}

// AnalyticsPoint adalah penjualan per hari (tanggal WIB, YYYY-MM-DD).
type AnalyticsPoint struct {
	Date    string `json:"date"`
	Revenue string `json:"revenue"`
	Orders  int    `json:"orders"`
}

// AnalyticsProduct adalah produk terlaris pada periode berjalan.
type AnalyticsProduct struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	Revenue   string `json:"revenue"`
}

// Analytics menghitung laporan untuk [start, end) dan periode sebelumnya
// dengan panjang sama [prevStart, start).
func (r *Repository) Analytics(ctx context.Context, prevStart, start, end time.Time) (Analytics, error) {
	a := Analytics{Statuses: map[string]int{}}
	b := &pgx.Batch{}
	totals := `SELECT COALESCE(SUM(o.total_amount) FILTER (WHERE o.order_status IN ` + paidStatuses + `), 0)::text,
	              COUNT(*) FILTER (WHERE o.order_status IN ` + paidStatuses + `),
	              COUNT(*),
	              COUNT(*) FILTER (WHERE o.order_status = 'cancelled'),
	              COUNT(DISTINCT o.user_id) FILTER (WHERE o.order_status IN ` + paidStatuses + `),
	              COALESCE((SELECT SUM(oi.quantity) FROM order_items oi JOIN orders o2 ON o2.id = oi.order_id
	                         WHERE o2.created_at >= $1 AND o2.created_at < $2 AND o2.order_status IN ` + paidStatuses + `), 0)
	         FROM orders o WHERE o.created_at >= $1 AND o.created_at < $2`
	b.Queue(totals, start, end)
	b.Queue(totals, prevStart, start)
	b.Queue(`SELECT d, COALESCE(SUM(o.total_amount), 0)::text, COUNT(o.id)
	           FROM generate_series($1::timestamptz, $2::timestamptz - interval '1 day', interval '1 day') d
	           LEFT JOIN orders o ON o.created_at >= d AND o.created_at < d + interval '1 day'
	                AND o.order_status IN `+paidStatuses+`
	          GROUP BY d ORDER BY d`, start, end)
	b.Queue(`SELECT COALESCE(oi.product_id::text, ''), oi.product_name, SUM(oi.quantity)::int, SUM(oi.subtotal)::text
	           FROM order_items oi JOIN orders o ON o.id = oi.order_id
	          WHERE o.created_at >= $1 AND o.created_at < $2 AND o.order_status IN `+paidStatuses+`
	          GROUP BY oi.product_id, oi.product_name
	          ORDER BY SUM(oi.subtotal) DESC, oi.product_name LIMIT 50`, start, end)
	b.Queue(`SELECT order_status, COUNT(*) FROM orders WHERE created_at >= $1 AND created_at < $2 GROUP BY order_status`, start, end)
	br := r.db.SendBatch(ctx, b)
	defer func() { _ = br.Close() }()

	for _, t := range []*AnalyticsTotals{&a.Current, &a.Previous} {
		if err := br.QueryRow().Scan(&t.Revenue, &t.PaidOrders, &t.AllOrders, &t.Cancelled, &t.Buyers, &t.ItemsSold); err != nil {
			return Analytics{}, wrap("analytics totals", err)
		}
	}
	rows, err := br.Query()
	if err != nil {
		return Analytics{}, wrap("analytics series", err)
	}
	a.Series, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (AnalyticsPoint, error) {
		var p AnalyticsPoint
		var d time.Time
		err := row.Scan(&d, &p.Revenue, &p.Orders)
		p.Date = d.In(wib).Format("2006-01-02")
		return p, err
	})
	if err != nil {
		return Analytics{}, wrap("analytics series scan", err)
	}
	rows, err = br.Query()
	if err != nil {
		return Analytics{}, wrap("analytics top", err)
	}
	a.TopProducts, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (AnalyticsProduct, error) {
		var p AnalyticsProduct
		err := row.Scan(&p.ProductID, &p.Name, &p.Quantity, &p.Revenue)
		return p, err
	})
	if err != nil {
		return Analytics{}, wrap("analytics top scan", err)
	}
	if a.TopProducts == nil {
		a.TopProducts = []AnalyticsProduct{}
	}
	rows, err = br.Query()
	if err != nil {
		return Analytics{}, wrap("analytics status", err)
	}
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			rows.Close()
			return Analytics{}, wrap("analytics status scan", err)
		}
		a.Statuses[st] = n
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Analytics{}, wrap("analytics status rows", err)
	}
	return a, nil
}

// analyticsRange mengembalikan batas periode N hari terakhir (termasuk hari
// ini, WIB) beserta periode pembanding sebelumnya.
func analyticsRange(now time.Time, days int) (prevStart, start, end time.Time) {
	n := now.In(wib)
	end = time.Date(n.Year(), n.Month(), n.Day()+1, 0, 0, 0, 0, wib)
	start = end.AddDate(0, 0, -days)
	prevStart = start.AddDate(0, 0, -days)
	return prevStart, start, end
}

// monthRange mengembalikan batas satu bulan kalender WIB beserta bulan
// sebelumnya sebagai pembanding (layar L-29 memilih satu bulan).
func monthRange(month time.Time) (prevStart, start, end time.Time) {
	start = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, wib)
	return start.AddDate(0, -1, 0), start, start.AddDate(0, 1, 0)
}

// Analytics mengembalikan laporan penjualan N hari terakhir.
func (s *Service) Analytics(ctx context.Context, days int) (Analytics, error) {
	prevStart, start, end := analyticsRange(time.Now(), days)
	a, err := s.repo.Analytics(ctx, prevStart, start, end)
	a.Days = days
	return a, err
}

// AnalyticsMonth mengembalikan laporan penjualan satu bulan kalender (WIB).
func (s *Service) AnalyticsMonth(ctx context.Context, month time.Time) (Analytics, error) {
	prevStart, start, end := monthRange(month)
	a, err := s.repo.Analytics(ctx, prevStart, start, end)
	a.Days = int(end.Sub(start).Hours() / 24)
	a.Month = start.Format("2006-01")
	return a, err
}

// AdminAnalytics melayani GET /admin/analytics?days=7|30|90 atau
// ?month=YYYY-MM (khusus owner, UC-19).
func (h *Handler) AdminAnalytics(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var (
		a   Analytics
		err error
	)
	if m := q.Get("month"); m != "" {
		month, perr := time.ParseInLocation("2006-01", m, wib)
		if perr != nil || month.After(time.Now().In(wib)) {
			response.Error(w, http.StatusBadRequest, "bad_request", "Parameter month harus berformat YYYY-MM dan tidak boleh bulan mendatang.")
			return
		}
		a, err = h.svc.AnalyticsMonth(r.Context(), month)
	} else {
		days := 30
		switch q.Get("days") {
		case "", "30":
		case "7":
			days = 7
		case "90":
			days = 90
		default:
			response.Error(w, http.StatusBadRequest, "bad_request", "Parameter days harus 7, 30, atau 90.")
			return
		}
		a, err = h.svc.Analytics(r.Context(), days)
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "order analytics failed", "error", err)
		response.Internal(w)
		return
	}
	response.JSON(w, http.StatusOK, a)
}
