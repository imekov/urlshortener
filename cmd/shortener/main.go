package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vladimirimekov/url-shortener/internal/handlers"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"log"
	"net/http"
)

const (
	filename          = "data.gob"
	lengthOfShortname = 8
	hostname          = "http://localhost:8080/"
)

func main() {

	s := storage.Storage{Filename: filename}
	h := handlers.Handler{Storage: s, LengthOfShortname: lengthOfShortname, Host: hostname}

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

	log.Fatal(http.ListenAndServe(":8080", r))
}
