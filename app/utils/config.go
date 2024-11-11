package utils

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	DatabaseName     string `env:"DATABASE_NAME"`
	DatabaseHost     string `env:"DATABASE_HOST"`
	DatabasePort     int    `env:"DATABASE_PORT"`
	DatabaseUser     string `env:"DATABASE_USER"`
	DatabasePassword string `env:"DATABASE_PASSWORD"`

	AppVersion string `env:"APP_VERSION"`
	AppPort    string `env:"APP_PORT"`

	SessionSecret      string        `env:"SESSION_SECRET"`
	SessionExpire      time.Duration `env:"SESSION_EXPIRE"`
	RefreshExpire      time.Duration `env:"REFRESH_EXPIRE"`
	VerificationExpire time.Duration `env:"VERIFICATION_EXPIRE"`

	SmtpHost     string `env:"SMTP_HOST"`
	SmtpPort     int    `env:"SMTP_PORT"`
	SmtpUser     string `env:"SMTP_USER"`
	SmtpPassword string `env:"SMTP_PASSWORD"`
	SmtpSender   string `env:"SMTP_SENDER"`
}

// LoadConfig загружает конфигурацию из .env и парсит длительности
func (config *AppConfig) LoadConfig() error {
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	viper.BindEnv("DatabaseName", "DATABASE_NAME")
	viper.BindEnv("DatabaseHost", "DATABASE_HOST")
	viper.BindEnv("DatabasePort", "DATABASE_PORT")
	viper.BindEnv("DatabaseUser", "DATABASE_USER")
	viper.BindEnv("DatabasePassword", "DATABASE_PASSWORD")

	viper.BindEnv("AppVersion", "APP_VERSION")
	viper.BindEnv("AppPort", "APP_PORT")

	viper.BindEnv("SessionSecret", "SESSION_SECRET")
	viper.BindEnv("SessionExpire", "SESSION_EXPIRE")
	viper.BindEnv("RefreshExpire", "REFRESH_EXPIRE")
	viper.BindEnv("VerificationExpire", "VERIFICATION_EXPIRE")

	viper.BindEnv("SmtpHost", "SMTP_HOST")
	viper.BindEnv("SmtpPort", "SMTP_PORT")
	viper.BindEnv("SmtpUser", "SMTP_USER")
	viper.BindEnv("SmtpPassword", "SMTP_PASSWORD")
	viper.BindEnv("SmtpSender", "SMTP_SENDER")

	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("unable to decode into struct: %w", err)
	}

	return nil
}
