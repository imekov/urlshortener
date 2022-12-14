package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vladimirimekov/url-shortener/internal/handlers"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"log"
	"net/http"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	Filename        string `env:"FILE_STORAGE_PATH" envDefault:"data.gob"`
	ShortnameLength int
}

func main() {

	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server start address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "the base address of the resulting shortened URL")
	flag.StringVar(&cfg.Filename, "f", cfg.Filename, "path to file with shortened URLs")
	flag.Parse()

	cfg.ShortnameLength = 8

	s := storage.Storage{Filename: cfg.Filename}
	h := handlers.Handler{Storage: s, LengthOfShortname: cfg.ShortnameLength, Host: cfg.BaseURL}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.MainHandler)
	})

	r.Post("/", h.MainHandler)

	r.Route("/api/shorten", func(r chi.Router) {
		r.Post("/", h.ShortenHandler)
	})

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
