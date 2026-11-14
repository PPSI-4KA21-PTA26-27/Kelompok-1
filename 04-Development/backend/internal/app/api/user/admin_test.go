package user

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func TestAdminUpdateChanges(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name       string
		body       string
		wantRole   *string
		wantStatus *string
		wantFields []string
	}{
		{name: "role only", body: `{"role":"owner"}`, wantRole: str("owner")},
		{name: "status only", body: `{"status":"inactive"}`, wantStatus: str("inactive")},
		{name: "both, trimmed", body: `{"role":" customer ","status":" active "}`, wantRole: str("customer"), wantStatus: str("active")},
		{name: "empty body", body: `{}`, wantFields: []string{"body"}},
		{name: "null fields are not changes", body: `{"role":null,"status":null}`, wantFields: []string{"body"}},
		{name: "unknown role", body: `{"role":"superadmin"}`, wantRole: str("superadmin"), wantFields: []string{"role"}},
		{name: "role is case sensitive", body: `{"role":"Owner"}`, wantRole: str("Owner"), wantFields: []string{"role"}},
		{name: "unknown status", body: `{"status":"banned"}`, wantStatus: str("banned"), wantFields: []string{"status"}},
		{
			name: "both invalid", body: `{"role":"x","status":"y"}`,
			wantRole: str("x"), wantStatus: str("y"), wantFields: []string{"role", "status"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var req adminUpdateRequest
			if err := json.Unmarshal([]byte(tc.body), &req); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			got, errs := req.changes()

			var gotFields []string
			for _, e := range errs {
				gotFields = append(gotFields, e.Field)
			}
			if !slices.Equal(gotFields, tc.wantFields) {
				t.Errorf("invalid fields = %v, want %v", gotFields, tc.wantFields)
			}
			if !equalPtr(got.Role, tc.wantRole) {
				t.Errorf("Role = %v, want %v", deref(got.Role), deref(tc.wantRole))
			}
			if !equalPtr(got.Status, tc.wantStatus) {
				t.Errorf("Status = %v, want %v", deref(got.Status), deref(tc.wantStatus))
			}
		})
	}
}

func TestListFilterWhere(t *testing.T) {
	tests := []struct {
		name     string
		filter   ListFilter
		wantSQL  string
		wantArgs []any
	}{
		{name: "no filter", filter: ListFilter{}, wantSQL: "", wantArgs: nil},
		{
			name:     "role and status",
			filter:   ListFilter{Role: "owner", Status: "inactive"},
			wantSQL:  " WHERE role_id = (SELECT id FROM roles WHERE code = $1) AND status = $2",
			wantArgs: []any{"owner", "inactive"},
		},
		{
			name:     "search only",
			filter:   ListFilter{Q: "lana"},
			wantSQL:  ` WHERE (name ILIKE $1 ESCAPE '\' OR email ILIKE $1 ESCAPE '\')`,
			wantArgs: []any{"%lana%"},
		},
		{
			name:     "search escapes like wildcards and keeps placeholder numbering",
			filter:   ListFilter{Q: `a_b%`, Role: "owner"},
			wantSQL:  ` WHERE (name ILIKE $1 ESCAPE '\' OR email ILIKE $1 ESCAPE '\') AND role_id = (SELECT id FROM roles WHERE code = $2)`,
			wantArgs: []any{`%a\_b\%%`, "owner"},
		},
		{
			name:     "input is never part of the SQL text",
			filter:   ListFilter{Status: "x'; DROP TABLE users; --"},
			wantSQL:  " WHERE status = $1",
			wantArgs: []any{"x'; DROP TABLE users; --"},
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

// Aturan self-lockout dicek sebelum repository dipanggil, jadi bisa diuji
// tanpa database. Kasus yang lolos aturan (dan butuh database) tidak diuji di sini.
func TestServiceAdminUpdateSelfLockout(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name    string
		changes AdminChanges
	}{
		{"demote self", AdminChanges{Role: str(RoleCustomer)}},
		{"deactivate self", AdminChanges{Status: str(StatusInactive)}},
		{"demote and deactivate self", AdminChanges{Role: str(RoleCustomer), Status: str(StatusInactive)}},
	}
	svc := NewService(nil)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.AdminUpdate(context.Background(), "same-id", "same-id", tc.changes)
			if !errors.Is(err, ErrSelfLockout) {
				t.Errorf("error = %v, want ErrSelfLockout", err)
			}
		})
	}
}
