package mailer

import (
	"strings"
	"testing"
)

func TestEmailRender(t *testing.T) {
	e := Email{
		Heading:    "Pesanan <dikirim>",
		Paragraphs: []string{"Halo & selamat"},
		Details:    [][2]string{{"No. resi", "P123"}},
		Button:     &Link{Label: "Lacak", URL: "https://toko.test/orders/1?a=1&b=2"},
		Links:      []Link{{Label: "cekresi", URL: "https://cekresi.com/?noresi=P123"}},
	}
	text, html := e.Render()
	for _, want := range []string{"No. resi: P123", "Lacak: https://toko.test/orders/1?a=1&b=2", "cekresi: https://cekresi.com/?noresi=P123"} {
		if !strings.Contains(text, want) {
			t.Errorf("text missing %q", want)
		}
	}
	if strings.Contains(html, "<dikirim>") || !strings.Contains(html, "Pesanan &lt;dikirim&gt;") {
		t.Error("heading not escaped in HTML")
	}
	if !strings.Contains(html, `href="https://toko.test/orders/1?a=1&amp;b=2"`) {
		t.Error("button URL not escaped in HTML")
	}
}

func TestBuildMessageMultipart(t *testing.T) {
	msg := string(buildMessage("a@x.test", "b@x.test", "Halo", "teks", "<p>html</p>", nil))
	if !strings.Contains(msg, "multipart/alternative") || !strings.Contains(msg, "text/html") {
		t.Fatalf("multipart parts missing:\n%s", msg)
	}
	plain := string(buildMessage("a@x.test", "b@x.test", "Halo", "teks", "", nil))
	if strings.Contains(plain, "multipart") {
		t.Error("plain message should not be multipart")
	}
}

func TestBuildMessageAttachment(t *testing.T) {
	msg := string(buildMessage("a@x.test", "b@x.test", "Halo", "teks", "<p>html</p>",
		&Attachment{Name: "bukti.png", Type: "image/png", Data: []byte("abc")}))
	for _, want := range []string{"multipart/mixed", "multipart/alternative", `filename="bukti.png"`, "YWJj"} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestAppURL(t *testing.T) {
	m := New("", "587", "", "", "", "", "http://localhost:3000/")
	if got := m.AppURL("/orders/1"); got != "http://localhost:3000/orders/1" {
		t.Errorf("AppURL = %q", got)
	}
	var nilMailer *Mailer
	if nilMailer.AppURL("/x") != "" || nilMailer.Enabled() {
		t.Error("nil mailer must be safe")
	}
}
