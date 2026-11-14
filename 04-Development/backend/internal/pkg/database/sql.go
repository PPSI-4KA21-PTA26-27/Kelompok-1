package database

import "strings"

// EscapeLike meng-escape \, % dan _ agar aman dipakai dalam pola LIKE.
func EscapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// NullIfEmpty mengubah string kosong menjadi NULL.
func NullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
