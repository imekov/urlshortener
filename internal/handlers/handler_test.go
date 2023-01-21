package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	urlshortener "github.com/vladimirimekov/url-shortener"
	"github.com/vladimirimekov/url-shortener/internal/middlewares"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const userKey string = "userid"
const secretKey string = "0Fg79lY0Tq3cdUTMHIcNBvDF0m6QfEZF"

var cfg urlshortener.Config

func TestHandler_MainHandler(t *testing.T) {

	cfg = urlshortener.GetConfig()

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name     string
		urlValue string
		wantPost want
		wantGet  want
	}{
		{
			name:     "simple url",
			urlValue: "https://google.com",
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
			wantGet:  want{contentType: "text/plain; charset=utf-8", statusCode: http.StatusTemporaryRedirect},
		},
		{
			name:     "long url",
			urlValue: "https://goiejrgoijergiojposd.com",
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
			wantGet:  want{contentType: "text/plain; charset=utf-8", statusCode: http.StatusTemporaryRedirect},
		},
		{
			name:     "long url with slugs",
			urlValue: "https://rthiiurgfougjfeorferguti.com/thgeufijrgeuhfjwer/gerhuiojgeuh",
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
			wantGet:  want{contentType: "text/plain; charset=utf-8", statusCode: http.StatusTemporaryRedirect},
		},
	}

	dbConnection, err := sql.Open("postgres", cfg.DBAddress)
	if err != nil {
		panic(err)
	}
	defer dbConnection.Close()

	if cfg.Filename == "" {
		cfg.Filename = "data.gob"
	}

	s := storage.FileSystemConnect{Filename: cfg.Filename}
	d := Handler{
		Storage:           s,
		LengthOfShortname: cfg.ShortnameLength,
		Host:              cfg.BaseURL,
		UserKey:           userKey,
		DBConnection:      dbConnection}

	m := middlewares.UserCookies{Storage: s, Secret: secretKey, UserKey: userKey}

	for _, tt := range tests {
		var shortURL string
		var userID string
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := chi.NewRouter()
			h.Use(m.CheckUserCookies)
			h.HandleFunc("/", d.MainHandler)

			requestPost := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.urlValue))
			h.ServeHTTP(w, requestPost)
			resultPost := w.Result()

			assert.Equal(t, tt.wantPost.statusCode, resultPost.StatusCode)
			assert.Equal(t, tt.wantPost.contentType, resultPost.Header.Get("Content-Type"))

			shortname, err := io.ReadAll(resultPost.Body)
			require.NoError(t, err)
			err = resultPost.Body.Close()
			require.NoError(t, err)
			shortURL = string(shortname)

			for _, cookie := range resultPost.Cookies() {
				if cookie.Name == "session_token" {
					userID = cookie.Value
				}
			}

		})

		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			h := chi.NewRouter()
			h.Use(m.CheckUserCookies)

			h.HandleFunc("/{id}", d.MainHandler)

			requestGet := httptest.NewRequest(http.MethodGet, shortURL, nil)

			requestGet.AddCookie(&http.Cookie{
				Name:  "session_token",
				Value: userID,
				Path:  "/",
			})

			h.ServeHTTP(w, requestGet)
			resultGet := w.Result()

			err := resultGet.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.wantGet.statusCode, resultGet.StatusCode)
			assert.Equal(t, tt.wantGet.contentType, resultGet.Header.Get("Content-Type"))
			assert.Equal(t, tt.urlValue, resultGet.Header.Get("Location"))
		})

	}

	err = os.Remove("data.gob")
	if err != nil {
		log.Print(err)
	}

}

func TestHandler_ShortenHandler(t *testing.T) {

	type sourceData struct {
		URL string `json:"url"`
	}

	type resultData struct {
		Result string `json:"result"`
	}
	type want struct {
		contentType string
		statusCode  int
		err         error
	}
	tests := []struct {
		name     string
		url      sourceData
		wantPost want
	}{
		{
			name:     "simple url",
			url:      sourceData{URL: "https://google.com"},
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated, err: nil},
		},
		{
			name:     "long url",
			url:      sourceData{URL: "https://goiejrgoijergiojposd.com"},
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated, err: nil},
		},
		{
			name:     "long url with slugs",
			url:      sourceData{URL: "https://rthiiurgfougjfeorferguti.com/thgeufijrgeuhfjwer/gerhuiojgeuh"},
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated, err: nil},
		},
		{
			name:     "int in url",
			url:      sourceData{URL: "6547898765"},
			wantPost: want{contentType: "text/plain; charset=utf-8", statusCode: http.StatusBadRequest, err: &json.SyntaxError{}},
		},
	}

	dbConnection, err := sql.Open("postgres", cfg.DBAddress)
	if err != nil {
		panic(err)
	}
	defer dbConnection.Close()

	s := storage.FileSystemConnect{Filename: cfg.Filename}
	d := Handler{
		Storage:           s,
		LengthOfShortname: cfg.ShortnameLength,
		Host:              cfg.BaseURL,
		UserKey:           userKey,
		DBConnection:      dbConnection}

	m := middlewares.UserCookies{Storage: s, Secret: secretKey, UserKey: userKey}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := chi.NewRouter()
			h.Use(m.CheckUserCookies)
			h.HandleFunc("/api/shorten", d.PostShortenHandler)

			sendJSON, err := json.Marshal(tt.url)
			require.NoError(t, err)

			requestPost := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(sendJSON))
			h.ServeHTTP(w, requestPost)
			resultPost := w.Result()

			assert.Equal(t, tt.wantPost.statusCode, resultPost.StatusCode)
			assert.Equal(t, tt.wantPost.contentType, resultPost.Header.Get("Content-Type"))

			b, err := io.ReadAll(resultPost.Body)
			require.NoError(t, err)
			err = resultPost.Body.Close()
			require.NoError(t, err)

			g := resultData{}

			err = json.Unmarshal(b, &g)
			assert.IsType(t, tt.wantPost.err, err)

		})

	}

	err = os.Remove("data.gob")
	if err != nil {
		log.Print(err)
	}

}
