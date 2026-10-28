package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"hii-stranger1/backend/config"
)

// Alur: AuthService.ForgotPassword -> ResetSender -> SMTPSender -> server SMTP (setelan dari config).
// Memakai paket net/smtp bawaan Go, tanpa dependency baru. Gunakan port STARTTLS (587).

// ResetSender mengirim tautan reset kata sandi. Dibuat interface supaya cara kirim
// (SMTP sekarang, WhatsApp nanti kalau diputuskan) bisa diganti tanpa mengubah AuthService.
type ResetSender interface {
	SendResetLink(ctx context.Context, toEmail, link string) error
}

type SMTPSender struct {
	cfg  config.SMTP
	from mail.Address
}

func NewSMTPSender(cfg config.SMTP) (*SMTPSender, error) {
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("SMTP_FROM tidak valid: %w", err)
	}
	return &SMTPSender{cfg: cfg, from: *from}, nil
}

func (s *SMTPSender) SendResetLink(ctx context.Context, toEmail, link string) error {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(s.cfg.Host, s.cfg.Port))
	if err != nil {
		return fmt.Errorf("hubungi server SMTP: %w", err)
	}
	// Batas waktu total supaya server SMTP yang menggantung tidak menahan request selamanya.
	if err := conn.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		conn.Close()
		return fmt.Errorf("atur batas waktu SMTP: %w", err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("mulai sesi SMTP: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			return fmt.Errorf("STARTTLS: %w", err)
		}
	}
	if s.cfg.User != "" {
		// PlainAuth menolak mengirim sandi lewat koneksi tanpa enkripsi (kecuali localhost).
		if err := client.Auth(smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)); err != nil {
			return fmt.Errorf("login SMTP: %w", err)
		}
	}

	if err := client.Mail(s.from.Address); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	if _, err := w.Write([]byte(s.message(toEmail, link))); err != nil {
		return fmt.Errorf("tulis isi email: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("selesaikan email: %w", err)
	}
	return client.Quit()
}

// message menyusun email teks biasa. toEmail sudah divalidasi (tanpa baris baru),
// jadi aman dipakai di header.
func (s *SMTPSender) message(toEmail, link string) string {
	lines := []string{
		"From: " + s.from.String(),
		"To: " + toEmail,
		"Subject: Reset kata sandi Hii Stranger",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		"Halo,",
		"",
		"Kami menerima permintaan reset kata sandi untuk akun Hii Stranger kamu.",
		"Buka tautan berikut untuk membuat kata sandi baru (berlaku 15 menit):",
		"",
		link,
		"",
		"Kalau kamu tidak meminta ini, abaikan saja email ini.",
		"",
		"Hii Stranger",
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}
