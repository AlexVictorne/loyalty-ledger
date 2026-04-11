package config

import (
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RunAddress  string `env:"RUN_ADDRESS" env-default:":8080"`
	DatabaseURI string `env:"DATABASE_URI"`
	JWTSecret   string `env:"JWT_SECRET" env-default:"dev_secret"`
}

func Load() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("config error: %v", err)
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database uri")
	flag.StringVar(&cfg.JWTSecret, "s", cfg.JWTSecret, "jwt secret")
	flag.Parse()

	return &cfg
}
