package main

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/vladimirimekov/url-shortener"
	"github.com/vladimirimekov/url-shortener/internal/handlers"
	"github.com/vladimirimekov/url-shortener/internal/middlewares"
	"github.com/vladimirimekov/url-shortener/internal/server"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"log"
	"net/http"
)

type userIDtype string

const userKey userIDtype = "userid"

func main() {

	cfg := urlshortener.GetConfig()
	dbConnection := server.Connect(cfg.DBAddress)
	defer dbConnection.Close()

	s := storage.Storage{Filename: cfg.Filename}

	h := handlers.Handler{
		Storage:           s,
		LengthOfShortname: cfg.ShortnameLength,
		Host:              cfg.BaseURL,
		UserKey:           userKey,
		DBConnection:      dbConnection}

	m := middlewares.UserCookies{Storage: s, Secret: cfg.Secret, UserKey: userKey}

	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Use(middlewares.GZIPRead)
	r.Use(middlewares.GZIPWrite)
	r.Use(m.CheckUserCookies)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.MainHandler)
	})

	r.Post("/", h.MainHandler)

	r.Route("/api/shorten", func(r chi.Router) {
		r.Post("/", h.ShortenHandler)
	})

	r.Route("/api/user/urls", func(r chi.Router) {
		r.Get("/", h.AllShorterURLsHandler)
	})

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", h.PingDBConnection)
	})

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
