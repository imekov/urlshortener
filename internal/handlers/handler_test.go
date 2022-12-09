package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandler_MainHandler(t *testing.T) {
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
			name:     "sipmle url",
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

	s := storage.Storage{Filename: "data.gob"}
	d := Handler{s, 8, "http://localhost:8080/"}

	for _, tt := range tests {
		var shortURL string
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := chi.NewRouter()
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
		})

		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			h := chi.NewRouter()
			h.HandleFunc("/{id}", d.MainHandler)

			requestGet := httptest.NewRequest(http.MethodGet, shortURL, nil)
			h.ServeHTTP(w, requestGet)
			resultGet := w.Result()

			err := resultGet.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.wantGet.statusCode, resultGet.StatusCode)
			assert.Equal(t, tt.wantGet.contentType, resultGet.Header.Get("Content-Type"))
			assert.Equal(t, tt.urlValue, resultGet.Header.Get("Location"))
		})

	}

	err := os.Remove("data.gob")
	if err != nil {
		log.Fatal(err)
	}

}

func TestHandler_ShortenHandler(t *testing.T) {

	type sourceData struct {
		Url string `json:"url"`
	}

	type resultData struct {
		Result string `json:"result"`
	}
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name     string
		url      sourceData
		wantPost want
	}{
		{
			name:     "sipmle url",
			url:      sourceData{Url: "https://google.com"},
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
		},
		{
			name:     "long url",
			url:      sourceData{Url: "https://goiejrgoijergiojposd.com"},
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
		},
		{
			name:     "long url with slugs",
			url:      sourceData{Url: "https://rthiiurgfougjfeorferguti.com/thgeufijrgeuhfjwer/gerhuiojgeuh"},
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
		},
	}

	s := storage.Storage{Filename: "data.gob"}
	d := Handler{s, 8, "http://localhost:8080/"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := chi.NewRouter()
			h.HandleFunc("/api/shorten", d.ShortenHandler)

			sendJson, err := json.Marshal(tt.url)
			require.NoError(t, err)

			requestPost := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(sendJson))
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
			require.NoError(t, err)

		})

	}

	err := os.Remove("data.gob")
	if err != nil {
		log.Fatal(err)
	}

}
