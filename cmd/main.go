package main

import (
	"imed/app"
	"imed/internal/config"
)

func main() {
	cfg := config.Load()

	app.Run(&cfg)
}
