package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"hii-stranger1/backend/config"
	"hii-stranger1/backend/db"
	"hii-stranger1/backend/handlers"
	"hii-stranger1/backend/middleware"
	"hii-stranger1/backend/repository"
	"hii-stranger1/backend/services"
)

// Alur penyambungan: db -> repository -> services -> handlers -> route.
// config (JWT, SMTP) dibagikan ke services dan middleware.Auth.
// Cara membacanya: baca dari atas ke bawah, tiap langkah memakai hasil langkah sebelumnya.

// registerAuthRoutes mendaftarkan route daftar/masuk/lupa sandi dan akun saya.
func registerAuthRoutes(mux *http.ServeMux, cfg config.Config) error {
	// 1. repository: akses database
	users := repository.NewUserRepository(db.DB)
	resets := repository.NewResetTokenRepository(db.DB)

	// 2. services: aturan bisnis. Pengirim email hanya aktif kalau SMTP sudah diatur di .env.
	var sender services.ResetSender
	if cfg.SMTP.Configured() {
		smtpSender, err := services.NewSMTPSender(cfg.SMTP)
		if err != nil {
			return err
		}
		sender = smtpSender
	} else {
		log.Println("SMTP belum diatur (SMTP_HOST, SMTP_PORT, SMTP_FROM): email reset kata sandi tidak akan terkirim")
	}
	auth, err := services.NewAuthService(users, resets, cfg, sender)
	if err != nil {
		return fmt.Errorf("siapkan auth service: %w", err)
	}

	// 3. handlers: urusan HTTP
	authHandler := handlers.NewAuthHandler(auth)
	accountHandler := handlers.NewAccountHandler(auth)

	// 4. middleware per route. Batas percobaan dihitung per IP, tiap route punya hitungan sendiri.
	requireLogin := middleware.Auth(cfg.JWTSecret)
	registerLimit := middleware.RateLimit(5, time.Minute)
	loginLimit := middleware.RateLimit(10, time.Minute)
	forgotLimit := middleware.RateLimit(5, time.Minute)
	resetLimit := middleware.RateLimit(10, time.Minute)

	// Publik (tanpa token), dibatasi rate limit
	mux.Handle("POST /api/v1/auth/register", middleware.Chain(http.HandlerFunc(authHandler.Register), registerLimit))
	mux.Handle("POST /api/v1/auth/login", middleware.Chain(http.HandlerFunc(authHandler.Login), loginLimit))
	mux.Handle("POST /api/v1/auth/forgot-password", middleware.Chain(http.HandlerFunc(authHandler.ForgotPassword), forgotLimit))
	mux.Handle("POST /api/v1/auth/reset-password", middleware.Chain(http.HandlerFunc(authHandler.ResetPassword), resetLimit))

	// Butuh login (token Bearer)
	mux.Handle("GET /api/v1/me", middleware.Chain(http.HandlerFunc(accountHandler.Me), requireLogin))
	mux.Handle("PUT /api/v1/me", middleware.Chain(http.HandlerFunc(accountHandler.UpdateMe), requireLogin))
	mux.Handle("PUT /api/v1/me/password", middleware.Chain(http.HandlerFunc(accountHandler.ChangePassword), requireLogin))

	return nil
}
