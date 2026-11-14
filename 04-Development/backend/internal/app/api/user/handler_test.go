package user

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestOptionalStringUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantSet   bool
		wantValue string // kosong berarti Value harus nil
		wantErr   bool
	}{
		{name: "field absent", body: `{}`, wantSet: false},
		{name: "explicit null", body: `{"phone":null}`, wantSet: true},
		{name: "text value", body: `{"phone":"0812"}`, wantSet: true, wantValue: "0812"},
		{name: "wrong type", body: `{"phone":123}`, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var req updateProfileRequest
			err := json.Unmarshal([]byte(tc.body), &req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if req.Phone.Set != tc.wantSet {
				t.Errorf("Set = %v, want %v", req.Phone.Set, tc.wantSet)
			}
			switch {
			case tc.wantValue == "" && req.Phone.Value != nil:
				t.Errorf("Value = %q, want nil", *req.Phone.Value)
			case tc.wantValue != "" && (req.Phone.Value == nil || *req.Phone.Value != tc.wantValue):
				t.Errorf("Value = %v, want %q", req.Phone.Value, tc.wantValue)
			}
		})
	}
}

func TestUpdateProfileChanges(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name       string
		body       string
		wantName   *string
		wantPhone  OptionalString
		wantFields []string
	}{
		{
			name:     "name only",
			body:     `{"name":"  Lana  "}`,
			wantName: str("Lana"),
		},
		{
			name:      "phone only",
			body:      `{"phone":" +62 812-3456 "}`,
			wantPhone: OptionalString{Set: true, Value: str("+62 812-3456")},
		},
		{
			name:      "null phone clears number",
			body:      `{"phone":null}`,
			wantPhone: OptionalString{Set: true},
		},
		{
			name:      "blank phone clears number",
			body:      `{"phone":"   "}`,
			wantPhone: OptionalString{Set: true},
		},
		{
			name:       "empty body",
			body:       `{}`,
			wantFields: []string{"body"},
		},
		{
			name:       "null name is not a change",
			body:       `{"name":null}`,
			wantFields: []string{"body"},
		},
		{
			name:       "blank name",
			body:       `{"name":"   "}`,
			wantName:   str(""),
			wantFields: []string{"name"},
		},
		{
			name:       "name too long",
			body:       `{"name":"` + strings.Repeat("a", 101) + `"}`,
			wantName:   str(strings.Repeat("a", 101)),
			wantFields: []string{"name"},
		},
		{
			name:       "invalid phone",
			body:       `{"phone":"telepon"}`,
			wantPhone:  OptionalString{Set: true, Value: str("telepon")},
			wantFields: []string{"phone"},
		},
		{
			name:     "email role and status are ignored",
			body:     `{"name":"Lana","email":"x@example.com","role":"owner","status":"inactive"}`,
			wantName: str("Lana"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var req updateProfileRequest
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

			if !equalPtr(got.Name, tc.wantName) {
				t.Errorf("Name = %v, want %v", deref(got.Name), deref(tc.wantName))
			}
			if got.Phone.Set != tc.wantPhone.Set || !equalPtr(got.Phone.Value, tc.wantPhone.Value) {
				t.Errorf("Phone = {Set:%v Value:%v}, want {Set:%v Value:%v}",
					got.Phone.Set, deref(got.Phone.Value), tc.wantPhone.Set, deref(tc.wantPhone.Value))
			}
		})
	}
}

func TestValidateChangePassword(t *testing.T) {
	tests := []struct {
		name       string
		req        changePasswordRequest
		wantFields []string
	}{
		{"valid", changePasswordRequest{CurrentPassword: "lama", NewPassword: "password123"}, nil},
		{"empty", changePasswordRequest{}, []string{"current_password", "new_password"}},
		{"new password too short", changePasswordRequest{CurrentPassword: "lama", NewPassword: "pendek"}, []string{"new_password"}},
		{"new password over bcrypt limit", changePasswordRequest{CurrentPassword: "lama", NewPassword: strings.Repeat("a", 73)}, []string{"new_password"}},
		{"missing current password", changePasswordRequest{NewPassword: "password123"}, []string{"current_password"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, e := range validateChangePassword(tc.req) {
				got = append(got, e.Field)
			}
			if !slices.Equal(got, tc.wantFields) {
				t.Errorf("invalid fields = %v, want %v", got, tc.wantFields)
			}
		})
	}
}

func equalPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func deref(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}
