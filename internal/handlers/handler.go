package handlers

import (
	"github.com/vladimirimekov/url-shortener/internal/server"
	"io"
	"net/http"
	"strings"
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

func (h Handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		data := h.Storage.ReadData()
		shortnameID := strings.ReplaceAll(r.URL.Path, "/", "")

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
		defer r.Body.Close()

		//TODO сделать проверку полученного запроса на корректность URL регуляркой, в случае ошибки 400

		var shortname string
		savedData := h.Storage.ReadData()

		//проверка на существование сгенерированного имени
		for {
			shortname = server.GenerateShortname(h.LengthOfShortname)
			if _, ok := savedData[shortname]; !ok {
				break
			}
		}

		savedData[shortname] = string(b)
		h.Storage.SaveData(savedData)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(h.Host + shortname))

	default:
		http.Error(w, "Bad Request", 400)
	}
}
