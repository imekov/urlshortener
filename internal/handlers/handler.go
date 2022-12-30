package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/vladimirimekov/url-shortener/internal/server"
	"io"
	"net/http"
	"net/url"
)

type Repositories interface {
	ReadData() map[string]map[string]string
	SaveData(map[string]map[string]string)
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

type SendData struct {
	Result string `json:"result"`
}

type AllUserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h Handler) getShortname(url string, userID string) string {
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

	savedData[userID][shortname] = url
	h.Storage.SaveData(savedData)

	return shortname
}

func (h Handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		var originalURL string

		data := h.Storage.ReadData()
		shortnameID := chi.URLParam(r, "id")

		for _, value := range data {
			if originalURL, ok := value[shortnameID]; ok {
				w.Header().Set("content-type", "text/plain; charset=utf-8")
				w.Header().Set("Location", originalURL)
				w.WriteHeader(http.StatusTemporaryRedirect)
				break
			}
		}

		if originalURL == "" {
			http.Error(w, "URL not found", http.StatusNotFound)
		}

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

		shortname := h.getShortname(currentURL, userID)

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

	shortname := h.getShortname(g.URL, userID)
	resultData := SendData{
		Result: h.Host + "/" + shortname,
	}

	resultJSON, err := json.Marshal(resultData)
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
