package server

import (
	"database/sql"
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vladimirimekov/url-shortener/internal"
	"github.com/vladimirimekov/url-shortener/internal/handlers"
	"github.com/vladimirimekov/url-shortener/internal/middlewares"
	"github.com/vladimirimekov/url-shortener/internal/storage"
)

type userIDtype string

const userKey userIDtype = "userid"

// GetServer возвращает Chi сервер со всеми хэндлерами и мидлвэрами.
func GetServer(dbConnection *sql.DB) (string, *chi.Mux) {

	cfg := internal.GetConfig()
	memoryVar := make(map[string]map[string]string)

	h := handlers.Handler{
		LengthOfShortname: cfg.ShortnameLength,
		Host:              cfg.BaseURL,
		UserKey:           userKey,
	}

	if cfg.DBAddress != "" {
		var err error

		for i := 1; i <= 5; i++ {
			dbConnection, err = sql.Open("postgres", cfg.DBAddress)
			if err == nil {
				break
			}
			time.Sleep(30 * time.Second)
		}

		if err != nil {
			log.Fatalf("unable to connect to database %v\n", cfg.DBAddress)
		}

		h.Storage = storage.GetNewConnection(dbConnection, cfg.DBAddress, "file://migrations/postgres")
	} else if cfg.Filename != "" {
		h.Storage = storage.FileSystemConnect{Filename: cfg.Filename}
	} else {
		h.Storage = storage.MemoryWork{UserData: memoryVar}
	}

	m := middlewares.UserCookies{Storage: h.Storage, Secret: cfg.Secret, UserKey: userKey}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GZIPRead)
	r.Use(middlewares.GZIPWrite)

	r.Use(m.CheckUserCookies)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.MainHandler)
	})

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.MainHandler)
	})

	r.Route("/api", func(r chi.Router) {

		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", h.PostShortenHandler)

			r.Route("/batch", func(r chi.Router) {
				r.Post("/", h.PostShortenBatchHandler)
			})
		})

		r.Route("/user/urls", func(r chi.Router) {
			r.Get("/", h.GetAllShorterURLsHandler)
			r.Delete("/", h.DeleteURLS)
		})

	})

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", h.PingDBConnection)
	})

	return cfg.ServerAddress, r
}
