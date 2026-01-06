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
	PAYMENT_PROVIDER_TOKEN string
	PAYMENT_CURRENCY       string
	PAYMENT_BOT_TOKEN      string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	config := Config{}

	config.AUTH_PORT = cast.ToString(coalesce("AUTH", ":8004"))

	config.DB_HOST = cast.ToString(coalesce("DB_HOST", "postgres-imed"))
	config.DB_PORT = cast.ToInt(coalesce("DB_PORT", 5432))
	config.DB_USER = cast.ToString(coalesce("DB_USER", "postgres"))
	config.DB_PASSWORD = cast.ToString(coalesce("DB_PASSWORD", "1111"))
	config.DB_NAME = cast.ToString(coalesce("DB_NAME", "imed"))

	config.REDIS_HOST = cast.ToString(coalesce("REDIS_HOST", "redis-imed"))
	config.REDIS_PORT = cast.ToString(coalesce("REDIS_PORT", ":6379"))

	config.BOT_TOKEN = cast.ToString(coalesce("BOT_TOKEN", "7295485024:AAFxtHnfohUF5VNhDdtfUrXQibTXBUIctrA"))
	config.GROUP_ID = cast.ToString(coalesce("GROUP_ID", "-4986197133"))
	config.GROUP_FOR_INFO_ID = cast.ToString(coalesce("GROUP_FOR_INFO_ID", "-4986197133"))

	config.BASE_URL = cast.ToString(coalesce("BASE_URL", "http://localhost:8004"))
	config.PAYMENT_PROVIDER_TOKEN = cast.ToString(coalesce("PAYMENT_PROVIDER_TOKEN", "TEST:631330408:TEST:631330408"))
	config.PAYMENT_CURRENCY = cast.ToString(coalesce("PAYMENT_CURRENCY", "UZS"))
	config.PAYMENT_BOT_TOKEN = cast.ToString(coalesce("PAYMENT_BOT_TOKEN", "8468636479:AAGEqQi-hQSWs7dQSiH-YX9e-ORqyNIw1sc"))

	return config
}

func coalesce(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
