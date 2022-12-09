package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/vladimirimekov/url-shortener/internal/server"
	"io"
	"net/http"
	"net/url"
)

type Repositories interface {
	ReadData() map[string]string
	SaveData(map[string]string)
}

type Handler struct {
	Storage           Repositories
	LengthOfShortname int
	Host              string
}

type GetData struct {
	URL string `json:"url"`
}

type SendData struct {
	Result string `json:"result"`
}

func (h Handler) getShortname(url string) string {
	var shortname string
	savedData := h.Storage.ReadData()

	//проверка на существование сгенерированного имени
	for {
		shortname = server.GenerateShortname(h.LengthOfShortname)
		if _, ok := savedData[shortname]; !ok {
			break
		}
	}

	savedData[shortname] = url
	h.Storage.SaveData(savedData)

	return shortname
}

func (h Handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		data := h.Storage.ReadData()
		shortnameID := chi.URLParam(r, "id")

		if url, ok := data[shortnameID]; ok {
			w.Header().Set("content-type", "text/plain; charset=utf-8")
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "URL not found", 404)
		}

	case http.MethodPost:

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}(r.Body)

		currentURL := string(b)

		//проверка на валидность url
		_, err = url.ParseRequestURI(currentURL)
		if err != nil {
			http.Error(w, "Invalid URL value", 400)
			return
		}

		shortname := h.getShortname(currentURL)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(h.Host + shortname))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	default:
		http.Error(w, "Bad Request", 400)
	}
}

func (h Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}(r.Body)

	g := GetData{}

	if err := json.Unmarshal(b, &g); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//проверка на валидность url
	_, err = url.ParseRequestURI(g.URL)
	if err != nil {
		http.Error(w, "Invalid URL value", 400)
		return
	}

	shortname := h.getShortname(g.URL)
	resultData := SendData{
		Result: h.Host + shortname,
	}

	resultJSON, err := json.Marshal(resultData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

}
