package internal

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config содержит ключевые параметры для работы программы.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	Filename        string `env:"FILE_STORAGE_PATH"`
	DBAddress       string `env:"DATABASE_DSN"`
	JSONConfig      string `env:"CONFIG"`
	ShortnameLength int    `env:"SHORTNAME_LENGTH" envDefault:"8"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
	Secret          []byte
}

// FileConfig содержит параметры для чтения из JSON.
type FileConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

// GetConfig читает данные из окружения и возвращает заполненный Config.
func GetConfig() Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.JSONConfig, "c", cfg.JSONConfig, "JSON configuration file name")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server start address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "the base address of the resulting shortened URL")
	flag.StringVar(&cfg.Filename, "f", cfg.Filename, "the path to file with shortened URLs")
	flag.StringVar(&cfg.DBAddress, "d", cfg.DBAddress, "the address of the connection to the database")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "classless address string representation (CIDR)")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "start server with HTTPS")
	flag.Parse()

	cfg.Secret = make([]byte, 16)
	_, err = rand.Read(cfg.Secret)
	if err != nil {
		log.Fatal(err)
	}

	if len(cfg.JSONConfig) != 0 {
		jsonData, err := os.ReadFile(cfg.JSONConfig)
		if err != nil {
			log.Fatal(err)
		}

		var fileconfig FileConfig
		err = json.Unmarshal(jsonData, &fileconfig)
		if err != nil {
			log.Fatal(err)
		}

		if len(cfg.ServerAddress) == 0 {
			cfg.ServerAddress = fileconfig.ServerAddress
		}
		if len(cfg.BaseURL) == 0 {
			cfg.BaseURL = fileconfig.BaseURL
		}
		if len(cfg.Filename) == 0 {
			cfg.Filename = fileconfig.FileStoragePath
		}
		if len(cfg.DBAddress) == 0 {
			cfg.DBAddress = fileconfig.DatabaseDsn
		}
		if !cfg.EnableHTTPS {
			cfg.EnableHTTPS = fileconfig.EnableHTTPS
		}
		if len(cfg.TrustedSubnet) == 0 {
			cfg.TrustedSubnet = fileconfig.TrustedSubnet
		}

	}

	return cfg
}
