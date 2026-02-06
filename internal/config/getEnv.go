package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	AUTH_PORT string

	DB_HOST     string
	DB_PORT     int
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string

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
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	config := Config{}

	config.AUTH_PORT = cast.ToString(coalesce("AUTH", ":1001"))

	config.DB_HOST = cast.ToString(coalesce("DB_HOST", "postgres"))
	config.DB_PORT = cast.ToInt(coalesce("DB_PORT", 1111))
	config.DB_USER = cast.ToString(coalesce("DB_USER", "postgres"))
	config.DB_PASSWORD = cast.ToString(coalesce("DB_PASSWORD", "1111"))
	config.DB_NAME = cast.ToString(coalesce("DB_NAME", "test"))

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

	return config
}

func coalesce(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
