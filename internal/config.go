package internal

import (
	"crypto/rand"
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

// Config содержит ключевые параметры для работы прогрмамы.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	Filename        string `env:"FILE_STORAGE_PATH"`
	DBAddress       string `env:"DATABASE_DSN"`
	ShortnameLength int    `env:"SHORTNAME_LENGTH" envDefault:"8"`
	Secret          []byte
}

// GetConfig читает данные из окружения и возвращает заполненный Config.
func GetConfig() Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server start address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "the base address of the resulting shortened URL")
	flag.StringVar(&cfg.Filename, "f", cfg.Filename, "the path to file with shortened URLs")
	flag.StringVar(&cfg.DBAddress, "d", cfg.DBAddress, "the address of the connection to the database")
	flag.Parse()

	cfg.Secret = make([]byte, 16)
	_, err = rand.Read(cfg.Secret)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
