package urlshortener

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	Filename        string `env:"FILE_STORAGE_PATH" envDefault:"data.gob"`
	Secret          string `env:"SECRET_KEY" envDefault:"y3T8h2wYJGlgzLmWjjflfcUW0NYBeEJ6"`
	DBAddress       string `env:"DATABASE_DSN" envDefault:"db.db"`
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
	flag.StringVar(&cfg.Filename, "f", cfg.Filename, "the path to file with shortened URLs")
	flag.StringVar(&cfg.Secret, "s", cfg.Secret, "secret key")
	flag.StringVar(&cfg.DBAddress, "d", cfg.DBAddress, "the address of the connection to the database")
	flag.Parse()

	return cfg
}
