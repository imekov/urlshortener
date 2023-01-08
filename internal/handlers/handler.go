package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/vladimirimekov/url-shortener/internal/server"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Repositories interface {
	ReadData() map[string]map[string]string
	SaveData(map[string]map[string]string) error
}

type Handler struct {
	Storage           Repositories
	LengthOfShortname int
	Host              string
	UserKey           interface{}
	DBConnection      *sql.DB
}

type GetData struct {
	URL string `json:"url"`
}

type AllUserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchData struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url"`
}

func (h Handler) getShortname() string {
	var shortname string
	var result bool

	savedData := h.Storage.ReadData()

	//проверка на существование сгенерированного имени
	for {
		result = true
		shortname = server.GenerateShortname(h.LengthOfShortname)

		for _, value := range savedData {
			if _, ok := value[shortname]; ok {
				result = false
				break
			}
		}

		if result {
			break
		}
	}

	return shortname
}

func (h Handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		data := h.Storage.ReadData()
		shortname := chi.URLParam(r, "id")

		for _, value := range data {
			if originalURL, ok := value[shortname]; ok {
				w.Header().Set("content-type", "text/plain; charset=utf-8")
				w.Header().Set("Location", originalURL)
				w.WriteHeader(http.StatusTemporaryRedirect)
				return
			}
		}

		http.Error(w, "URL not found", http.StatusNotFound)

	case http.MethodPost:

		userID := r.Context().Value(h.UserKey).(string)

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(r.Body)

		currentURL := string(b)

		//проверка на валидность url
		_, err = url.ParseRequestURI(currentURL)
		if err != nil {
			http.Error(w, "Invalid URL value", http.StatusBadRequest)
			return
		}

		shortname := h.getShortname()
		resultData := map[string]map[string]string{userID: {shortname: currentURL}}

		if err = h.Storage.SaveData(resultData); err != nil {
			switch e := err.(type) {
			case *pq.Error:
				if pgerrcode.IsIntegrityConstraintViolation(string(e.Code)) {
					w.Header().Set("content-type", "text/plain; charset=utf-8")
					w.WriteHeader(http.StatusConflict)

					savedData := h.Storage.ReadData()

					for _, value := range savedData {
						for short, original := range value {
							if original == currentURL {
								_, err = w.Write([]byte(h.Host + "/" + short))
								if err != nil {
									http.Error(w, err.Error(), http.StatusInternalServerError)
									return
								}
								break
							}
						}
					}

					return
				}
			default:
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(h.Host + "/" + shortname))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Bad Request", http.StatusBadRequest)

	}
}

func (h Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(h.UserKey).(string)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}(r.Body)

	g := GetData{}

	if err := json.Unmarshal(b, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//проверка на валидность url
	_, err = url.ParseRequestURI(g.URL)
	if err != nil {
		http.Error(w, "Invalid URL value", http.StatusBadRequest)
		return
	}

	shortname := h.getShortname()

	resultData := map[string]map[string]string{userID: {shortname: g.URL}}

	if err = h.Storage.SaveData(resultData); err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if pgerrcode.IsIntegrityConstraintViolation(string(e.Code)) {
				w.Header().Set("content-type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusConflict)

				savedData := h.Storage.ReadData()

				for _, value := range savedData {
					for short, original := range value {
						if original == g.URL {
							_, err = w.Write([]byte(h.Host + "/" + short))
							if err != nil {
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}
							break
						}
					}
				}

				return
			}
		default:
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
	}

	resultJSON, err := json.Marshal(map[string]string{"result": h.Host + "/" + shortname})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h Handler) ShortenBatchHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(h.UserKey).(string)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}(r.Body)

	var g []BatchData

	if err := json.Unmarshal(b, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataToSave := make(map[string]map[string]string)
	dataToSave[userID] = map[string]string{}

	for index, value := range g {
		shortname := h.getShortname()
		dataToSave[userID][shortname] = value.OriginalURL
		g[index].ShortURL = h.Host + "/" + shortname
		g[index].OriginalURL = ""
	}

	h.Storage.SaveData(dataToSave)

	resultJSON, err := json.Marshal(g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) AllShorterURLsHandler(w http.ResponseWriter, r *http.Request) {

	var result []AllUserURLs

	userID := r.Context().Value(h.UserKey).(string)

	savedData := h.Storage.ReadData()
	userData := savedData[userID]
	if len(userData) == 0 {
		err := errors.New("there are no shortened links")
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	for key, value := range userData {
		result = append(result, AllUserURLs{ShortURL: h.Host + "/" + key, OriginalURL: value})
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) PingDBConnection(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.DBConnection.PingContext(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
