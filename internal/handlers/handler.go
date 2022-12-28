package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vladimirimekov/url-shortener/internal/server"
	"io"
	"net/http"
	"net/url"
)

type Repositories interface {
	ReadData() map[string]map[string]string
	SaveData(map[string]map[string]string)
	Encrypt(string, string) (string, error)
	Decrypt(string, string) (string, error)
}

type Handler struct {
	Storage           Repositories
	LengthOfShortname int
	Host              string
	Salt              string
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

	savedData := h.Storage.ReadData()

	//проверка на существование сгенерированного имени
	for {
		shortname = server.GenerateShortname(h.LengthOfShortname)
		if _, ok := savedData[userID][shortname]; !ok {
			break
		}
	}

	savedData[userID][shortname] = url
	h.Storage.SaveData(savedData)

	return shortname
}

func (h Handler) getUserID(r *http.Request) (result string, err error) {
	st, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	userID, err := h.Storage.Decrypt(st.Value, h.Salt)
	if err != nil {
		return "", err
	}

	return userID, nil

}

func (h Handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		userID, err := h.getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := h.Storage.ReadData()
		shortnameID := chi.URLParam(r, "id")

		if originalURL, ok := data[userID][shortnameID]; ok {
			w.Header().Set("content-type", "text/plain; charset=utf-8")
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "URL not found", http.StatusNotFound)
		}

	case http.MethodPost:

		userID, err := h.getUserID(r)
		if err != nil {
			err401 := errors.New("unauthorized user")
			http.Error(w, err401.Error(), http.StatusUnauthorized)
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

	st, _ := r.Cookie("session_token")
	userID, err := h.Storage.Decrypt(st.Value, h.Salt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	savedData := h.Storage.ReadData()
	userData := savedData[userID]
	if len(userData) == 0 {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	for key, value := range userData {
		result = append(result, AllUserURLs{ShortURL: key, OriginalURL: value})
	}

	resultJSON, err := json.Marshal(result)
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

func (h Handler) CheckUserCookies(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		st, err := r.Cookie("session_token")
		if err == nil {
			userID, errDecrypt := h.Storage.Decrypt(st.Value, h.Salt)

			savedData := h.Storage.ReadData()
			_, ok := savedData[userID]

			if errDecrypt == nil && ok {
				next.ServeHTTP(w, r)
				return
			}

		}

		sessionToken := uuid.NewString()
		savedData := h.Storage.ReadData()
		savedData[sessionToken] = map[string]string{}
		h.Storage.SaveData(savedData)

		enc, err := h.Storage.Encrypt(sessionToken, h.Salt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: enc,
			Path:  "/",
		})

		next.ServeHTTP(w, r)
	})

}
