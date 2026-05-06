package config

import (
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RunAddress     string `env:"RUN_ADDRESS" env-default:":8080"`
	DatabaseURI    string `env:"DATABASE_URI"`
	JWTSecret      string `env:"JWT_SECRET" env-default:"dev_secret"`
	AccrualAddr    string `env:"ACCRUAL_SYSTEM_ADDRESS" env-default:"http://localhost:8081"`
	AccrualPoller  int    `env:"ACCRUAL_WORKER_INTERVAL" env-default:"5"` // в секундах
	AccrualWorkers int    `env:"ACCRUAL_WORKER_COUNT" env-default:"4"`    // кол-во параллельных воркеров для обработки заказов
}

func Load() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("config error: %v", err)
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database uris")
	flag.StringVar(&cfg.JWTSecret, "s", cfg.JWTSecret, "jwt secret")
	flag.StringVar(&cfg.AccrualAddr, "r", cfg.AccrualAddr, "accrual system address")
	flag.IntVar(&cfg.AccrualPoller, "accrual-interval", cfg.AccrualPoller, "accrual poll interval (seconds)")
	flag.IntVar(&cfg.AccrualWorkers, "accrual-workers", cfg.AccrualWorkers, "number of parallel accrual workers")
	flag.Parse()

	return &cfg
}
