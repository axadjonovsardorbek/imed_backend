package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	AUTH_PORT string
	GIN_MODE  string

	POSTGRES_HOST     string
	POSTGRES_PORT     int
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_NAME     string

	REDIS_HOST string
	REDIS_PORT string

	NEW_BOT_TOKEN     string
	BOT_TOKEN         string
	GROUP_ID          string
	GROUP_FOR_INFO_ID string

	BASE_URL               string
	CLICK_MERCHANT_ID      int
	CLICK_SERVICE_ID       int
	CLICK_SECRET_KEY       string
	PAYMENT_PROVIDER_TOKEN string
	PAYMENT_CURRENCY       string
	PAYMENT_BOT_TOKEN      string

	FIREBASE_CREDENTIALS string
	AUTO_CONTENT_PUSH    bool
	// TARIFF_SALES_CLOSED true bo'lsa eski tariflar bo'yicha yangi to'lov va tarif yaratish yopiladi (T-11).
	TARIFF_SALES_CLOSED bool

	MINIO_ENDPOINT    string
	MINIO_PORT        string
	MINIO_ACCESS_KEY  string
	MINIO_SECRET_KEY  string
	MINIO_BUCKET_NAME string
	MINIO_DOMAIN      string

	CONFIRM_PAGE_URL string

	EMAIL_NAME string
	EMAIL_LOGO string

	BREVO_SERVER string
	BREVO_PORT   string
	BREVO_LOGIN  string
	BREVO_KEY    string

	S3_ACCESS_KEY  string
	S3_SECRET_KEY  string
	S3_ENDPOINT    string
	S3_BUCKET_NAME string
	S3_DOMAIN      string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, falling back to environment and defaults")
	}

	config := Config{}

	config.AUTH_PORT = cast.ToString(coalesce("AUTH_PORT", ":1001"))
	config.GIN_MODE = cast.ToString(coalesce("GIN_MODE", "release"))
	config.AUTO_CONTENT_PUSH = cast.ToBool(coalesce("AUTO_CONTENT_PUSH", false))
	config.TARIFF_SALES_CLOSED = cast.ToBool(coalesce("TARIFF_SALES_CLOSED", false))

	config.POSTGRES_HOST = cast.ToString(coalesce("POSTGRES_HOST", "postgres"))
	config.POSTGRES_PORT = cast.ToInt(coalesce("POSTGRES_PORT", 5432))
	config.POSTGRES_USER = cast.ToString(coalesce("POSTGRES_USER", "postgres"))
	config.POSTGRES_PASSWORD = cast.ToString(coalesce("POSTGRES_PASSWORD", "1111"))
	config.POSTGRES_NAME = cast.ToString(coalesce("POSTGRES_NAME", "test"))

	config.REDIS_HOST = cast.ToString(coalesce("REDIS_HOST", "redis"))
	config.REDIS_PORT = cast.ToString(coalesce("REDIS_PORT", ":1111"))

	config.NEW_BOT_TOKEN = cast.ToString(coalesce("NEW_BOT_TOKEN", "token"))
	config.BOT_TOKEN = cast.ToString(coalesce("BOT_TOKEN", "token"))
	config.GROUP_ID = cast.ToString(coalesce("GROUP_ID", "-1"))
	config.GROUP_FOR_INFO_ID = cast.ToString(coalesce("GROUP_FOR_INFO_ID", "-1"))

	config.BASE_URL = cast.ToString(coalesce("BASE_URL", "http://localhost:8004"))
	config.CLICK_MERCHANT_ID = cast.ToInt(coalesce("CLICK_MERCHANT_ID", 1))
	config.CLICK_SERVICE_ID = cast.ToInt(coalesce("CLICK_SERVICE_ID", 1))
	config.CLICK_SECRET_KEY = cast.ToString(coalesce("CLICK_SECRET_KEY", "test_click_secret_key"))

	config.PAYMENT_PROVIDER_TOKEN = cast.ToString(coalesce("PAYMENT_PROVIDER_TOKEN", "TEST:631330408:TEST:631330408"))
	config.PAYMENT_CURRENCY = cast.ToString(coalesce("PAYMENT_CURRENCY", "UZS"))
	config.PAYMENT_BOT_TOKEN = cast.ToString(coalesce("PAYMENT_BOT_TOKEN", "token"))

	config.FIREBASE_CREDENTIALS = cast.ToString(coalesce("FIREBASE_CREDENTIALS", "firebase"))

	config.MINIO_ENDPOINT = cast.ToString(coalesce("MINIO_ENDPOINT", "minio"))
	config.MINIO_PORT = cast.ToString(coalesce("MINIO_PORT", ":1000"))
	config.MINIO_ACCESS_KEY = cast.ToString(coalesce("MINIO_ACCESS_KEY", "minio"))
	config.MINIO_SECRET_KEY = cast.ToString(coalesce("MINIO_SECRET_KEY", "minio"))
	config.MINIO_BUCKET_NAME = cast.ToString(coalesce("MINIO_BUCKET_NAME", "test-bucket"))
	config.MINIO_DOMAIN = cast.ToString(coalesce("MINIO_DOMAIN", "minio"))

	config.CONFIRM_PAGE_URL = cast.ToString(coalesce("CONFIRM_PAGE_URL", "test"))

	config.EMAIL_NAME = cast.ToString(coalesce("EMAIL_NAME", "test"))
	config.EMAIL_LOGO = cast.ToString(coalesce("EMAIL_LOGO", "test"))

	config.BREVO_SERVER = cast.ToString(coalesce("BREVO_SERVER", "smtp-relay.brevo.com"))
	config.BREVO_PORT = cast.ToString(coalesce("BREVO_PORT", "587"))
	config.BREVO_LOGIN = cast.ToString(coalesce("BREVO_LOGIN", "test"))
	config.BREVO_KEY = cast.ToString(coalesce("BREVO_KEY", "test"))

	config.S3_ACCESS_KEY = cast.ToString(coalesce("S3_ACCESS_KEY", "test"))
	config.S3_SECRET_KEY = cast.ToString(coalesce("S3_SECRET_KEY", "test"))
	config.S3_ENDPOINT = cast.ToString(coalesce("S3_ENDPOINT", "eu2.contabostorage.com"))
	config.S3_BUCKET_NAME = cast.ToString(coalesce("S3_BUCKET_NAME", "imed"))
	config.S3_DOMAIN = cast.ToString(coalesce("S3_DOMAIN", "https://eu2.contabostorage.com"))
	return config
}

func coalesce(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
