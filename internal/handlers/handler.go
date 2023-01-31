package handlers

import "C"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type Repositories interface {
	ReadData() map[string]map[string]string
	SaveData(map[string]map[string]string) error
	DeleteData([]string, string)
	GetURLByShortname(string) (string, bool)
	PingDBConnection(ctx context.Context) error
}

type Handler struct {
	Storage           Repositories
	LengthOfShortname int
	Host              string
	UserKey           interface{}
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

		letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

		s := make([]rune, h.LengthOfShortname)
		for i := range s {
			s[i] = letters[rand.Intn(len(letters))]
		}

		shortname = string(s)

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

func (h Handler) getUserID(r *http.Request) (string, error) {
	var userID string
	uk := r.Context().Value(h.UserKey)

	switch v := uk.(type) {
	default:
		err := fmt.Errorf("unexpected type of user key %T", v)
		return "", err
	case []byte:
		userID = string(uk.([]byte))
	case string:
		userID = uk.(string)
	}

	return userID, nil
}

func (h Handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
		defer cancel()
		r = r.WithContext(ctx)

		shortname := chi.URLParam(r, "id")
		w.Header().Set("content-type", "text/plain; charset=utf-8")

		if originalURL, isDelete := h.Storage.GetURLByShortname(shortname); isDelete {
			w.WriteHeader(http.StatusGone)
		} else if originalURL == "" {
			http.Error(w, "URL not found", http.StatusNotFound)
		} else {
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}

	case http.MethodPost:

		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
		defer cancel()
		r = r.WithContext(ctx)

		userID, err := h.getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
					w.Header().Set("content-type", "application/json")
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

func (h Handler) PostShortenHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
	defer cancel()
	r = r.WithContext(ctx)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusConflict)

				savedData := h.Storage.ReadData()

				for _, value := range savedData {
					for short, original := range value {
						if original == g.URL {

							resultJSON, err := json.Marshal(map[string]string{"result": h.Host + "/" + short})

							if err != nil {
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}

							_, err = w.Write(resultJSON)
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

func (h Handler) PostShortenBatchHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
	defer cancel()
	r = r.WithContext(ctx)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

func (h Handler) GetAllShorterURLsHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
	defer cancel()
	r = r.WithContext(ctx)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	savedData := h.Storage.ReadData()
	userData := savedData[userID]

	if len(userData) == 0 {
		err := errors.New("there are no shortened links")
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	result := make([]AllUserURLs, 0, len(userData))

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

func (h Handler) DeleteURLS(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
	defer cancel()
	r = r.WithContext(ctx)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	var g []string

	if err = json.Unmarshal(b, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go h.Storage.DeleteData(g, userID)

}

func (h Handler) PingDBConnection(w http.ResponseWriter, r *http.Request) {
	if err := h.Storage.PingDBConnection(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
