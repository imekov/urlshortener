package url_shortener

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	Filename        string `env:"FILE_STORAGE_PATH" envDefault:"data.gob"`
	Salt            string `env:"CIPHER_KEY" envDefault:"y3T8h2wYJGlgzLmWjjflfcUW0NYBeEJ6"`
	ShortnameLength int    `env:"SHORTNAME_LENGTH" envDefault:"8"`
}

func GetConfig() Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server start address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "the base address of the resulting shortened URL")
	flag.StringVar(&cfg.Filename, "f", cfg.Filename, "path to file with shortened URLs")
	flag.StringVar(&cfg.Salt, "s", cfg.Salt, "cipher key")
	flag.Parse()

	return cfg
}