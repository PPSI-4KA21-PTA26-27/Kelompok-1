package mailer

import (
	"html"
	"strings"
)

// Link adalah tautan yang ditampilkan di email (tombol atau tautan biasa).
type Link struct {
	Label string
	URL   string
}

// Email adalah isi email transaksional toko. Render menghasilkan versi teks
// biasa dan HTML dengan isi yang sama, jadi klien email tanpa HTML tetap
// menampilkan semua informasi (termasuk tautan).
type Email struct {
	Heading    string
	Paragraphs []string
	// Details ditampilkan sebagai tabel label: nilai (mis. no. order, total).
	Details [][2]string
	// Button adalah aksi utama (mis. "Lihat pesanan").
	Button *Link
	// Links adalah tautan tambahan di bawah tombol (mis. lacak di situs kurir).
	Links []Link
	Note  string
}

const brand = "HII STRANGER"

// Render mengembalikan isi email versi teks dan HTML.
func (e Email) Render() (text, htmlBody string) {
	var t strings.Builder
	t.WriteString(e.Heading + "\n\n")
	for _, p := range e.Paragraphs {
		t.WriteString(p + "\n\n")
	}
	for _, d := range e.Details {
		t.WriteString(d[0] + ": " + d[1] + "\n")
	}
	if len(e.Details) > 0 {
		t.WriteString("\n")
	}
	if e.Button != nil {
		t.WriteString(e.Button.Label + ": " + e.Button.URL + "\n")
	}
	for _, l := range e.Links {
		t.WriteString(l.Label + ": " + l.URL + "\n")
	}
	if e.Note != "" {
		t.WriteString("\n" + e.Note + "\n")
	}
	t.WriteString("\n— " + brand)

	esc := html.EscapeString
	var h strings.Builder
	h.WriteString(`<!doctype html><html><body style="margin:0;background:#fbf9f5;font-family:Arial,Helvetica,sans-serif;color:#141413">`)
	h.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#fbf9f5;padding:24px 12px"><tr><td align="center">`)
	h.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:560px;background:#ffffff;border:1px solid #e6e2d8;border-radius:16px">`)
	h.WriteString(`<tr><td style="padding:20px 28px;border-bottom:1px solid #e6e2d8;font-weight:800;letter-spacing:.08em;font-size:14px">` + brand + `</td></tr>`)
	h.WriteString(`<tr><td style="padding:28px">`)
	h.WriteString(`<h1 style="margin:0 0 16px;font-size:22px;line-height:28px;font-weight:800;text-transform:uppercase;letter-spacing:-.01em">` + esc(e.Heading) + `</h1>`)
	for _, p := range e.Paragraphs {
		h.WriteString(`<p style="margin:0 0 14px;font-size:15px;line-height:22px;color:#3d3b37">` + esc(p) + `</p>`)
	}
	if len(e.Details) > 0 {
		h.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin:8px 0 20px;background:#f5f2eb;border:1px solid #e6e2d8;border-radius:12px">`)
		for _, d := range e.Details {
			h.WriteString(`<tr><td style="padding:10px 16px;font-size:12px;color:#6e6b64;text-transform:uppercase;letter-spacing:.05em">` + esc(d[0]) +
				`</td><td align="right" style="padding:10px 16px;font-size:14px;font-weight:700">` + esc(d[1]) + `</td></tr>`)
		}
		h.WriteString(`</table>`)
	}
	if e.Button != nil {
		h.WriteString(`<p style="margin:8px 0 20px"><a href="` + esc(e.Button.URL) + `" style="display:inline-block;background:#d4ff14;color:#141413;text-decoration:none;font-weight:700;font-size:14px;padding:14px 22px;border-radius:12px">` + esc(e.Button.Label) + `</a></p>`)
	}
	for _, l := range e.Links {
		h.WriteString(`<p style="margin:0 0 8px;font-size:14px"><a href="` + esc(l.URL) + `" style="color:#141413;font-weight:700">` + esc(l.Label) + ` &rarr;</a></p>`)
	}
	if e.Note != "" {
		h.WriteString(`<p style="margin:20px 0 0;font-size:12px;line-height:18px;color:#6e6b64">` + esc(e.Note) + `</p>`)
	}
	h.WriteString(`</td></tr></table></td></tr></table></body></html>`)
	return t.String(), h.String()
}
