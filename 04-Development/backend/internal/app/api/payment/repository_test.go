package payment

import (
	"slices"
	"testing"
)

func TestFilterWhere(t *testing.T) {
	tests := []struct {
		name     string
		filter   Filter
		wantSQL  string
		wantArgs []any
	}{
		{name: "no filter", filter: Filter{}, wantSQL: "", wantArgs: nil},
		{
			name:     "status and order without search is unchanged",
			filter:   Filter{Status: "menunggu", OrderID: "o-1"},
			wantSQL:  " WHERE p.status = $1 AND p.order_id = $2",
			wantArgs: []any{"menunggu", "o-1"},
		},
		{
			name:     "search only",
			filter:   Filter{Q: "lana"},
			wantSQL:  " WHERE " + `(o.order_number ILIKE $1 ESCAPE '\' OR p.id::text ILIKE $1 ESCAPE '\' OR u.name ILIKE $1 ESCAPE '\' OR u.email ILIKE $1 ESCAPE '\')`,
			wantArgs: []any{"%lana%"},
		},
		{
			name:     "search is ANDed with status and escapes like wildcards",
			filter:   Filter{Status: "paid", Q: `a_b%\`},
			wantSQL:  " WHERE p.status = $1 AND " + `(o.order_number ILIKE $2 ESCAPE '\' OR p.id::text ILIKE $2 ESCAPE '\' OR u.name ILIKE $2 ESCAPE '\' OR u.email ILIKE $2 ESCAPE '\')`,
			wantArgs: []any{"paid", `%a\_b\%\\%`},
		},
		{
			name:     "input is never part of the SQL text",
			filter:   Filter{Q: "x'; DROP TABLE payments; --"},
			wantSQL:  " WHERE " + `(o.order_number ILIKE $1 ESCAPE '\' OR p.id::text ILIKE $1 ESCAPE '\' OR u.name ILIKE $1 ESCAPE '\' OR u.email ILIKE $1 ESCAPE '\')`,
			wantArgs: []any{"%x'; DROP TABLE payments; --%"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSQL, gotArgs := tc.filter.where()
			if gotSQL != tc.wantSQL {
				t.Errorf("SQL = %q, want %q", gotSQL, tc.wantSQL)
			}
			if !slices.Equal(gotArgs, tc.wantArgs) {
				t.Errorf("args = %v, want %v", gotArgs, tc.wantArgs)
			}
		})
	}
}
