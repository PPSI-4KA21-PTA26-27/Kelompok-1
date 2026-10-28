package config

import "os"

type Config struct {
	Port string
}

// Load membaca konfigurasi dari environment. Koneksi database ditangani paket db.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port}
}
