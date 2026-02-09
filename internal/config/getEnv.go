package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	AUTH_PORT string

	POSTGRES_HOST     string
	POSTGRES_PORT     int
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_NAME     string

	REDIS_HOST string
	REDIS_PORT string

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

	MINIO_ENDPOINT    string
	MINIO_PORT        string
	MINIO_ACCESS_KEY  string
	MINIO_SECRET_KEY  string
	MINIO_BUCKET_NAME string
	MINIO_DOMAIN      string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	config := Config{}

	config.AUTH_PORT = cast.ToString(coalesce("AUTH_PORT", ":1001"))

	config.POSTGRES_HOST = cast.ToString(coalesce("POSTGRES_HOST", "postgres"))
	config.POSTGRES_PORT = cast.ToInt(coalesce("POSTGRES_PORT", 5432))
	config.POSTGRES_USER = cast.ToString(coalesce("POSTGRES_USER", "postgres"))
	config.POSTGRES_PASSWORD = cast.ToString(coalesce("POSTGRES_PASSWORD", "1111"))
	config.POSTGRES_NAME = cast.ToString(coalesce("POSTGRES_NAME", "test"))

	config.REDIS_HOST = cast.ToString(coalesce("REDIS_HOST", "redis"))
	config.REDIS_PORT = cast.ToString(coalesce("REDIS_PORT", ":1111"))

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
	config.MINIO_DOMAIN = cast.ToString(coalesce(("MINIO_DOMAIN"), "minio"))

	return config
}

func coalesce(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
