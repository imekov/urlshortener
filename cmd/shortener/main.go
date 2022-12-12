package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vladimirimekov/url-shortener/internal/handlers"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"log"
	"net/http"
)

type Config struct {
	SERVER_ADDRESS   string
	BASE_URL         string
	FILENAME         string
	SHORTNAME_LENGTH int
}

func main() {

	conf := Config{SERVER_ADDRESS: ":8080", BASE_URL: "http://localhost:8080/", FILENAME: "data.gob", SHORTNAME_LENGTH: 8}

	s := storage.Storage{Filename: conf.FILENAME}
	h := handlers.Handler{Storage: s, LengthOfShortname: conf.SHORTNAME_LENGTH, Host: conf.BASE_URL}

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

	log.Fatal(http.ListenAndServe(conf.SERVER_ADDRESS, r))
}
