// Package mailer mengirim email notifikasi lewat SMTP (misalnya Gmail).
//
// Untuk Gmail pakai SMTP_HOST=smtp.gmail.com, SMTP_PORT=587, dan App Password
// (bukan password akun Google biasa) di SMTP_PASS.
package mailer

import (
	"context"
	"encoding/base64"
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// sendTimeout adalah batas waktu satu kali kirim email di background.
const sendTimeout = 30 * time.Second

// Mailer menyimpan pengaturan SMTP. Mailer dengan host kosong dianggap
// nonaktif: semua pengiriman dilewati tanpa error (berguna saat development).
type Mailer struct {
	host  string
	port  string
	user  string
	pass  string
	from  string
	owner string
	app   string
}

// New membuat Mailer. from boleh berbentuk "Nama Toko <alamat@gmail.com>";
// bila kosong, user dipakai sebagai pengirim. owner adalah email pemilik toko
// yang menerima notifikasi order dan pembayaran baru.
func New(host, port, user, pass, from, owner, appURL string) *Mailer {
	if from == "" {
		from = user
	}
	return &Mailer{host: host, port: port, user: user, pass: pass, from: from, owner: owner,
		app: strings.TrimRight(appURL, "/")}
}

// AppURL menggabungkan alamat frontend (FRONTEND_URL) dengan path, untuk
// tautan di email (mis. /orders/{id}). Kosong bila FRONTEND_URL tidak diisi.
func (m *Mailer) AppURL(path string) string {
	if m == nil || m.app == "" {
		return ""
	}
	return m.app + path
}

// Enabled bernilai true bila SMTP sudah dikonfigurasi.
func (m *Mailer) Enabled() bool {
	return m != nil && m.host != ""
}

// NotifyEmail mengirim Email (versi teks + HTML) di background supaya response
// API tidak ikut menunggu server SMTP. Kegagalan hanya dicatat ke log karena
// email bersifat tambahan.
func (m *Mailer) NotifyEmail(to, subject string, e Email) {
	text, html := e.Render()
	m.notify(to, subject, text, html, e.Attachment)
}

// NotifyOwnerEmail sama dengan NotifyEmail, dengan penerima email pemilik toko.
func (m *Mailer) NotifyOwnerEmail(subject string, e Email) {
	if m == nil {
		return
	}
	m.NotifyEmail(m.owner, subject, e)
}

func (m *Mailer) notify(to, subject, text, html string, att *Attachment) {
	if !m.Enabled() || to == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
		defer cancel()
		if err := m.send(ctx, to, subject, text, html, att); err != nil {
			slog.Error("mailer: send failed", "subject", subject, "error", err)
		}
	}()
}

// send mengirim satu email dan menunggu sampai selesai. Port 465 memakai TLS
// langsung; port lain (mis. 587) memakai STARTTLS.
func (m *Mailer) send(ctx context.Context, to, subject, body, htmlBody string, att *Attachment) error {
	if !m.Enabled() {
		return nil
	}
	fromAddr, err := mail.ParseAddress(m.from)
	if err != nil {
		return fmt.Errorf("mailer: parse from: %w", err)
	}
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("mailer: parse to: %w", err)
	}

	conn, err := m.dial(ctx)
	if err != nil {
		return err
	}
	// Seluruh percakapan SMTP ikut dibatasi waktu dari ctx.
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	c, err := smtp.NewClient(conn, m.host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("mailer: smtp client: %w", err)
	}
	defer func() { _ = c.Close() }()

	if m.port != "465" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
				return fmt.Errorf("mailer: starttls: %w", err)
			}
		}
	}
	if m.user != "" {
		if err := c.Auth(smtp.PlainAuth("", m.user, m.pass, m.host)); err != nil {
			return fmt.Errorf("mailer: auth: %w", err)
		}
	}
	if err := c.Mail(fromAddr.Address); err != nil {
		return fmt.Errorf("mailer: mail from: %w", err)
	}
	if err := c.Rcpt(toAddr.Address); err != nil {
		return fmt.Errorf("mailer: rcpt to: %w", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("mailer: data: %w", err)
	}
	if _, err := w.Write(buildMessage(fromAddr.String(), toAddr.String(), subject, body, htmlBody, att)); err != nil {
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}
	return c.Quit()
}

// dial membuka koneksi ke server SMTP, langsung TLS bila port 465.
func (m *Mailer) dial(ctx context.Context) (net.Conn, error) {
	addr := net.JoinHostPort(m.host, m.port)
	if m.port == "465" {
		d := tls.Dialer{Config: &tls.Config{ServerName: m.host}}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("mailer: dial tls: %w", err)
		}
		return conn, nil
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("mailer: dial: %w", err)
	}
	return conn, nil
}

// buildMessage menyusun header dan isi email. Subject di-encode supaya huruf
// non-ASCII aman, dan baris baru dibuang agar header tidak bisa disisipi.
func buildMessage(from, to, subject, body, htmlBody string, att *Attachment) []byte {
	subject = strings.NewReplacer("\r", " ", "\n", " ").Replace(subject)
	crlf := func(v string) string {
		return strings.ReplaceAll(strings.ReplaceAll(v, "\r\n", "\n"), "\n", "\r\n")
	}

	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	if htmlBody == "" {
		b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		b.WriteString("\r\n")
		b.WriteString(crlf(body))
		b.WriteString("\r\n")
		return []byte(b.String())
	}

	// multipart/alternative: klien email memilih versi HTML bila didukung.
	boundary := "hs-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	mixed := boundary + "-m"
	if att != nil {
		b.WriteString("Content-Type: multipart/mixed; boundary=\"" + mixed + "\"\r\n\r\n--" + mixed + "\r\n")
	}
	b.WriteString("Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n")
	b.WriteString("\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(crlf(body) + "\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(crlf(htmlBody) + "\r\n")
	b.WriteString("--" + boundary + "--\r\n")
	if att != nil {
		name := mime.QEncoding.Encode("utf-8", strings.NewReplacer("\r", "", "\n", "", "\"", "").Replace(att.Name))
		b.WriteString("--" + mixed + "\r\n")
		b.WriteString("Content-Type: " + att.Type + "; name=\"" + name + "\"\r\n")
		b.WriteString("Content-Transfer-Encoding: base64\r\n")
		b.WriteString("Content-Disposition: attachment; filename=\"" + name + "\"\r\n\r\n")
		enc := base64.StdEncoding.EncodeToString(att.Data)
		for len(enc) > 76 {
			b.WriteString(enc[:76] + "\r\n")
			enc = enc[76:]
		}
		b.WriteString(enc + "\r\n--" + mixed + "--\r\n")
	}
	return []byte(b.String())
}
