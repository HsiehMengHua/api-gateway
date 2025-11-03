package main

import (
	"api-gateway/router"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	loadEnv()
}

func main() {
	r := router.Setup()
	r.Run(os.Getenv("APP_PORT"))
}

func loadEnv() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // The Original .env
}
