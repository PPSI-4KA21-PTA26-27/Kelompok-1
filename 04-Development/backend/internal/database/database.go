package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kel1/backend/internal/config"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password,
		cfg.DB.Name, cfg.DB.SSLMode, cfg.DB.TimeZone,
	)

	logLevel := logger.Warn
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
}
