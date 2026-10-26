package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Settings struct {
	DatabaseURL    string
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	SMTPFrom       string
	AppBaseURL     string
	SessionSecret  []byte
	CookieSecure   bool
	MaxUploadBytes int64
}

var App Settings

func LoadEnv() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(".env dosyası okunamadı: %w", err)
	}

	smtpPort, err := requiredPositiveInt("SMTP_PORT")
	if err != nil {
		return err
	}

	maxUploadMB, err := optionalPositiveInt("MAX_UPLOAD_SIZE_MB", 25)
	if err != nil {
		return err
	}

	cookieSecure, err := strconv.ParseBool(envOrDefault("COOKIE_SECURE", "false"))
	if err != nil {
		return fmt.Errorf("COOKIE_SECURE true veya false olmalı")
	}

	sessionSecret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if len(sessionSecret) < 32 {
		return fmt.Errorf("SESSION_SECRET en az 32 karakter olmalı")
	}

	smtpFrom := envOrDefault("SMTP_FROM", os.Getenv("SMTP_USERNAME"))
	App = Settings{
		DatabaseURL:    requiredEnv("DATABASE_URL"),
		SMTPHost:       requiredEnv("SMTP_HOST"),
		SMTPPort:       smtpPort,
		SMTPUsername:   requiredEnv("SMTP_USERNAME"),
		SMTPPassword:   requiredEnv("SMTP_PASSWORD"),
		SMTPFrom:       smtpFrom,
		AppBaseURL:     strings.TrimRight(envOrDefault("APP_BASE_URL", "http://localhost:8081"), "/"),
		SessionSecret:  []byte(sessionSecret),
		CookieSecure:   cookieSecure,
		MaxUploadBytes: int64(maxUploadMB) * 1024 * 1024,
	}

	if App.DatabaseURL == "" || App.SMTPHost == "" || App.SMTPUsername == "" || App.SMTPPassword == "" || App.SMTPFrom == "" {
		return fmt.Errorf("DATABASE_URL, SMTP_HOST, SMTP_USERNAME, SMTP_PASSWORD ve SMTP_FROM ayarlanmalı")
	}

	return nil
}

func ConnectDB() error {
	db, err := gorm.Open(postgres.Open(App.DatabaseURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("veritabanı bağlantısı başarısız: %w", err)
	}

	DB = db
	log.Println("Veritabanı bağlantısı hazır")
	return nil
}

func requiredEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func requiredPositiveInt(key string) (int, error) {
	value := requiredEnv(key)
	if value == "" {
		return 0, fmt.Errorf("%s ayarlanmalı", key)
	}
	return parsePositiveInt(key, value)
}

func optionalPositiveInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	return parsePositiveInt(key, value)
}

func parsePositiveInt(key, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s pozitif bir sayı olmalı", key)
	}
	return parsed, nil
}
